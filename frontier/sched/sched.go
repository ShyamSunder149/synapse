package sched

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ritvikos/synapse/frontier/source"
	"github.com/ritvikos/synapse/model"
)

// TODO: Track relevent metrics for decision making in PrefetchState and FlushState in Scheduler
// Or combine them into a single State, if relevant.

type Scheduler[T any] struct {
	source source.Source[T]
	policy BufferPolicy

	prefetchBuf  chan *model.Task[T]
	flushBuf     chan *model.Task[T]
	tickInterval time.Duration

	// Internal
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	wg     sync.WaitGroup
}

func NewScheduler[T any](
	source source.Source[T],
	policy BufferPolicy,
	prefetchBufSize uint,
	flushBufSize uint,
	tickInterval time.Duration,
) *Scheduler[T] {
	return &Scheduler[T]{
		source:       source,
		policy:       policy,
		prefetchBuf:  make(chan *model.Task[T], prefetchBufSize),
		flushBuf:     make(chan *model.Task[T], flushBufSize),
		tickInterval: tickInterval,
	}
}

func (s *Scheduler[T]) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		return fmt.Errorf("Scheduler: already started")
	}

	s.ctx, s.cancel = context.WithCancel(ctx)
	s.mu.Unlock()

	s.wg.Add(2)
	go s.prefetchWorker()
	go s.flushWorker()

	return nil
}

func (s *Scheduler[T]) Stop(ctx context.Context) error {
	s.mu.Lock()
	if s.cancel == nil {
		s.mu.Unlock()
		return fmt.Errorf("Scheduler: not started")
	}

	s.cancel()
	s.cancel = nil
	s.mu.Unlock()

	s.wg.Wait()
	return nil
}

func (s *Scheduler[T]) Write(task *model.Task[T]) error {
	select {
	case <-s.ctx.Done():
		return s.ctx.Err()
	case s.flushBuf <- task:
		return nil
	}
}

func (s *Scheduler[T]) prefetchWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.checkAndPrefetch()
		}
	}
}

func (s *Scheduler[T]) checkAndPrefetch() {
	state := PrefetchState{
		Capacity: cap(s.prefetchBuf),
		Size:     len(s.prefetchBuf),
	}

	decision := s.policy.ShouldPrefetch(s.ctx, state)
	shouldFetch := decision.ShouldFetch
	if !shouldFetch {
		return
	}

	if shouldFetch && decision.Delay > 0 {
		time.Sleep(decision.Delay)
	}

	fetchCount := decision.Count
	if fetchCount == 0 {
		return
	}

	if err := s.prefetch(fetchCount); err != nil {
		_ = fmt.Errorf("prefetch error: %v\n", err)
	}
}

func (s *Scheduler[T]) prefetch(count int) error {
	tasks, err := s.source.Consume(s.ctx, count)
	if err != nil {
		return fmt.Errorf("Scheduler: prefetch dequeue error: %v\n", err)
	}

	for _, task := range tasks {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		case s.prefetchBuf <- task:
		}
	}

	return nil
}

func (s *Scheduler[T]) flushWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			s.flush(0) // flush all
			return
		case <-ticker.C:
			s.checkAndFlush()
		}
	}
}

func (s *Scheduler[T]) checkAndFlush() {
	state := FlushState{
		Capacity: cap(s.flushBuf),
		Size:     len(s.flushBuf),
	}

	decision := s.policy.ShouldFlush(s.ctx, state)
	shouldFlush := decision.ShouldFlush

	if shouldFlush {
		return
	}

	if shouldFlush && decision.Delay > 0 {
		time.Sleep(decision.Delay)
	}

	flushCount := decision.Count

	s.flush(flushCount)
}

func (s *Scheduler[T]) flush(count int) error {
	flushCount := count
	if flushCount == 0 {
		flushCount = len(s.flushBuf)
	}

	if flushCount == 0 {
		return nil
	}

	tasks := make([]*model.Task[T], 0, flushCount)

LOOP:
	for range flushCount {
		select {
		case task := <-s.flushBuf:
			tasks = append(tasks, task)
		default:
			break LOOP
		}
	}

	if len(tasks) == 0 {
		return nil
	}

	if err := s.source.Produce(s.ctx, tasks); err != nil {
		return fmt.Errorf("scheduler: flush enqueue error: %v\n", err)
	}

	return nil
}
