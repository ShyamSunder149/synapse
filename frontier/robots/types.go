package robots

import (
	"context"
	"time"

	"github.com/ritvikos/synapse/internal/lifecycle"
)

// Blocking operation: called from dedicated workers
type RobotsTxtFetcher interface {
	lifecycle.Lifecycle
	Fetch(ctx context.Context, host string) (RobotsEntry, error)
}

type RobotsTxtBackend interface {
	lifecycle.Lifecycle
	Set(ctx context.Context, entry RobotsEntry) error
	Get(ctx context.Context, domain string) (RobotsEntry, error)
	Has(ctx context.Context, domain string) (bool, error)
}

type RobotsEntry struct {
	CrawlDelay time.Duration
	IsAllowed  bool
}
