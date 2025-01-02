package redis

import (
	"context"
	"gobase/pkg/trace/jaeger"
)

// startSpan 创建一个新的追踪span
func startSpan(ctx context.Context, tracer *jaeger.Provider, operationName string) (*jaeger.Span, context.Context) {
	if tracer == nil {
		return nil, ctx
	}

	span, err := jaeger.NewSpan(operationName,
		jaeger.WithParent(ctx),
		jaeger.WithTag("db.type", "redis"),
	)
	if err != nil {
		return nil, ctx
	}
	return span, span.Context()
}
