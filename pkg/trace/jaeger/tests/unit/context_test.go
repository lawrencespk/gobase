package unit

import (
	"context"
	"net/http"
	"testing"

	"gobase/pkg/trace/jaeger"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"google.golang.org/grpc/metadata"
)

func TestHTTPCarrier(t *testing.T) {
	// 创建HTTP头
	header := http.Header{}
	carrier := jaeger.HTTPCarrier(header)

	// 测试Set方法
	carrier.Set("test-key", "test-value")
	if got := carrier.Get("test-key"); got != "test-value" {
		t.Errorf("HTTPCarrier.Get() = %v, want %v", got, "test-value")
	}

	// 测试Get方法对不存在的key
	if got := carrier.Get("non-existent"); got != "" {
		t.Errorf("HTTPCarrier.Get() = %v, want empty string", got)
	}
}

func TestMetadataCarrier(t *testing.T) {
	// 创建gRPC元数据
	md := metadata.New(nil)
	carrier := jaeger.MetadataCarrier(md)

	// 测试Set方法
	carrier.Set("test-key", "test-value")
	if got := carrier.Get("test-key"); got != "test-value" {
		t.Errorf("MetadataCarrier.Get() = %v, want %v", got, "test-value")
	}

	// 测试Get方法对不存在的key
	if got := carrier.Get("non-existent"); got != "" {
		t.Errorf("MetadataCarrier.Get() = %v, want empty string", got)
	}
}

func TestInjectExtract(t *testing.T) {
	// 使用mock tracer
	mockTracer := mocktracer.New()
	opentracing.SetGlobalTracer(mockTracer)

	// 创建一个span
	span := mockTracer.StartSpan("test-operation")
	ctx := opentracing.ContextWithSpan(context.Background(), span)

	// 测试HTTP carrier
	t.Run("HTTP carrier", func(t *testing.T) {
		header := http.Header{}
		carrier := jaeger.HTTPCarrier(header)

		// 注入
		err := jaeger.Inject(ctx, carrier)
		if err != nil {
			t.Fatalf("Inject() error = %v", err)
		}

		// 提取
		extractedCtx, err := jaeger.Extract(carrier)
		if err != nil {
			t.Fatalf("Extract() error = %v", err)
		}

		if extractedCtx == nil {
			t.Error("Extract() returned nil context")
		}
	})

	// 测试Metadata carrier
	t.Run("Metadata carrier", func(t *testing.T) {
		md := metadata.New(nil)
		carrier := jaeger.MetadataCarrier(md)

		// 注入
		err := jaeger.Inject(ctx, carrier)
		if err != nil {
			t.Fatalf("Inject() error = %v", err)
		}

		// 提取
		extractedCtx, err := jaeger.Extract(carrier)
		if err != nil {
			t.Fatalf("Extract() error = %v", err)
		}

		if extractedCtx == nil {
			t.Error("Extract() returned nil context")
		}
	})
}

func TestStartSpanFromContext(t *testing.T) {
	// 使用mock tracer
	mockTracer := mocktracer.New()
	opentracing.SetGlobalTracer(mockTracer)

	// 测试从空context开始
	t.Run("from empty context", func(t *testing.T) {
		span, ctx := jaeger.StartSpanFromContext(context.Background(), "test-operation")
		if span == nil {
			t.Error("StartSpanFromContext() returned nil span")
		}
		if ctx == nil {
			t.Error("StartSpanFromContext() returned nil context")
		}
		span.Finish()
	})

	// 测试从带span的context开始
	t.Run("from context with span", func(t *testing.T) {
		parentSpan := mockTracer.StartSpan("parent-operation")
		parentCtx := opentracing.ContextWithSpan(context.Background(), parentSpan)

		childSpan, ctx := jaeger.StartSpanFromContext(parentCtx, "child-operation")
		if childSpan == nil {
			t.Error("StartSpanFromContext() returned nil span")
		}
		if ctx == nil {
			t.Error("StartSpanFromContext() returned nil context")
		}

		childSpan.Finish()
		parentSpan.Finish()
	})
}

func TestStartSpanFromHTTP(t *testing.T) {
	// 使用mock tracer
	mockTracer := mocktracer.New()
	opentracing.SetGlobalTracer(mockTracer)

	// 创建HTTP请求
	req, _ := http.NewRequest("GET", "http://example.com", nil)

	// 创建并注入父span
	parentSpan := mockTracer.StartSpan("parent-operation")
	carrier := jaeger.HTTPCarrier(req.Header)
	err := mockTracer.Inject(parentSpan.Context(), opentracing.TextMap, carrier)
	if err != nil {
		t.Fatalf("Failed to inject span: %v", err)
	}

	// 测试从HTTP请求创建span
	span, err := jaeger.StartSpanFromHTTP(req, "test-operation")
	if err != nil {
		t.Fatalf("StartSpanFromHTTP() error = %v", err)
	}
	if span == nil {
		t.Error("StartSpanFromHTTP() returned nil span")
	}

	span.Finish()
	parentSpan.Finish()
}

func TestStartSpanFromGRPC(t *testing.T) {
	// 使用mock tracer
	mockTracer := mocktracer.New()
	opentracing.SetGlobalTracer(mockTracer)

	// 创建gRPC元数据
	md := metadata.New(nil)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	// 创建并注入父span
	parentSpan := mockTracer.StartSpan("parent-operation")
	carrier := jaeger.MetadataCarrier(md)
	err := mockTracer.Inject(parentSpan.Context(), opentracing.TextMap, carrier)
	if err != nil {
		t.Fatalf("Failed to inject span: %v", err)
	}

	// 测试从gRPC上下文创建span
	span, err := jaeger.StartSpanFromGRPC(ctx, "test-operation")
	if err != nil {
		t.Fatalf("StartSpanFromGRPC() error = %v", err)
	}
	if span == nil {
		t.Error("StartSpanFromGRPC() returned nil span")
	}

	span.Finish()
	parentSpan.Finish()
}
