package benchmark_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
)

func BenchmarkLogger(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "logger_bench_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	ctx := context.Background()

	// 基本同步测试
	b.Run("Sync Logger", func(b *testing.B) {
		logFile := filepath.Join(tempDir, "sync.log")

		// 创建 FileManager
		fm := logrus.NewFileManager(logrus.FileOptions{
			BufferSize:    32 * 1024,
			FlushInterval: time.Millisecond * 100,
			MaxOpenFiles:  100,
			DefaultPath:   logFile,
		})

		// 创建 QueueConfig
		queueConfig := logrus.QueueConfig{
			MaxSize: 1024,
			Workers: 4,
		}

		// 创建 Options
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{logFile},
			CompressConfig: logrus.CompressConfig{
				Enable: false,
			},
		}

		logger, err := logrus.NewLogger(fm, queueConfig, opts)
		if err != nil {
			b.Fatalf("Failed to create logger: %v", err)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(ctx, "benchmark test message")
			}
		})
	})

	// 异步测试
	b.Run("Async Logger", func(b *testing.B) {
		logFile := filepath.Join(tempDir, "async.log")

		fm := logrus.NewFileManager(logrus.FileOptions{
			BufferSize:    32 * 1024,
			FlushInterval: time.Millisecond * 100,
			MaxOpenFiles:  100,
			DefaultPath:   logFile,
		})

		queueConfig := logrus.QueueConfig{
			MaxSize: 1024,
			Workers: 4,
		}

		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{logFile},
			AsyncConfig: logrus.AsyncConfig{
				Enable:        true,
				BufferSize:    1024 * 1024,
				FlushInterval: time.Millisecond * 100,
				DropOnFull:    true,
			},
			CompressConfig: logrus.CompressConfig{
				Enable: false,
			},
		}

		logger, err := logrus.NewLogger(fm, queueConfig, opts)
		if err != nil {
			b.Fatalf("Failed to create logger: %v", err)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.Info(ctx, "benchmark test message")
			}
		})
	})

	// 消息大小测试
	b.Run("Message Size Impact", func(b *testing.B) {
		sizes := []int{64, 256, 512, 1024}
		for _, size := range sizes {
			message := strings.Repeat("a", size)
			b.Run(fmt.Sprintf("size_%d", size), func(b *testing.B) {
				logFile := filepath.Join(tempDir, fmt.Sprintf("msgsize_%d.log", size))

				fm := logrus.NewFileManager(logrus.FileOptions{
					BufferSize:    32 * 1024,
					FlushInterval: time.Millisecond * 100,
					MaxOpenFiles:  100,
					DefaultPath:   logFile,
				})

				queueConfig := logrus.QueueConfig{
					MaxSize: 1024,
					Workers: 4,
				}

				opts := &logrus.Options{
					Level:       types.InfoLevel,
					OutputPaths: []string{logFile},
					AsyncConfig: logrus.AsyncConfig{
						Enable:        true,
						BufferSize:    1024 * 1024,
						FlushInterval: time.Millisecond * 100,
					},
				}

				logger, err := logrus.NewLogger(fm, queueConfig, opts)
				if err != nil {
					b.Fatalf("Failed to create logger: %v", err)
				}

				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						logger.Info(ctx, message)
					}
				})
			})
		}
	})

	// 其他测试用例也需要类似的修改...
}
