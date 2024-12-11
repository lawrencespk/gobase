package jaeger

import (
	"context"
	"fmt"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

// SpanOption Span选项
type SpanOption func(*SpanOptions)

// SpanOptions Span选项集合
type SpanOptions struct {
	// 父级上下文
	Parent context.Context
	// 标签
	Tags map[string]interface{}
	// 是否采样
	Sample bool
	// 开始时间
	StartTime time.Time
	// 引用类型
	References []opentracing.SpanReference
}

// Span 封装opentracing.Span
type Span struct {
	tracer opentracing.Tracer
	span   opentracing.Span
	ctx    context.Context
}

// NewSpan 创建新的Span
func NewSpan(operationName string, opts ...SpanOption) (*Span, error) {
	options := &SpanOptions{
		Parent:    context.Background(),
		Tags:      make(map[string]interface{}),
		Sample:    true,
		StartTime: time.Now(),
	}

	// 应用选项
	for _, opt := range opts {
		opt(options)
	}

	// 获取tracer
	tracer := opentracing.GlobalTracer()
	if tracer == nil {
		return nil, errors.NewThirdPartyError(fmt.Sprintf("[%s] global tracer not set", codes.JaegerSpanError), nil)
	}

	// 创建span选项
	spanOpts := []opentracing.StartSpanOption{
		opentracing.StartTime(options.StartTime),
	}

	// 添加父级上下文
	if parentSpan := opentracing.SpanFromContext(options.Parent); parentSpan != nil {
		spanOpts = append(spanOpts, opentracing.ChildOf(parentSpan.Context()))
	}

	// 添加引用
	if len(options.References) > 0 {
		for _, ref := range options.References {
			spanOpts = append(spanOpts, ref)
		}
	}

	// 添加标签
	for k, v := range options.Tags {
		spanOpts = append(spanOpts, opentracing.Tag{Key: k, Value: v})
	}

	// 设置采样
	if !options.Sample {
		spanOpts = append(spanOpts, opentracing.Tag{Key: "sampling.priority", Value: 0})
	}

	// 创建span
	span := tracer.StartSpan(operationName, spanOpts...)

	// 创建新的上下文
	ctx := opentracing.ContextWithSpan(options.Parent, span)

	return &Span{
		tracer: tracer,
		span:   span,
		ctx:    ctx,
	}, nil
}

// Context 获取上下文
func (s *Span) Context() context.Context {
	return s.ctx
}

// SetTag 设置标签
func (s *Span) SetTag(key string, value interface{}) {
	s.span.SetTag(key, value)
}

// SetBaggageItem 设置透传数据
func (s *Span) SetBaggageItem(key, value string) {
	s.span.SetBaggageItem(key, value)
}

// LogFields 记录字段
func (s *Span) LogFields(fields ...log.Field) {
	s.span.LogFields(fields...)
}

// LogKV 记录KV
func (s *Span) LogKV(alternatingKeyValues ...interface{}) {
	s.span.LogKV(alternatingKeyValues...)
}

// SetError 设置错误
func (s *Span) SetError(err error) {
	ext.Error.Set(s.span, true)
	s.span.LogFields(
		log.String("event", "error"),
		log.String("error.kind", fmt.Sprintf("%T", err)),
		log.String("error.message", err.Error()),
		log.String("stack", fmt.Sprintf("%+v", err)),
	)
}

// Finish 结束Span
func (s *Span) Finish() {
	s.span.Finish()
}

// FinishWithOptions 使用选项结束Span
func (s *Span) FinishWithOptions(opts opentracing.FinishOptions) {
	s.span.FinishWithOptions(opts)
}

// WithTag Span选项:添加标签
func WithTag(key string, value interface{}) SpanOption {
	return func(o *SpanOptions) {
		o.Tags[key] = value
	}
}

// WithParent Span选项:设置父级上下文
func WithParent(ctx context.Context) SpanOption {
	return func(o *SpanOptions) {
		o.Parent = ctx
	}
}

// WithStartTime Span选项:设置开始时间
func WithStartTime(t time.Time) SpanOption {
	return func(o *SpanOptions) {
		o.StartTime = t
	}
}

// WithSample Span选项:设置是否采样
func WithSample(sample bool) SpanOption {
	return func(o *SpanOptions) {
		o.Sample = sample
	}
}

// WithReferences Span选项:添加引用
func WithReferences(refs ...opentracing.SpanReference) SpanOption {
	return func(o *SpanOptions) {
		o.References = append(o.References, refs...)
	}
}
