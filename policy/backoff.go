package policy

import (
	"math"
	"time"
)

// Fixed interval between retries.
type ConstantBackoff struct {
	interval time.Duration
}

func (b ConstantBackoff) NextRetry(attempt uint) time.Duration {
	return b.interval
}

// The wait time increases exponentially with each attempt.
type ExponentialBackoff struct {
	// Initial interval
	base time.Duration

	// Multiplier for each subsequent attempt
	factor float64
}

func (b ExponentialBackoff) NextRetry(attempt int) time.Duration {
	return time.Duration(float64(b.base) * math.Pow(b.factor, float64(attempt-1)))
}

func DefaultExponentialBackoff() ExponentialBackoff {
	return ExponentialBackoff{
		base:   500 * time.Millisecond,
		factor: 2.0,
	}
}
