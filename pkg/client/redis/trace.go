package redis

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// startSpan 创建一个新的追踪span
func startSpan(ctx context.Context, tracer opentracing.Tracer, operationName string) (opentracing.Span, context.Context) {
	var span opentracing.Span
	if parent := opentracing.SpanFromContext(ctx); parent != nil {
		span = tracer.StartSpan(operationName, opentracing.ChildOf(parent.Context()))
	} else {
		span = tracer.StartSpan(operationName)
	}

	ext.DBType.Set(span, "redis")
	return span, opentracing.ContextWithSpan(ctx, span)
}
