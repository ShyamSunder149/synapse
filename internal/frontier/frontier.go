package frontier

import (
	"context"
	"log"
	"net/url"
	"sync"
	"time"

	"github.com/ritvikos/synapse/internal/frontier/robots"
	"github.com/ritvikos/synapse/pkg/frontier/score"
	model "github.com/ritvikos/synapse/pkg/model"
)

type Config struct {
	IngressBufSize        int
	RobotsResolvedBufSize int
	ScoreBufSize          int
	DefaultCrawlDelay     time.Duration

	ScoreWorkerCount     uint
	RobotsWorkerCount    uint
	SchedulerWorkerCount uint
}

// T represents crawl metadata (e.g., URL, Request).
type Frontier[T any] struct {
	robotstxt *robots.RobotsTxtHandler
	scorer    score.ScorePolicy[T]
	scheduler *Scheduler[T]
	config    Config

	// Channels
	ingressCh        chan *model.Task[T]
	robotsResolvedCh chan *model.Task[T]
	scoredCh         chan *model.Task[T]

	// Internal
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewFrontier[T any](
	robotsHandler *robots.RobotsTxtHandler,
	scorer score.ScorePolicy[T],
	scheduler *Scheduler[T],
	config Config,
) *Frontier[T] {
	return &Frontier[T]{
		robotstxt: robotsHandler,
		scorer:    scorer,
		scheduler: scheduler,
		config:    config,
	}
}

func (f *Frontier[T]) Submit(ctx context.Context, endpoint string, metadata *T) error {
	url, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	task := model.Task[T]{
		Url:      url,
		Metadata: metadata,
	}

	select {
	case f.ingressCh <- &task:
	case <-ctx.Done():
	}

	return nil
}

func (f *Frontier[T]) Start(ctx context.Context) error {
	f.ctx, f.cancel = context.WithCancel(ctx)

	f.ingressCh = make(chan *model.Task[T], f.config.IngressBufSize)
	f.robotsResolvedCh = make(chan *model.Task[T], f.config.RobotsResolvedBufSize)
	f.scoredCh = make(chan *model.Task[T], f.config.ScoreBufSize)

	if err := f.robotstxt.Start(f.ctx); err != nil {
		return err
	}
	if err := f.scheduler.Start(f.ctx); err != nil {
		return err
	}

	for range f.config.RobotsWorkerCount {
		f.wg.Add(1)
		go f.robotsWorker()
	}

	for range f.config.ScoreWorkerCount {
		f.wg.Add(1)
		go f.scoreWorker()
	}

	for range f.config.SchedulerWorkerCount {
		f.wg.Add(1)
		go f.scheduleWorker()
	}

	return f.scheduler.Start(ctx)
}

func (f *Frontier[T]) robotsWorker() {
	defer f.wg.Done()

	for {
		select {
		case <-f.ctx.Done():
			return

		case task, ok := <-f.ingressCh:
			if !ok {
				log.Println("scored channel closed, stopping robots worker")
				return
			}

			entry, err := f.robotstxt.Retrieve(task.Url.Host)
			if err == robots.RobotsTxtNeedsFetchError {
				if err := f.robotstxt.Submit(f.ctx, task.Url.Host); err != nil {
					log.Printf("error requesting robots.txt fetch for host %s: %v", task.Url.Host, err)
					continue
				}

				// TODO: Decide on a better strategy than immediate re-enqueue
				// Probably, maintain a pending queue and process after robots.txt is fetched
				// For now, re-enqueue the task for later processing
				select {
				case f.scoredCh <- task:
				case <-f.ctx.Done():
					return
				}
				continue
			}
			if !entry.IsAllowed {
				log.Printf("disallowed by robots.txt: host=%s url=%s", task.Url.Host, task.Url.String())
				continue
			}

			now := time.Now()
			if entry.CrawlDelay != 0 {
				task.ExecuteAt = now.Add(entry.CrawlDelay)
			} else {
				task.ExecuteAt = now.Add(f.config.DefaultCrawlDelay)
			}

			select {
			case f.robotsResolvedCh <- task:
			case <-f.ctx.Done():
				return
			}
		}
	}
}

func (f *Frontier[T]) scoreWorker() {
	defer f.wg.Done()

	for {
		select {
		case <-f.ctx.Done():
			return

		case task, ok := <-f.robotsResolvedCh:
			if !ok {
				log.Println("ingress channel closed, stopping score worker")
				return
			}

			score, err := f.scorer.Score(f.ctx, task)
			if err != nil {
				log.Printf("error scoring item: %v", err)
				continue
			}
			task.Score = score

			select {
			case f.scoredCh <- task:
			case <-f.ctx.Done():
				return
			}
		}
	}
}

func (f *Frontier[T]) scheduleWorker() {
	defer f.wg.Done()

	for {
		select {
		case <-f.ctx.Done():
			return

		case task, ok := <-f.scoredCh:
			if !ok {
				log.Println("robots resolved channel closed, stopping scheduler worker")
				return
			}

			err := f.scheduler.Write(task)
			if err != nil {
				log.Printf("error scheduling task for url %s: %v", task.Url.String(), err)
				continue
			}
		}
	}
}
