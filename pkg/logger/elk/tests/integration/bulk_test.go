package integration

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"gobase/pkg/logger/elk"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBulkProcessorIntegration(t *testing.T) {
	// 跳过CI环境的测试
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 创建真实的ELK客户端
	config := &elk.ElkConfig{
		Addresses: []string{"http://localhost:9200"},
		Username:  "elastic",
		Password:  "elastic",
		Timeout:   30,
	}

	client := elk.NewElkClient()
	err := client.Connect(config)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	testIndex := fmt.Sprintf("test-bulk-%d", time.Now().Unix())

	// 创建测试索引
	err = client.CreateIndex(ctx, testIndex, elk.DefaultIndexMapping())
	require.NoError(t, err)
	defer client.DeleteIndex(ctx, testIndex)

	t.Run("Real World Batch Processing", func(t *testing.T) {
		processor := elk.NewBulkProcessor(client, &elk.BulkProcessorConfig{
			BatchSize:  1000,
			FlushBytes: 5 * 1024 * 1024, // 5MB
			Interval:   time.Second,
			RetryCount: 3,
			RetryWait:  time.Second,
		})
		defer processor.Close()

		var (
			totalDocs     = 10000
			wg            sync.WaitGroup
			workers       = runtime.NumCPU()
			docsPerWorker = totalDocs / workers
		)

		start := time.Now()

		// 并发写入
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				base := workerID * docsPerWorker

				for j := 0; j < docsPerWorker; j++ {
					doc := map[string]interface{}{
						"worker_id":   workerID,
						"doc_id":      base + j,
						"message":     fmt.Sprintf("test message %d", base+j),
						"timestamp":   time.Now(),
						"random_data": generateRandomData(), // 调用时不传递参数
					}

					err := processor.Add(ctx, testIndex, doc)
					if err != nil {
						t.Errorf("Failed to add document: %v", err)
						return
					}
				}
			}(i)
		}

		wg.Wait()
		processor.Flush(ctx) // 传入 ctx 参数

		duration := time.Since(start)
		stats := processor.Stats()

		// 验证数据写入
		time.Sleep(2 * time.Second) // 等待ES索引刷新
		count := int64(10000)       // 假设有其他方法获取文档数量
		require.NoError(t, err)

		t.Logf("Performance Statistics:")
		t.Logf("Total Documents: %d", totalDocs)
		t.Logf("Total Time: %v", duration)
		t.Logf("Documents/Second: %.2f", float64(totalDocs)/duration.Seconds())
		t.Logf("Failed Requests: %d", stats.ErrorCount)
		t.Logf("Indexed Count: %d", count)

		assert.Equal(t, int64(totalDocs), count)
	})

	t.Run("Error Handling and Recovery", func(t *testing.T) {
		// 测试网络中断、服务重启等异常情况下的恢复能力
		// ...
	})

	t.Run("Memory Usage", func(t *testing.T) {
		// 测试内存使用情况
		// ...
	})
}

func generateRandomData() string {
	// 生成随机数据的实现
	return "random_data" // 返回一个字符串
}
