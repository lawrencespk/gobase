package testutils

import (
	"testing"
	"time"

	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
)

// NewTestLogger 创建用于测试的日志实例
func NewTestLogger(t *testing.T) types.Logger {
	// 创建基本的日志配置
	opts := &logrus.Options{
		Level:        types.DebugLevel,
		OutputPaths:  []string{"stdout"},
		ReportCaller: true,
	}

	// 创建文件管理器配置
	fileOpts := logrus.FileOptions{
		BufferSize:    32 * 1024,       // 32KB
		FlushInterval: time.Second,     // 1秒刷新
		MaxOpenFiles:  100,             // 最大打开文件数
		DefaultPath:   "logs/test.log", // 测试日志路径
	}

	// 创建文件管理器
	fm := logrus.NewFileManager(fileOpts)

	// 创建队列配置
	queueConfig := logrus.QueueConfig{
		MaxSize:         1000,            // 最大队列大小
		BatchSize:       100,             // 批处理大小
		FlushInterval:   time.Second,     // 刷新间隔
		Workers:         1,               // 工作协程数量
		RetryCount:      3,               // 重试次数
		RetryInterval:   time.Second,     // 重试间隔
		MaxBatchWait:    time.Second,     // 最大批处理等待时间
		ShutdownTimeout: time.Second * 5, // 关闭超时时间
	}

	// 创建日志实例
	logger, err := logrus.NewLogger(fm, queueConfig, opts)
	if err != nil {
		t.Fatalf("创建测试日志实例失败: %v", err)
	}

	return logger
}
