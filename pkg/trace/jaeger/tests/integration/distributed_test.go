package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/trace/jaeger"
)

// 模拟微服务
type mockService struct {
	name    string
	tracer  opentracing.Tracer
	handler http.HandlerFunc
}

func newMockService(name string, next *mockService) (*mockService, error) {
	// 创建服务专用的tracer
	provider, err := jaeger.NewProvider()
	if err != nil {
		log.Errorf(context.Background(), "create provider error: %v", err)
		return nil, fmt.Errorf("create provider error: %v", err)
	}

	svc := &mockService{
		name:   name,
		tracer: provider.Tracer(),
	}

	// 设置HTTP处理函数
	svc.handler = func(w http.ResponseWriter, r *http.Request) {
		// 从请求中提取span context
		spanCtx, _ := svc.tracer.Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header),
		)

		// 创建服务的span
		span := svc.tracer.StartSpan(
			fmt.Sprintf("%s.handle", name),
			ext.RPCServerOption(spanCtx),
		)
		defer span.Finish()

		// 记录一些span信息
		span.SetTag("service.name", name)
		span.LogKV("event", "handling request")

		// 如果有下游服务,则继续调用
		if next != nil {
			nextReq := httptest.NewRequest("GET", "/", nil)

			// 注入span context到下游请求
			err := svc.tracer.Inject(
				span.Context(),
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(nextReq.Header),
			)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			rec := httptest.NewRecorder()
			next.handler(rec, nextReq)

			// 传递下游响应
			for k, v := range rec.Header() {
				w.Header()[k] = v
			}
			w.WriteHeader(rec.Code)
			w.Write(rec.Body.Bytes())
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}

	return svc, nil
}

func TestDistributedTracing(t *testing.T) {
	helper := setupTest(t)
	defer helper.cleanup()

	// 启动多个服务实例的测试
	t.Run("multi_service_tracing", func(t *testing.T) {
		// 创建服务链 A -> B -> C
		svcC, err := newMockService("service-c", nil)
		require.NoError(t, err)

		svcB, err := newMockService("service-b", svcC)
		require.NoError(t, err)

		svcA, err := newMockService("service-a", svcB)
		require.NoError(t, err)

		// 发起请求
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		// 创建根span
		span := svcA.tracer.StartSpan("test.request")
		defer span.Finish()

		// 注入span context
		err = svcA.tracer.Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(req.Header),
		)
		require.NoError(t, err)

		// 执行请求
		svcA.handler(rec, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, rec.Code)

		// 使用统一的等待函数
		WaitForSpans(2 * time.Second)
	})

	// 测试跨进程传播
	t.Run("cross_process_propagation", func(t *testing.T) {
		// 创建两个服务
		svcB, err := newMockService("service-b", nil)
		require.NoError(t, err)

		svcA, err := newMockService("service-a", svcB)
		require.NoError(t, err)

		// 创建带有baggage item的span
		span := svcA.tracer.StartSpan("test.baggage")
		span.SetBaggageItem("test.key", "test.value")
		defer span.Finish()

		// 创建请求并注入context
		req := httptest.NewRequest("GET", "/", nil)
		err = svcA.tracer.Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(req.Header),
		)
		require.NoError(t, err)

		// 执行请求
		rec := httptest.NewRecorder()
		svcA.handler(rec, req)

		// 验证响应
		assert.Equal(t, http.StatusOK, rec.Code)

		// 使用统一的等待函数
		WaitForSpans(2 * time.Second)
	})
}
