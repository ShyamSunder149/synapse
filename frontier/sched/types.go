package sched

import (
	"context"
	"time"
)

type PrefetchState struct {
	Capacity int
	Size     int
}

type FlushState struct {
	Capacity int
	Size     int
}

type PrefetchDecision struct {
	ShouldFetch bool
	Count       int
	Delay       time.Duration
}

type FlushDecision struct {
	ShouldFlush bool
	Count       int // 0 will flush all
	Delay       time.Duration
}

type BufferPolicy interface {
	ShouldPrefetch(ctx context.Context, state PrefetchState) PrefetchDecision
	ShouldFlush(ctx context.Context, state FlushState) FlushDecision
}
