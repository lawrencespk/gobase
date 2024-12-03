package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gobase/pkg/logger"
	"gobase/pkg/logger/logrus"
)

func TestLoggerRecovery(t *testing.T) {
	// 模拟磁盘写满的情况
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	opts := []logrus.Option{
		logrus.WithRecoveryConfig(logrus.RecoveryConfig{
			Enable:        true,
			MaxRetries:    3,
			RetryInterval: time.Millisecond * 100,
		}),
	}

	log, err := logger.NewLogger(opts...)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 模拟文件系统故障
	if err := os.Chmod(tmpDir, 0000); err != nil {
		t.Fatalf("Failed to change directory permissions: %v", err)
	}

	// 尝试写入日志
	ctx := context.Background()
	log.Info(ctx, "test message")

	// 恢复文件系统
	if err := os.Chmod(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to restore directory permissions: %v", err)
	}

	// 等待重试完成
	time.Sleep(time.Second)

	// 验证日志是否最终写入
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Error("Log file should exist after recovery")
	}
}
