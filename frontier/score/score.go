package score

import (
	"context"

	model "github.com/ritvikos/synapse/model"
)

type Score[T any] interface {
	Score(ctx context.Context, item *model.Task[T]) (float64, error)
}
