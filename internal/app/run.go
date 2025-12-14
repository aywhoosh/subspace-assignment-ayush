package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type ctxKey int

const (
	ctxKeyRunID ctxKey = iota
	ctxKeyLogger
)

func NewRunID() string {
	return uuid.NewString()
}

func WithRunID(ctx context.Context, runID string) context.Context {
	return context.WithValue(ctx, ctxKeyRunID, runID)
}

func RunID(ctx context.Context) string {
	if v := ctx.Value(ctxKeyRunID); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func WithLogger(ctx context.Context, log zerolog.Logger) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, log)
}

func Logger(ctx context.Context, fallback zerolog.Logger) zerolog.Logger {
	if v := ctx.Value(ctxKeyLogger); v != nil {
		if l, ok := v.(zerolog.Logger); ok {
			return l
		}
	}
	return fallback
}
