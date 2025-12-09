package source

import (
	"context"

	"github.com/Workiva/go-datastructures/queue"
	"github.com/ritvikos/synapse/model"
)

type MemQueue[T any] struct {
	queue queue.PriorityQueue
}

type taskItem[T any] struct {
	task *model.Task[T]
}

func (t *taskItem[T]) Compare(other queue.Item) int {
	otherTask, ok := other.(*taskItem[T])
	if !ok {
		return 0
	}
	if t.task.Score < otherTask.task.Score {
		return 1
	} else if t.task.Score > otherTask.task.Score {
		return -1
	}
	return 0
}

func NewMemQueue[T any](hint int) *MemQueue[T] {
	return &MemQueue[T]{
		queue: *queue.NewPriorityQueue(hint, false),
	}
}

func (m *MemQueue[T]) Start(ctx context.Context) error {
	return nil
}

func (m *MemQueue[T]) Stop(ctx context.Context) error {
	return nil
}

func (m *MemQueue[T]) Produce(ctx context.Context, tasks []*model.Task[T]) error {
	items := make([]queue.Item, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, &taskItem[T]{task: task})
	}
	return m.queue.Put(items...)
}

func (m *MemQueue[T]) Consume(ctx context.Context, batchSize int) ([]*model.Task[T], error) {
	items, err := m.queue.Get(batchSize)
	if err != nil {
		return nil, err
	}

	tasks := make([]*model.Task[T], 0, len(items))
	for _, item := range items {
		i := item.(*taskItem[T])
		tasks = append(tasks, i.task)
	}

	return tasks, nil
}

func (m *MemQueue[T]) Count(ctx context.Context) int {
	return m.queue.Len()
}
