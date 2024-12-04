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
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	t.Logf("Using temporary directory: %s", tmpDir)
	t.Logf("Log file path: %s", logPath)

	opts := []logrus.Option{
		logrus.WithOutputPaths([]string{logPath}),
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

	// 修改目录权限前检查
	t.Log("Checking initial directory permissions...")
	if info, err := os.Stat(tmpDir); err == nil {
		t.Logf("Initial directory permissions: %v", info.Mode())
	}

	if err := os.Chmod(tmpDir, 0000); err != nil {
		t.Fatalf("Failed to change directory permissions: %v", err)
	}

	t.Log("Writing test message...")
	ctx := context.Background()
	log.Info(ctx, "test message")

	t.Log("Restoring directory permissions...")
	if err := os.Chmod(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to restore directory permissions: %v", err)
	}

	t.Log("Closing logger...")
	if closer, ok := log.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			t.Logf("Warning: error while closing logger: %v", err)
		}
	}

	t.Log("Checking for log file...")
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		// 列出目录内容以帮助调试
		files, _ := os.ReadDir(tmpDir)
		t.Log("Directory contents:")
		for _, file := range files {
			t.Logf("- %s", file.Name())
		}
		t.Error("Log file should exist after recovery")
	} else if err != nil {
		t.Errorf("Error checking log file: %v", err)
	} else {
		t.Log("Log file exists")
	}
}
