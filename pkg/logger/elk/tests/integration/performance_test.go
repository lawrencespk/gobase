package integration

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gobase/pkg/logger/elk"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestELKPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}

	config := getElkConfig()
	client := elk.NewElkClient()
	ctx := context.Background()

	err := client.Connect(config)
	require.NoError(t, err)
	defer client.Close()

	testIndex := fmt.Sprintf("test-perf-%d", time.Now().Unix())

	// 创建测试索引
	err = client.CreateIndex(ctx, testIndex, elk.DefaultIndexMapping())
	require.NoError(t, err)
	defer client.DeleteIndex(ctx, testIndex)

	t.Run("BulkIndexingPerformance", func(t *testing.T) {
		var (
			totalDocs     = 100000               // 总文档数
			batchSize     = 5000                 // 每批次文档数
			workers       = runtime.NumCPU() * 2 // 工作协程数
			docsPerWorker = totalDocs / workers
			successCount  atomic.Int64
			errorCount    atomic.Int64
			totalBytes    atomic.Int64
			errorMsgs     sync.Map // 用于记录错误信息
		)

		processor := elk.NewBulkProcessor(client, &elk.BulkProcessorConfig{
			BatchSize:    batchSize,
			FlushBytes:   5 * 1024 * 1024, // 5MB
			Interval:     time.Second,
			RetryCount:   3,
			RetryWait:    100 * time.Millisecond,
			CloseTimeout: 30 * time.Second,
		})
		defer processor.Close()

		start := time.Now()
		var wg sync.WaitGroup

		// 并发写入
		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				base := workerID * docsPerWorker

				for j := 0; j < docsPerWorker; j++ {
					doc := generateTestDocument(base + j)
					docSize := len(fmt.Sprintf("%v", doc))

					err := processor.Add(ctx, testIndex, doc)
					if err != nil {
						errorCount.Add(1)
						errorMsgs.Store(fmt.Sprintf("worker-%d-doc-%d", workerID, base+j), err.Error())
						continue
					}

					successCount.Add(1)
					totalBytes.Add(int64(docSize))
				}
			}(i)
		}

		wg.Wait()
		err = processor.Flush(ctx)
		if err != nil {
			t.Logf("最终刷新时发生错误: %v", err)
		}

		duration := time.Since(start)
		stats := processor.Stats()

		// 等待ES索引刷新
		time.Sleep(2 * time.Second)

		// 输出性能统计
		t.Logf("\n性能测试结果:")
		t.Logf("总文档数: %d", totalDocs)
		t.Logf("总耗时: %v", duration)
		t.Logf("每秒文档数: %.2f", float64(successCount.Load())/duration.Seconds())
		t.Logf("每秒字节数: %.2f MB", float64(totalBytes.Load())/(duration.Seconds()*1024*1024))
		t.Logf("成功文档数: %d", successCount.Load())
		t.Logf("失败文档数: %d", errorCount.Load())
		t.Logf("批量操作次数: %d", stats.FlushCount)
		t.Logf("错误次数: %d", stats.ErrorCount)
		t.Logf("CPU核心数: %d", runtime.NumCPU())
		t.Logf("Goroutine数量: %d", workers)

		// 如果有错误，输出错误信息
		if errorCount.Load() > 0 {
			t.Log("\n错误详情:")
			errorMsgs.Range(func(key, value interface{}) bool {
				t.Logf("%v: %v", key, value)
				return true
			})
		}

		totalProcessed := successCount.Load() + errorCount.Load()
		successRate := float64(successCount.Load()) / float64(totalDocs) * 100

		t.Logf("成功率: %.2f%%", successRate)

		// 允许最多 1% 的文档处理偏差
		maxDeviation := int64(float64(totalDocs) * 0.01)
		actualDeviation := int64(totalDocs) - totalProcessed

		assert.LessOrEqual(t, actualDeviation, maxDeviation,
			"处理的文档总数偏差(%d)不应超过允许的最大偏差(%d), 成功数:%d, 失败数:%d",
			actualDeviation, maxDeviation, successCount.Load(), errorCount.Load())
		assert.GreaterOrEqual(t, successRate, 99.0, "成功率应大于99%")
	})

	t.Run("QueryPerformance", func(t *testing.T) {
		var (
			queryCount       = 1000
			workers          = runtime.NumCPU()
			queriesPerWorker = queryCount / workers
			successCount     atomic.Int64
			errorCount       atomic.Int64
			totalLatency     atomic.Int64
			errorMsgs        sync.Map // 用于记录错误信息
		)

		var wg sync.WaitGroup
		start := time.Now()

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for j := 0; j < queriesPerWorker; j++ {
					queryStart := time.Now()

					query := generateTestQuery(j)
					_, err := client.Query(ctx, testIndex, query)

					latency := time.Since(queryStart)
					totalLatency.Add(latency.Microseconds())

					if err != nil {
						errorCount.Add(1)
						errorMsgs.Store(fmt.Sprintf("worker-%d-query-%d", workerID, j), err.Error())
						continue
					}
					successCount.Add(1)
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		avgLatency := time.Duration(totalLatency.Load()/successCount.Load()) * time.Microsecond

		// 输出查询性能统计
		t.Logf("\n查询性能测试结果:")
		t.Logf("总查询数: %d", queryCount)
		t.Logf("总耗时: %v", duration)
		t.Logf("每秒查询数: %.2f", float64(successCount.Load())/duration.Seconds())
		t.Logf("平均延迟: %v", avgLatency)
		t.Logf("成功查询数: %d", successCount.Load())
		t.Logf("失败查询数: %d", errorCount.Load())
		t.Logf("CPU核心数: %d", runtime.NumCPU())
		t.Logf("Goroutine数量: %d", workers)

		// 如果有错误，输出错误信息
		if errorCount.Load() > 0 {
			t.Log("\n查询错误详情:")
			errorMsgs.Range(func(key, value interface{}) bool {
				t.Logf("%v: %v", key, value)
				return true
			})
		}

		totalProcessed := successCount.Load() + errorCount.Load()
		successRate := float64(successCount.Load()) / float64(queryCount) * 100

		t.Logf("查询成功率: %.2f%%", successRate)

		// 允许最多 2% 的查询偏差
		maxDeviation := int64(float64(queryCount) * 0.02)
		actualDeviation := int64(queryCount) - totalProcessed

		assert.LessOrEqual(t, actualDeviation, maxDeviation,
			"查询总数偏差(%d)不应超过允许的最大偏差(%d), 成功数:%d, 失败数:%d",
			actualDeviation, maxDeviation, successCount.Load(), errorCount.Load())
		assert.GreaterOrEqual(t, successRate, 98.0, "查询成功率应大于98%")
		assert.Less(t, avgLatency, 100*time.Millisecond, "平均查询延迟应小于100ms")
	})
}

func generateTestDocument(id int) map[string]interface{} {
	return map[string]interface{}{
		"id":        id,
		"timestamp": time.Now(),
		"message":   fmt.Sprintf("性能测试消息 %d", id),
		"level":     "info",
		"metadata": map[string]interface{}{
			"host":    fmt.Sprintf("host-%d", id%100),
			"service": fmt.Sprintf("service-%d", id%10),
			"region":  fmt.Sprintf("region-%d", id%5),
		},
		"metrics": map[string]interface{}{
			"cpu":    float64(id%100) / 100.0,
			"memory": float64(id%1024) * 1024 * 1024,
			"disk":   float64(id%500) * 1024 * 1024 * 1024,
		},
	}
}

func generateTestQuery(id int) map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{
						"range": map[string]interface{}{
							"timestamp": map[string]interface{}{
								"gte": "now-1h",
								"lt":  "now",
							},
						},
					},
					{
						"term": map[string]interface{}{
							"metadata.host": fmt.Sprintf("host-%d", id%100),
						},
					},
				},
			},
		},
		"sort": []map[string]interface{}{
			{
				"timestamp": map[string]interface{}{
					"order": "desc",
				},
			},
		},
	}
}
