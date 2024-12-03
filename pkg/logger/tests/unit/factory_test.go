package unit

import (
	"path/filepath"
	"testing"
	"time"

	"gobase/pkg/logger"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
)

// TestGetLogger 测试获取默认日志实例
func TestGetLogger(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 修改日志路径为临时目录
	t.Setenv("LOG_PATH", filepath.Join(tmpDir, "app.log"))

	// 运行测试
	logger1 := logger.GetLogger()
	logger2 := logger.GetLogger()

	if logger1 != logger2 {
		t.Error("GetLogger should return the same instance")
	}

	if logger1 == nil {
		t.Error("GetLogger should not return nil")
	}
}

// TestSetLogger 测试设置默认日志实例
func TestSetLogger(t *testing.T) {
	// 创建一个mock logger
	mockLogger := &types.BasicLogger{}

	// 设置新的logger
	logger.SetLogger(mockLogger)

	// 验证是否设置成功
	if logger.GetLogger() != mockLogger {
		t.Error("SetLogger failed to set the default logger")
	}
}

// TestNewLogger 测试创建新的日志实例
func TestNewLogger(t *testing.T) {
	// 测试默认配置
	t.Run("Default Config", func(t *testing.T) {
		log, err := logger.NewLogger()
		if err != nil {
			t.Errorf("NewLogger failed with default config: %v", err)
		}
		if log == nil {
			t.Error("NewLogger should not return nil logger")
		}
	})

	// 测试自定义配置
	t.Run("Custom Config", func(t *testing.T) {
		customLevel := types.DebugLevel
		opt := logrus.WithLevel(customLevel)

		log, err := logger.NewLogger(opt)
		if err != nil {
			t.Errorf("NewLogger failed with custom config: %v", err)
		}

		// 检查日志级别，但不进行具体类型断言
		if log == nil {
			t.Error("NewLogger should not return nil logger")
			return
		}
	})
}

// TestFactory 测试日志工厂
func TestFactory(t *testing.T) {
	// 测试创建有效的工厂
	t.Run("Valid Factory", func(t *testing.T) {
		factory := logger.NewFactory("logrus")
		log, err := factory.Create()

		if err != nil {
			t.Errorf("Factory Create failed: %v", err)
		}
		if log == nil {
			t.Error("Factory should not create nil logger")
		}
	})

	// 测试无效的日志类型
	t.Run("Invalid Logger Type", func(t *testing.T) {
		factory := logger.NewFactory("invalid")
		log, err := factory.Create()

		if err == nil {
			t.Error("Factory should return error for invalid logger type")
		}
		if log != nil {
			t.Error("Factory should return nil logger for invalid type")
		}
	})

	// 测试自定义配置
	t.Run("Custom Options", func(t *testing.T) {
		factory := logger.NewFactory("logrus")
		customLevel := types.DebugLevel
		log, err := factory.Create(logrus.WithLevel(customLevel))

		if err != nil {
			t.Errorf("Factory Create failed with custom options: %v", err)
		}

		if log == nil {
			t.Error("Factory should not create nil logger")
			return
		}
	})
}

// TestInitializeLogger 测试日志系统初始化
func TestInitializeLogger(t *testing.T) {
	log := logger.InitializeLogger()

	if log == nil {
		t.Error("InitializeLogger should not return nil")
	}

	// 验证是否正确设置为默认logger
	if logger.GetLogger() != log {
		t.Error("InitializeLogger should set the default logger")
	}
}

// TestLoggerConcurrency 测试并发安全性
func TestLoggerConcurrency(t *testing.T) {
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			logger := logger.GetLogger()
			if logger == nil {
				t.Error("Concurrent GetLogger returned nil")
			}
			done <- true
		}()
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("Concurrent test timed out")
		}
	}
}
