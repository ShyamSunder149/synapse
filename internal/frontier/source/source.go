package source

import (
	"context"

	"github.com/ritvikos/synapse/internal/lifecycle"
	model "github.com/ritvikos/synapse/pkg/model"
)

type Source[T any] interface {
	lifecycle.Lifecycle

	Consume(ctx context.Context, batchSize uint) ([]*model.Task[T], error)

	Produce(ctx context.Context, task []*model.Task[T]) error

	Count(ctx context.Context) int
}
