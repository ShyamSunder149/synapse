package robots

import (
	"context"
	"fmt"
	"sync"

	"github.com/ritvikos/synapse/pkg/frontier/robots"
)

// TODO: Pending implementation

var RobotsTxtNeedsFetchError = robots.RobotsTxtNeedsFetchError

type RobotsTxtHandler struct {
	userAgent string
	fetcher   robots.RobotsTxtFetcher
	backend   robots.RobotsTxtBackend
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
	fetcher robots.RobotsTxtFetcher,
	backend robots.RobotsTxtBackend,
	requestChSize uint,
	workers uint,
) *RobotsTxtHandler {
	return &RobotsTxtHandler{
		userAgent: userAgent,
		fetcher:   fetcher,
		backend:   backend,
		requestCh: make(chan string, requestChSize),
		workers:   workers,
	}
}

func (r *RobotsTxtHandler) Start(ctx context.Context) error {
	r.ctx, r.cancel = context.WithCancel(ctx)

	for range r.workers {
		r.wg.Add(1)
		go r.fetchWorker()
	}

	return nil
}

func (r *RobotsTxtHandler) Stop() error {
	r.wg.Wait()
	return nil
}

func (r *RobotsTxtHandler) fetchWorker() {
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

func (r *RobotsTxtHandler) Retrieve(host string) (robots.RobotsEntry, error) {
	// try retrieve from store
	// if not found, fetch and parse

	if has, err := r.backend.Has(r.ctx, host); err != nil {
		return robots.RobotsEntry{}, err
	} else if has {
		return r.backend.Get(r.ctx, host)
	}

	return robots.RobotsEntry{}, nil
}

func (r *RobotsTxtHandler) Submit(ctx context.Context, host string) error {
	select {
	case r.requestCh <- host:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// func (r *RobotsTxtHandler) IsAllowed(ctx context.Context, url string) (bool, error) {
// 	return true, nil
// }
