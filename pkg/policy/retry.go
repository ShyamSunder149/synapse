package policy

import (
	"time"
)

type RetryController struct {
	maxRetries int
	backoff    BackoffPolicy
	attempts   int
}

func DefaultRetryController(maxRetries int) RetryController {
	strategy := ExponentialBackoff{
		base:   1 * time.Second,
		factor: 2.0,
	}
	return RetryController{
		maxRetries: maxRetries,
		backoff:    strategy,
		attempts:   0,
	}
}

func (rc *RetryController) Next() (time.Duration, bool) {
	rc.attempts++
	if rc.attempts > rc.maxRetries {
		return 0, false
	}
	return rc.backoff.NextRetry(rc.attempts), true
}
