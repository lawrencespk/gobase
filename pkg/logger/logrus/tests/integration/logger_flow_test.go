package integration

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gobase/pkg/logger"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
)

func TestCompleteLoggingFlow(t *testing.T) {
	// 创建一个buffer来捕获输出
	var buf bytes.Buffer

	// 创建临时目录
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	t.Logf("Using log path: %s", logPath)

	// 配置完整的日志系统
	opts := []logrus.Option{
		logrus.WithOutputPaths([]string{logPath}),
		logrus.WithOutput(&buf),
		logrus.WithAsyncConfig(logrus.AsyncConfig{
			Enable:        true,
			BufferSize:    1024,
			FlushInterval: time.Millisecond * 100,
			FlushOnExit:   true,
		}),
		logrus.WithCompressConfig(logrus.CompressConfig{
			Enable:       true,
			Algorithm:    "gzip",
			DeleteSource: true,
			Interval:     time.Second * 2,
		}),
		logrus.WithCleanupConfig(logrus.CleanupConfig{
			Enable:     true,
			MaxBackups: 3,
			MaxAge:     1,
			Interval:   time.Second,
			LogPaths:   []string{logPath},
		}),
	}

	// 创建日志实例
	log, err := logger.NewLogger(opts...)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 写入不同级别的日志
	ctx := context.Background()
	log.Info(ctx, "test info message", types.Field{Key: "key", Value: "value"})
	log.Error(ctx, "test error message", types.Field{Key: "error", Value: errors.New("test error")})

	// 等待异步操作完成并确保日志被刷新
	time.Sleep(time.Second)
	if closer, ok := log.(interface{ Sync() error }); ok {
		if err := closer.Sync(); err != nil {
			t.Logf("Warning: Failed to sync logger: %v", err)
		}
	}

	// 关闭日志实例以释放文件句柄
	if closer, ok := log.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			t.Logf("Warning: Failed to close logger: %v", err)
		}
	}

	// 验证缓冲区中的日志内容
	output := buf.String()
	t.Logf("Buffer content length: %d", len(output))
	if !bytes.Contains([]byte(output), []byte("test info message")) {
		t.Error("Output should contain info message")
		t.Logf("Actual output: %s", output)
	}
	if !bytes.Contains([]byte(output), []byte("test error message")) {
		t.Error("Output should contain error message")
		t.Logf("Actual output: %s", output)
	}

	// 等待压缩完成
	t.Log("Waiting for compression...")
	time.Sleep(time.Second * 3)

	// 验证压缩功能
	compressedPath := logPath + ".gz"
	if _, err := os.Stat(compressedPath); os.IsNotExist(err) {
		t.Error("Compressed log file should exist")
		// 列出目录内容
		if files, err := os.ReadDir(filepath.Dir(logPath)); err == nil {
			t.Log("Directory contents:")
			for _, file := range files {
				t.Logf("- %s", file.Name())
			}
		}
		// 检查原始文件是否存在
		if _, err := os.Stat(logPath); err == nil {
			t.Log("Original log file still exists")
			// 获取文件大小
			if info, err := os.Stat(logPath); err == nil {
				t.Logf("Original file size: %d bytes", info.Size())
			}
		} else {
			t.Log("Original log file does not exist")
		}
	} else {
		t.Log("Compressed file exists")
		// 获取压缩文件大小
		if info, err := os.Stat(compressedPath); err == nil {
			t.Logf("Compressed file size: %d bytes", info.Size())
		}
	}
}
