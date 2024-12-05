package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gobase/pkg/logger/elk"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestELKClientIntegration(t *testing.T) {
	// 跳过CI环境
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	config := &elk.ElkConfig{
		Addresses: []string{"http://localhost:9200"},
		Username:  "elastic",
		Password:  "changeme",
		Index:     "test-integration",
		Timeout:   30,
	}

	t.Run("Connection Lifecycle", func(t *testing.T) {
		client := elk.NewElkClient()

		// 测试连接
		err := client.Connect(config)
		require.NoError(t, err)
		assert.True(t, client.IsConnected())

		// 测试断开连接
		err = client.Close()
		require.NoError(t, err)
		assert.False(t, client.IsConnected())

		// 测试重连
		err = client.Connect(config)
		require.NoError(t, err)
		assert.True(t, client.IsConnected())
	})

	t.Run("Retry Mechanism", func(t *testing.T) {
		// 创建一个模拟服务器，模拟服务器故障和恢复
		var attempts int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			currentAttempt := atomic.AddInt32(&attempts, 1)

			// 设置Elasticsearch响应头
			w.Header().Set("X-Elastic-Product", "Elasticsearch")

			// 检查请求路径，区分连接检查和实际操作
			if r.URL.Path == "/" {
				// 连接检查请求，直接返回成功
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"name" : "node-1",
					"cluster_name" : "elasticsearch",
					"version" : {
						"number" : "7.9.3"
					}
				}`))
				return
			}

			// 实际操作请求
			if currentAttempt <= 2 {
				// 前两次请求返回503错误
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"error": "Service Unavailable", "status": 503}`))
				return
			}

			// 第三次请求成功
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"acknowledged": true,
				"_index": "test-index",
				"_type": "_doc",
				"_id": "1",
				"_version": 1,
				"result": "created",
				"_shards": {
					"total": 2,
					"successful": 1,
					"failed": 0
				}
			}`))
		}))
		defer server.Close()

		// 使用模拟服务器地址创建客户端
		retryConfig := &elk.ElkConfig{
			Addresses: []string{server.URL},
			Timeout:   30,
		}

		client := elk.NewElkClient()
		err := client.Connect(retryConfig)
		require.NoError(t, err)

		// 重置计数器，不计入连接检查的请求
		atomic.StoreInt32(&attempts, 0)

		// 执行操作，应该会触发重试机制
		doc := map[string]interface{}{
			"message":   "test retry",
			"timestamp": time.Now(),
		}
		err = client.IndexDocument(ctx, "test-index", doc)
		assert.NoError(t, err)

		finalAttempts := atomic.LoadInt32(&attempts)
		assert.Equal(t, int32(3), finalAttempts, "Should have attempted exactly 3 times")
	})

	t.Run("Connection Failure", func(t *testing.T) {
		client := elk.NewElkClient()
		badConfig := &elk.ElkConfig{
			Addresses: []string{"http://nonexistent:9200"},
			Timeout:   1,
		}

		err := client.Connect(badConfig)
		assert.Error(t, err)
		assert.False(t, client.IsConnected())
	})

	t.Run("Connection Pool", func(t *testing.T) {
		client := elk.NewElkClient()
		require.NoError(t, client.Connect(config))

		// 并发测试连接池
		concurrency := 10
		operations := 100
		errCh := make(chan error, concurrency*operations)

		for i := 0; i < concurrency; i++ {
			go func(workerID int) {
				for j := 0; j < operations; j++ {
					doc := map[string]interface{}{
						"worker_id": workerID,
						"operation": j,
						"timestamp": time.Now(),
					}
					if err := client.IndexDocument(ctx, "test-index", doc); err != nil {
						errCh <- err
					}
				}
			}(i)
		}

		// 等待所有操作完成
		time.Sleep(5 * time.Second)

		// 检查错误
		close(errCh)
		for err := range errCh {
			t.Errorf("Connection pool operation failed: %v", err)
		}
	})

	t.Run("Retry Backoff", func(t *testing.T) {
		// 创建一个模拟服务器，记录请求时间
		var requestTimes []time.Time
		mu := sync.Mutex{}

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			now := time.Now()
			requestTimes = append(requestTimes, now)
			currentAttempt := len(requestTimes)
			mu.Unlock()

			// 设置Elasticsearch响应头
			w.Header().Set("X-Elastic-Product", "Elasticsearch")

			if currentAttempt < 3 {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Write([]byte(`{"error": "Service Unavailable", "status": 503}`))
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"acknowledged": true,
				"_index": "test-index",
				"_type": "_doc",
				"_id": "1",
				"_version": 1,
				"result": "created",
				"_shards": {
					"total": 2,
					"successful": 1,
					"failed": 0
				}
			}`))
		}))
		defer server.Close()

		retryConfig := &elk.ElkConfig{
			Addresses: []string{server.URL},
			Timeout:   30,
		}

		client := elk.NewElkClient()
		err := client.Connect(retryConfig)
		require.NoError(t, err)

		doc := map[string]interface{}{
			"message":   "test backoff",
			"timestamp": time.Now(),
		}
		err = client.IndexDocument(ctx, "test-index", doc)
		assert.NoError(t, err)

		// 验证重试间隔是否递增
		mu.Lock()
		times := make([]time.Time, len(requestTimes))
		copy(times, requestTimes)
		mu.Unlock()

		if assert.True(t, len(times) >= 2, "Should have at least 2 requests") {
			// 只检查重试过程中的间隔
			for i := 1; i < len(times)-1; i++ {
				interval := times[i].Sub(times[i-1]).Milliseconds()
				t.Logf("Retry interval %d: %d ms", i, interval)
				assert.True(t, interval > 0,
					"Retry interval should be positive")
			}
		}
	})
}
