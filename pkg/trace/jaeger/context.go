package jaeger

import (
	"context"
	"fmt"
	"net/http"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc/metadata"
)

// Carrier 定义传播载体接口
type Carrier interface {
	Set(key, val string)
	Get(key string) string
	ForeachKey(handler func(key, val string) error) error
}

// HTTPCarrier HTTP头部载体
type HTTPCarrier http.Header

// Set 实现 Carrier 接口
func (c HTTPCarrier) Set(key, val string) {
	http.Header(c).Set(key, val)
}

// Get 实现 Carrier 接口
func (c HTTPCarrier) Get(key string) string {
	return http.Header(c).Get(key)
}

// ForeachKey 实现 Carrier 接口
func (c HTTPCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range c {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// MetadataCarrier gRPC元数据载体
type MetadataCarrier metadata.MD

// Set 实现 Carrier 接口
func (c MetadataCarrier) Set(key, val string) {
	metadata.MD(c).Set(key, val)
}

// Get 实现 Carrier 接口
func (c MetadataCarrier) Get(key string) string {
	vals := metadata.MD(c).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// ForeachKey 实现 Carrier 接口
func (c MetadataCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range metadata.MD(c) {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// carrierAdapter 适配我们的 Carrier 到 opentracing.TextMapCarrier
type carrierAdapter struct {
	Carrier
}

// Inject 注入追踪上下文到载体
func Inject(ctx context.Context, carrier Carrier) error {
	span := opentracing.SpanFromContext(ctx)
	if span == nil {
		return nil
	}

	// 使用适配器包装 carrier
	adapter := carrierAdapter{carrier}
	err := opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.TextMap,
		adapter,
	)
	if err != nil {
		return errors.NewThirdPartyError(fmt.Sprintf("[%s] failed to inject context", codes.JaegerPropagateError), err)
	}

	return nil
}

// Extract 从载体提取追踪上下文
func Extract(carrier Carrier) (opentracing.SpanContext, error) {
	// 使用适配器包装 carrier
	adapter := carrierAdapter{carrier}
	spanCtx, err := opentracing.GlobalTracer().Extract(
		opentracing.TextMap,
		adapter,
	)
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		return nil, errors.NewThirdPartyError(fmt.Sprintf("[%s] failed to extract context", codes.JaegerPropagateError), err)
	}

	return spanCtx, nil
}

// StartSpanFromContext 从上下文开始新的Span
func StartSpanFromContext(ctx context.Context, operationName string, opts ...opentracing.StartSpanOption) (opentracing.Span, context.Context) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		opts = append(opts, opentracing.ChildOf(span.Context()))
	}
	span := opentracing.StartSpan(operationName, opts...)
	return span, opentracing.ContextWithSpan(ctx, span)
}

// StartSpanFromHTTP 从HTTP请求开始新的Span
func StartSpanFromHTTP(r *http.Request, operationName string) (opentracing.Span, error) {
	carrier := HTTPCarrier(r.Header)
	spanCtx, err := Extract(carrier)
	if err != nil {
		return nil, err
	}

	span := opentracing.StartSpan(
		operationName,
		ext.RPCServerOption(spanCtx),
		ext.SpanKindRPCServer,
		opentracing.Tag{Key: string(ext.HTTPMethod), Value: r.Method},
		opentracing.Tag{Key: string(ext.HTTPUrl), Value: r.URL.String()},
	)

	return span, nil
}

// StartSpanFromGRPC 从gRPC元数据开始新的Span
func StartSpanFromGRPC(ctx context.Context, operationName string) (opentracing.Span, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	carrier := MetadataCarrier(md)
	spanCtx, err := Extract(carrier)
	if err != nil {
		return nil, err
	}

	span := opentracing.StartSpan(
		operationName,
		ext.RPCServerOption(spanCtx),
		ext.SpanKindRPCServer,
	)

	return span, nil
}
