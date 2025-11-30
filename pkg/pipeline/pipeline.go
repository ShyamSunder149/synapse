package pipeline

import (
	"fmt"
	"io"
)

type ExecutionContext[T any] struct {
	parser     Parser[T]
	processors []Processor[T]
	sink       Sink[T]
}

func (ctx *ExecutionContext[T]) ExecuteDefault(r io.Reader) error {
	data, err := ctx.parser.Parse(r)
	if err != nil {
		return err
	}

	for _, proc := range ctx.processors {
		data, err = proc.Process(data)
		if err != nil {
			return fmt.Errorf("processor %T failed: %w", proc, err)
		}
	}

	if err := ctx.sink.Write(data); err != nil {
		return fmt.Errorf("sink failed: %w", err)
	}

	return nil
}

type Pipeline[T any] struct {
	parser      Parser[T]
	processors  []Processor[T]
	sink        Sink[T]
	contentType string
	executor    func(*ExecutionContext[T], io.Reader) error
}

func NewPipeline[T any](contentType string, parser Parser[T], opts ...PipelineOption[T]) (*Pipeline[T], error) {
	p := &Pipeline[T]{
		parser:      parser,
		contentType: contentType,
	}

	for _, opt := range opts {
		opt(p)
	}

	if p.processors == nil {
		p.processors = []Processor[T]{}
	}

	if p.sink == nil {
		p.sink = DiscardSink[T]{}
	}

	if p.executor == nil {
		p.executor = func(ctx *ExecutionContext[T], r io.Reader) error {
			return ctx.ExecuteDefault(r)
		}
	}

	return p, nil
}
