package robots

import (
	"context"
	"time"
)

// Blocking operation: called from dedicated workers
type RobotsTxtFetcher interface {
	Fetch(ctx context.Context, host string) (RobotsEntry, error)
}

type RobotsTxtBackend interface {
	Set(ctx context.Context, entry RobotsEntry) error
	Get(ctx context.Context, domain string) (RobotsEntry, error)
	Has(ctx context.Context, domain string) (bool, error)
}

type RobotsEntry struct {
	CrawlDelay time.Duration
	IsAllowed  bool
}
