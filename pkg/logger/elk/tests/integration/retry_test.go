package integration

import (
	"context"
	"fmt"
	"gobase/pkg/logger/elk"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ = elk.RetryConfig{} // 验证类型

func TestRetryMechanism(t *testing.T) {
	t.Run("RetryOnConnectionFailure", func(t *testing.T) {
		var requestCount int32
		logger := logrus.New()
		logger.SetLevel(logrus.DebugLevel)

		retryConfig := elk.RetryConfig{
			MaxRetries:  2,
			InitialWait: 10 * time.Millisecond,
			MaxWait:     50 * time.Millisecond,
		}

		// 在重试机制开始之前，确保请求计数器被正确初始化
		atomic.StoreInt32(&requestCount, 0)

		// 启动服务器
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&requestCount, 1)

			w.Header().Set("X-Elastic-Product", "Elasticsearch")
			w.Header().Set("content-type", "application/json")

			// 检查是否是连接检查请求
			if r.URL.Path == "/" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"version":{"number":"8.12.0"}}`))
				return
			}

			if count < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"error":{"type":"server_error","reason":"Service Unavailable"}}`))
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"acknowledged":true}`))
		}))
		defer server.Close()

		config := &elk.ElkConfig{
			Addresses: []string{server.URL},
			Timeout:   500 * time.Millisecond,
		}

		client := elk.NewElkClient()
		err := client.Connect(config)
		require.NoError(t, err)

		doc := map[string]interface{}{
			"message": "test retry",
			"time":    time.Now(),
		}

		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// 执行重试机制
		err = elk.WithRetry(ctxWithTimeout, retryConfig, func() error {
			return client.IndexDocument(ctxWithTimeout, "test-index", doc)
		}, logger)

		finalCount := atomic.LoadInt32(&requestCount)

		assert.NoError(t, err)
		// 不计算连接检查请求
		actualRequests := finalCount
		assert.LessOrEqual(t, actualRequests, int32(retryConfig.MaxRetries+1),
			fmt.Sprintf("总尝试次数不应超过%d次(1次初始尝试 + %d次重试)",
				retryConfig.MaxRetries+1, retryConfig.MaxRetries))
	})

	t.Run("RetryOnTimeout", func(t *testing.T) {
		var requestCount int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			count := atomic.AddInt32(&requestCount, 1)
			t.Logf("Request count: %d", count)

			w.Header().Set("X-Elastic-Product", "Elasticsearch")
			w.Header().Set("content-type", "application/json")

			// 强制超时
			time.Sleep(200 * time.Millisecond)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"acknowledged":true}`))
		}))
		defer server.Close()

		config := &elk.ElkConfig{
			Addresses: []string{server.URL},
			Timeout:   50 * time.Millisecond,
		}

		client := elk.NewElkClient()
		err := client.Connect(config)
		require.NoError(t, err)

		doc := map[string]interface{}{
			"message": "test timeout retry",
			"time":    time.Now(),
		}

		retryConfig := elk.RetryConfig{
			MaxRetries:  1,
			InitialWait: 10 * time.Millisecond,
			MaxWait:     50 * time.Millisecond,
		}

		ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		start := time.Now()
		err = elk.WithRetry(ctxWithTimeout, retryConfig, func() error {
			return client.IndexDocument(ctxWithTimeout, "test-index", doc)
		}, logrus.New())
		duration := time.Since(start)

		t.Logf("Operation took: %v", duration)
		t.Logf("Request count: %d", atomic.LoadInt32(&requestCount))

		assert.Error(t, err, "应该返回超时错误")
		if err != nil {
			assert.Contains(t, err.Error(), "deadline exceeded", "错误应该包含超时信息")
		}
	})
}
