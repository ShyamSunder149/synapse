package model

import (
	"net/url"
	"time"
)

type Task[T any] struct {
	Url       *url.URL
	ExecuteAt time.Time
	Score     float64
	Metadata  *T
	// Fingerprint string
}
