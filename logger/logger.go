package logger

import (
	"context"
	"log/slog"
	"os"
)

type slogWrapper struct {
	log *slog.Logger
}

type Logger interface {
	Info(ctx context.Context, msg string, args map[string]any)
	Error(ctx context.Context, err error, args map[string]any)
}

var _ Logger = (*slogWrapper)(nil)

func New() Logger {
	slogLogger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{}),
	)

	return &slogWrapper{log: slogLogger}
}

func (sw *slogWrapper) Info(ctx context.Context, msg string, attrs map[string]any) {
	sw.log.LogAttrs(ctx, slog.LevelInfo, msg, sw.mapAttrs(attrs)...)
}

func (sw *slogWrapper) Error(ctx context.Context, err error, attrs map[string]any) {
	sw.log.LogAttrs(ctx, slog.LevelError, err.Error(), sw.mapAttrs(attrs)...)
}

func (sw *slogWrapper) mapAttrs(attrs map[string]any) []slog.Attr {
	logAttrs := make([]slog.Attr, 0, len(attrs))

	if attrs == nil {
		return logAttrs
	}

	for k, v := range attrs {
		logAttrs = append(logAttrs, slog.Attr{Key: k, Value: slog.AnyValue(v)})
	}

	return logAttrs
}
