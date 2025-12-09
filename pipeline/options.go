package pipeline

import "io"

type PipelineOption[T any] func(*Pipeline[T])

func WithProcessors[T any](procs ...Processor[T]) PipelineOption[T] {
	return func(p *Pipeline[T]) {
		p.processors = append(p.processors, procs...)
	}
}

func WithSink[T any](sink Sink[T]) PipelineOption[T] {
	return func(p *Pipeline[T]) {
		p.sink = sink
	}
}

func WithExecutor[T any](executor func(*ExecutionContext[T], io.Reader) error) PipelineOption[T] {
	return func(p *Pipeline[T]) {
		p.executor = executor
	}
}

func (p Pipeline[T]) Execute(r io.Reader) error {
	return p.executor(&ExecutionContext[T]{
		parser:     p.parser,
		processors: p.processors,
		sink:       p.sink,
	}, r)
}

func (p Pipeline[T]) ContentType() string {
	return p.contentType
}
