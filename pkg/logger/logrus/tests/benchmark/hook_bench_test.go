package benchmark

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"gobase/pkg/logger/elk"

	"github.com/sirupsen/logrus"
)

// 模拟的 Hook 实现，用于基准测试
type benchmarkHook struct {
	mu     sync.Mutex
	buffer []string
}

func newBenchmarkHook() *benchmarkHook {
	return &benchmarkHook{
		buffer: make([]string, 0, 1000),
	}
}

func (h *benchmarkHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *benchmarkHook) Fire(entry *logrus.Entry) error {
	h.mu.Lock()
	h.buffer = append(h.buffer, entry.Message)
	h.mu.Unlock()
	return nil
}

func BenchmarkHook(b *testing.B) {
	b.Run("SimpleHook", func(b *testing.B) {
		logger := logrus.New()
		hook := newBenchmarkHook()
		logger.AddHook(hook)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.WithFields(logrus.Fields{
				"count": i,
				"time":  time.Now(),
			}).Info("benchmark test message")
		}
	})

	b.Run("ConcurrentHook", func(b *testing.B) {
		logger := logrus.New()
		hook := newBenchmarkHook()
		logger.AddHook(hook)

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			count := 0
			for pb.Next() {
				logger.WithFields(logrus.Fields{
					"count": count,
					"time":  time.Now(),
				}).Info("benchmark test message")
				count++
			}
		})
	})
}

func BenchmarkELKHook(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping ELK hook benchmark in short mode")
	}

	// 创建 ELK 客户端
	config := &elk.ElkConfig{
		Addresses: []string{"http://localhost:9200"},
		Username:  "elastic",
		Password:  "changeme",
		Timeout:   30,
	}

	client := elk.NewElkClient()
	if err := client.Connect(config); err != nil {
		b.Fatal("Failed to connect to Elasticsearch:", err)
	}
	defer client.Close()

	// 创建测试索引
	testIndex := fmt.Sprintf("benchmark-logs-%d", time.Now().Unix())
	ctx := context.Background()
	if err := client.CreateIndex(ctx, testIndex, elk.DefaultIndexMapping()); err != nil {
		b.Fatal("Failed to create index:", err)
	}
	defer client.DeleteIndex(ctx, testIndex)

	// 创建 ELK Hook
	hook, err := elk.NewElkHook(elk.ElkHookOptions{
		Config: config,
		Index:  testIndex,
		BatchConfig: &elk.BulkProcessorConfig{
			BatchSize:    1000,
			FlushBytes:   5 * 1024 * 1024,
			Interval:     time.Second,
			RetryCount:   3,
			RetryWait:    time.Second,
			CloseTimeout: 30 * time.Second,
		},
		ErrorLogger: logrus.StandardLogger(),
	})
	if err != nil {
		b.Fatal("Failed to create ELK hook:", err)
	}
	defer hook.Close()

	logger := logrus.New()
	logger.AddHook(hook)

	b.Run("SingleELKHook", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.WithFields(logrus.Fields{
				"benchmark_id": i,
				"timestamp":    time.Now(),
				"level":        "info",
				"component":    "benchmark",
				"metadata": map[string]interface{}{
					"host":    fmt.Sprintf("host-%d", i%100),
					"service": "benchmark-service",
					"region":  fmt.Sprintf("region-%d", i%5),
				},
			}).Info("ELK hook benchmark message")
		}
	})

	b.Run("ConcurrentELKHook", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			count := 0
			for pb.Next() {
				logger.WithFields(logrus.Fields{
					"benchmark_id": count,
					"timestamp":    time.Now(),
					"level":        "info",
					"component":    "benchmark",
					"metadata": map[string]interface{}{
						"host":    fmt.Sprintf("host-%d", count%100),
						"service": "benchmark-service",
						"region":  fmt.Sprintf("region-%d", count%5),
					},
				}).Info("ELK hook concurrent benchmark message")
				count++
			}
		})
	})

	b.Run("BulkELKHook", func(b *testing.B) {
		// 设置较大的批量大小以测试批处理性能
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			count := 0
			for pb.Next() {
				logger.WithFields(logrus.Fields{
					"benchmark_id": count,
					"timestamp":    time.Now(),
					"level":        "info",
					"component":    "benchmark",
					"metadata": map[string]interface{}{
						"host":    fmt.Sprintf("host-%d", count%100),
						"service": "benchmark-service",
						"region":  fmt.Sprintf("region-%d", count%5),
					},
					"metrics": map[string]interface{}{
						"cpu":    float64(count%100) / 100.0,
						"memory": float64(count%1024) * 1024 * 1024,
						"disk":   float64(count%500) * 1024 * 1024 * 1024,
					},
				}).Info("ELK hook bulk benchmark message")
				count++
			}
		})
	})

	b.Run("HighLoadELKHook", func(b *testing.B) {
		// 模拟高负载情况
		logger.SetLevel(logrus.DebugLevel)
		levels := []logrus.Level{
			logrus.DebugLevel,
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.ErrorLevel,
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			count := 0
			for pb.Next() {
				level := levels[count%len(levels)]
				entry := logger.WithFields(logrus.Fields{
					"benchmark_id": count,
					"timestamp":    time.Now(),
					"level":        level.String(),
					"component":    "benchmark",
					"metadata": map[string]interface{}{
						"host":    fmt.Sprintf("host-%d", count%100),
						"service": "benchmark-service",
						"region":  fmt.Sprintf("region-%d", count%5),
					},
					"metrics": map[string]interface{}{
						"cpu":    float64(count%100) / 100.0,
						"memory": float64(count%1024) * 1024 * 1024,
						"disk":   float64(count%500) * 1024 * 1024 * 1024,
					},
				})

				switch level {
				case logrus.DebugLevel:
					entry.Debug("Debug level benchmark message")
				case logrus.InfoLevel:
					entry.Info("Info level benchmark message")
				case logrus.WarnLevel:
					entry.Warn("Warn level benchmark message")
				case logrus.ErrorLevel:
					entry.Error("Error level benchmark message")
				}
				count++
			}
		})
	})
}
