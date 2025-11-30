package score

import (
	"context"

	model "github.com/ritvikos/synapse/pkg/model"
)

type ScorePolicy[T any] interface {
	Score(ctx context.Context, item *model.Task[T]) (float64, error)
}
