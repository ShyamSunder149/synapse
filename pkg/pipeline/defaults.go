package pipeline

import "io"

type DiscardParser[T any] struct{}

func (d DiscardParser[T]) Parse(r io.Reader) (T, error) {
	var zero T
	return zero, nil
}

type DiscardSink[T any] struct{}

func (w DiscardSink[T]) Write(data T) error {
	return nil
}

func (w DiscardSink[T]) Close() error {
	return nil
}
