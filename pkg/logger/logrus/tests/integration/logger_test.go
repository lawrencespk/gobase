package integration

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"

	"github.com/stretchr/testify/require"
)

// TestLoggerIntegration 测试日志的完整工作流程
func TestLoggerIntegration(t *testing.T) {
	// 创建临时目录
	tempDir, err := os.MkdirTemp("", "logger_integration_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("Complete Logging Workflow", func(t *testing.T) {
		logFile := filepath.Join(tempDir, "test.log")
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

		// 创建 Options 实例
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{logFile},
			AsyncConfig: logrus.AsyncConfig{
				Enable:        true,
				BufferSize:    1024,
				FlushInterval: time.Millisecond * 100,
				DropOnFull:    true,
				FlushOnExit:   true,
			},
			CompressConfig: logrus.CompressConfig{
				Enable:       true,
				Algorithm:    "gzip",
				Level:        6,
				Interval:     time.Second,
				DeleteSource: false,
			},
			CleanupConfig: logrus.CleanupConfig{
				Enable:     true,
				MaxAge:     7,
				MaxBackups: 3,
				Interval:   24 * time.Hour,
			},
		}

		// 调用 NewLogger
		logger, err := logrus.NewLogger(fm, queueConfig, opts)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		ctx := context.Background()
		logger.Debug(ctx, "Debug message")
		logger.Info(ctx, "Info message")
		logger.Warn(ctx, "Warn message")
		logger.Error(ctx, "Error message")

		// 增加等待时间，确保异步写入和压缩完成
		time.Sleep(3 * time.Second)

		// 检查日志文件内容
		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		expectedMessages := []string{
			"Info message",
			"Warn message",
			"Error message",
		}

		for _, msg := range expectedMessages {
			if !bytes.Contains(content, []byte(msg)) {
				t.Errorf("Expected log to contain %q", msg)
			}
		}

		if bytes.Contains(content, []byte("Debug message")) {
			t.Error("Debug message should not be logged at info level")
		}
	})

	t.Run("Concurrent Logging", func(t *testing.T) {
		logFile := filepath.Join(tempDir, "concurrent.log")
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

		// 创建 Options 实例
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{logFile},
			AsyncConfig: logrus.AsyncConfig{
				Enable:        true,
				BufferSize:    1024,
				FlushInterval: time.Millisecond * 100,
				DropOnFull:    false,
			},
		}

		// 调用 NewLogger
		logger, err := logrus.NewLogger(fm, queueConfig, opts)
		require.NoError(t, err)

		var wg sync.WaitGroup
		numWorkers := 10
		messagesPerWorker := 100
		totalMessages := numWorkers * messagesPerWorker
		ctx := context.Background()

		wg.Add(totalMessages)
		for i := 0; i < numWorkers; i++ {
			go func(workerID int) {
				for j := 0; j < messagesPerWorker; j++ {
					logger.Info(ctx, fmt.Sprintf("Worker %d - Message %d", workerID, j))
					wg.Done()
				}
			}(i)
		}

		wg.Wait()
		if err := logger.Sync(); err != nil {
			t.Fatalf("Failed to sync logger: %v", err)
		}

		time.Sleep(time.Second * 2)

		content, err := os.ReadFile(logFile)
		require.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(string(content)), "\n")
		expectedLines := numWorkers * messagesPerWorker

		if len(lines) != expectedLines {
			t.Errorf("Expected %d log lines, got %d", expectedLines, len(lines))
			t.Logf("First few lines: %v", lines[:min(10, len(lines))])
			t.Logf("Last few lines: %v", lines[max(0, len(lines)-10):])
		}
	})

	t.Run("Error Recovery", func(t *testing.T) {
		// 创建一个无法写入的目录
		readOnlyDir := filepath.Join(tempDir, "readonly")
		if err := os.Mkdir(readOnlyDir, 0500); err != nil {
			t.Fatalf("Failed to create readonly dir: %v", err)
		}

		logFile := filepath.Join(readOnlyDir, "test.log")
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

		// 创建 Options 实例
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{logFile},
			RecoveryConfig: logrus.RecoveryConfig{
				Enable:        true,
				MaxRetries:    3,
				RetryInterval: time.Millisecond * 100,
			},
			CompressConfig: logrus.CompressConfig{
				Enable:    true,
				Algorithm: "gzip",
				Level:     6,
				Interval:  time.Second,
			},
		}

		// 调用 NewLogger
		logger, err := logrus.NewLogger(fm, queueConfig, opts)
		if err != nil {
			t.Fatalf("Failed to create logger: %v", err)
		}

		ctx := context.Background()
		// 尝试写入日志
		logger.Info(ctx, "Test message")

		// 等待重试完成
		time.Sleep(time.Second)
	})

	t.Run("Log Rotation", func(t *testing.T) {
		logFile := filepath.Join(tempDir, "rotation.log")
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

		// 创建 Options 实例
		opts := &logrus.Options{
			Level:       types.InfoLevel,
			OutputPaths: []string{logFile},
			CleanupConfig: logrus.CleanupConfig{
				Enable:     true,
				MaxAge:     1, // 1天
				MaxBackups: 3,
				Interval:   24 * time.Hour,
			},
		}

		// 调用 NewLogger
		logger, err := logrus.NewLogger(fm, queueConfig, opts)
		require.NoError(t, err)

		ctx := context.Background()
		// 写入足够多的日志以触发轮转
		for i := 0; i < 100000; i++ {
			logger.Info(ctx, fmt.Sprintf("Log message %d with some padding to increase file size...", i))
		}

		// 等待文件系统操作完成
		time.Sleep(time.Second)

		// 检查是否创建了备份文件
		files, err := filepath.Glob(logFile + ".*")
		require.NoError(t, err)
		if len(files) == 0 {
			t.Error("Expected backup files to be created")
			// 输出目录内容以帮助调试
			dirEntries, _ := os.ReadDir(tempDir)
			for _, entry := range dirEntries {
				info, _ := entry.Info()
				t.Logf("File in temp dir: %s (size: %d)", entry.Name(), info.Size())
			}
		}
	})
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
