package robots

import (
	"context"
	"fmt"
	"sync"
)

// TODO: Pending implementation

var ErrRobotsTxtNeedsFetch = fmt.Errorf("robots.txt needs fetch")

type RobotsHandler struct {
	userAgent string
	fetcher   RobotsTxtFetcher
	backend   RobotsTxtBackend
	requestCh chan string
	workers   uint

	// Internal
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex
}

func NewRobotsHandler(
	userAgent string,
	fetcher RobotsTxtFetcher,
	backend RobotsTxtBackend,
	requestChSize uint,
	workers uint,
) *RobotsHandler {
	return &RobotsHandler{
		userAgent: userAgent,
		fetcher:   fetcher,
		backend:   backend,
		requestCh: make(chan string, requestChSize),
		workers:   workers,
	}
}

func (r *RobotsHandler) Start(ctx context.Context) error {
	r.ctx, r.cancel = context.WithCancel(ctx)

	for range r.workers {
		r.wg.Add(1)
		go r.fetchWorker()
	}

	return nil
}

func (r *RobotsHandler) Submit(ctx context.Context, host string) error {
	select {
	case r.requestCh <- host:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *RobotsHandler) fetchWorker() {
	defer r.wg.Done()

	for {
		select {
		case <-r.ctx.Done():
			return

		case host, ok := <-r.requestCh:
			if !ok {
				fmt.Println("robots fetch worker: channel closed, exiting")
				return
			}

			entry, err := r.fetcher.Fetch(r.ctx, host)
			if err != nil {
				fmt.Printf("robots fetch worker: error fetching robots.txt for host %s: %v\n", host, err)
				continue
			}

			if err := r.backend.Set(r.ctx, entry); err != nil {
				fmt.Printf("robots fetch worker: error storing robots.txt for host %s: %v\n", host, err)
				continue
			}
		}
	}
}

func (r *RobotsHandler) Retrieve(host string) (RobotsEntry, error) {
	if has, err := r.backend.Has(r.ctx, host); err != nil {
		return RobotsEntry{}, err
	} else if has {
		return r.backend.Get(r.ctx, host)
	}

	return RobotsEntry{}, nil
}
