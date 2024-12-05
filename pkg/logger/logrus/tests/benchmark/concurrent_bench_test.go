package benchmark

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func BenchmarkConcurrentLogging(b *testing.B) {
	// 设置基准测试场景
	scenarios := []struct {
		name      string
		workers   int
		batchSize int
	}{
		{"Small_Concurrency", 4, 100},
		{"Medium_Concurrency", runtime.NumCPU(), 100},
		{"High_Concurrency", runtime.NumCPU() * 2, 100},
		{"Very_High_Concurrency", runtime.NumCPU() * 4, 100},
	}

	for _, scenario := range scenarios {
		b.Run(fmt.Sprintf("%s/Workers_%d/BatchSize_%d", scenario.name, scenario.workers, scenario.batchSize), func(b *testing.B) {
			logger := logrus.New()
			logger.SetLevel(logrus.InfoLevel)
			logger.SetOutput(new(syncBuffer))

			var (
				wg           sync.WaitGroup
				successCount int64
				errorCount   int64
				totalLatency int64
			)

			// 确保每个worker至少处理一个批次
			logsPerWorker := b.N / scenario.workers
			if logsPerWorker == 0 {
				logsPerWorker = 1
			}

			b.ResetTimer()

			// 启动多个worker并发写入日志
			for i := 0; i < scenario.workers; i++ {
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()

					for j := 0; j < logsPerWorker; j++ {
						start := time.Now()

						// 模拟实际日志写入场景
						logger.WithFields(logrus.Fields{
							"worker_id":  workerID,
							"batch_id":   j,
							"timestamp":  time.Now().UnixNano(),
							"request_id": fmt.Sprintf("req-%d-%d", workerID, j),
							"metadata": map[string]interface{}{
								"service": "test-service",
								"version": "1.0.0",
								"env":     "benchmark",
							},
						}).Info("concurrent logging benchmark test message")

						latency := time.Since(start).Microseconds()
						atomic.AddInt64(&totalLatency, latency)
						atomic.AddInt64(&successCount, 1)
					}
				}(i)
			}

			wg.Wait()
			b.StopTimer()

			// 计算并报告性能指标
			totalOps := atomic.LoadInt64(&successCount)
			if totalOps > 0 {
				avgLatency := time.Duration(atomic.LoadInt64(&totalLatency)/totalOps) * time.Microsecond
				errorRate := float64(atomic.LoadInt64(&errorCount)) / float64(totalOps) * 100

				b.ReportMetric(float64(avgLatency.Nanoseconds()), "ns/op")
				b.ReportMetric(float64(errorRate), "error_rate")
				b.ReportMetric(float64(totalOps), "total_ops")
				b.ReportMetric(float64(runtime.NumGoroutine()), "goroutines")
			}
		})
	}
}

// syncBuffer 是一个线程安全的buffer实现
type syncBuffer struct {
	mu  sync.Mutex
	buf []byte
}

func (b *syncBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buf = append(b.buf, p...)
	return len(p), nil
}

// BenchmarkConcurrentLoggingWithFields 测试带有不同数量字段的并发日志写入性能
func BenchmarkConcurrentLoggingWithFields(b *testing.B) {
	fieldSets := []struct {
		name   string
		fields logrus.Fields
	}{
		{
			name: "Basic_Fields",
			fields: logrus.Fields{
				"service": "test",
				"version": "1.0",
			},
		},
		{
			name: "Medium_Fields",
			fields: logrus.Fields{
				"service":    "test",
				"version":    "1.0",
				"request_id": "req-123",
				"user_id":    "user-456",
				"ip":         "192.168.1.1",
			},
		},
		{
			name: "Many_Fields",
			fields: logrus.Fields{
				"service":     "test",
				"version":     "1.0",
				"request_id":  "req-123",
				"user_id":     "user-456",
				"ip":          "192.168.1.1",
				"method":      "POST",
				"path":        "/api/v1/test",
				"status":      200,
				"latency_ms":  150,
				"user_agent":  "Mozilla/5.0",
				"referer":     "https://example.com",
				"trace_id":    "trace-789",
				"environment": "production",
			},
		},
	}

	workers := runtime.NumCPU()

	for _, fieldSet := range fieldSets {
		b.Run(fmt.Sprintf("%s/Workers_%d", fieldSet.name, workers), func(b *testing.B) {
			logger := logrus.New()
			logger.SetLevel(logrus.InfoLevel)
			logger.SetOutput(new(syncBuffer))

			var wg sync.WaitGroup
			var totalLatency int64
			var totalOps int64

			// 确保每个worker至少处理一个日志
			logsPerWorker := b.N / workers
			if logsPerWorker == 0 {
				logsPerWorker = 1
			}

			b.ResetTimer()

			for i := 0; i < workers; i++ {
				wg.Add(1)
				go func(workerID int) {
					defer wg.Done()

					localFields := logrus.Fields{}
					for k, v := range fieldSet.fields {
						localFields[k] = v
					}
					localFields["worker_id"] = workerID

					for j := 0; j < logsPerWorker; j++ {
						start := time.Now()
						logger.WithFields(localFields).Info("benchmark test message")
						latency := time.Since(start).Microseconds()
						atomic.AddInt64(&totalLatency, latency)
						atomic.AddInt64(&totalOps, 1)
					}
				}(i)
			}

			wg.Wait()
			b.StopTimer()

			if ops := atomic.LoadInt64(&totalOps); ops > 0 {
				avgLatency := time.Duration(atomic.LoadInt64(&totalLatency)/ops) * time.Microsecond
				b.ReportMetric(float64(avgLatency.Nanoseconds()), "ns/op")
				b.ReportMetric(float64(ops), "total_ops")
			}
		})
	}
}
