package integration

import (
	"context"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/elk"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestRecoveryMechanism(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 设置测试环境
	ctx := context.Background()
	logger := logrus.New()

	// 创建ELK配置
	elkConfig := elk.DefaultElkConfig()
	elkConfig.Addresses = []string{"http://localhost:9200"}
	elkConfig.Index = "test-recovery-logs"
	elkConfig.Timeout = 5 * time.Second // 增加超时时间

	// 创建一个ES客户端并确保连接
	client := elk.NewElkClient()
	err := client.Connect(elkConfig)
	if err != nil {
		t.Skipf("Elasticsearch 未启动或无法连接: %v", err)
		return
	}
	defer client.Close()

	// 创建Hook配置
	hookOpts := elk.ElkHookOptions{
		Config: elkConfig,
		Levels: []logrus.Level{logrus.InfoLevel, logrus.ErrorLevel},
		Index:  "test-recovery-logs",
		BatchConfig: &elk.BulkProcessorConfig{
			BatchSize:    10,
			FlushBytes:   1024,
			Interval:     100 * time.Millisecond,
			RetryCount:   3,
			RetryWait:    100 * time.Millisecond,
			CloseTimeout: 1 * time.Second,
		},
		ErrorLogger: logger,
	}

	// 创建Hook
	hook, err := elk.NewElkHook(hookOpts)
	if err != nil {
		t.Fatalf("Failed to create ELK hook: %v", err)
	}
	defer hook.Close()

	// 创建重试配置
	retryConfig := elk.RetryConfig{
		MaxRetries:  3,
		InitialWait: 100 * time.Millisecond,
		MaxWait:     1 * time.Second,
	}

	// 模拟失败计数器
	var failureCount int32

	// 创建一个会失败的操作
	operation := func() error {
		if atomic.LoadInt32(&failureCount) < 2 {
			atomic.AddInt32(&failureCount, 1)
			return errors.NewError(codes.ELKBulkError, "模拟操作失败", nil)
		}
		return nil
	}

	// 测试重试机制
	t.Run("TestRetryMechanism", func(t *testing.T) {
		err := elk.WithRetry(ctx, retryConfig, operation, logger)
		assert.NoError(t, err)
		assert.Equal(t, int32(2), atomic.LoadInt32(&failureCount))
	})

	// 测试批量处理恢复
	t.Run("TestBulkProcessorRecovery", func(t *testing.T) {
		// 确保索引存在
		exists, err := client.IndexExists(ctx, hookOpts.Index)
		assert.NoError(t, err)

		if !exists {
			mapping := elk.DefaultIndexMapping()
			err = client.CreateIndex(ctx, hookOpts.Index, mapping)
			assert.NoError(t, err)
		}

		// 发送一些测试文档
		for i := 0; i < 5; i++ {
			entry := &logrus.Entry{
				Message: "test message " + strconv.Itoa(i),
				Time:    time.Now(),
				Level:   logrus.InfoLevel,
			}
			err := hook.Fire(entry)
			assert.NoError(t, err)
		}

		// 等待处理完成
		time.Sleep(500 * time.Millisecond)

		// 模拟系统崩溃
		hook.Close()

		// 创建新的Hook模拟重启
		newHook, err := elk.NewElkHook(hookOpts)
		assert.NoError(t, err)
		defer newHook.Close()

		// 发送更多文档
		for i := 5; i < 10; i++ {
			entry := &logrus.Entry{
				Message: "test message " + strconv.Itoa(i),
				Time:    time.Now(),
				Level:   logrus.InfoLevel,
			}
			err := newHook.Fire(entry)
			assert.NoError(t, err)
		}

		// 等待处理完成
		time.Sleep(500 * time.Millisecond)

		// 刷新索引以确保所有文档都可搜索
		err = client.RefreshIndex(ctx, hookOpts.Index)
		assert.NoError(t, err)

		// 获取索引统计信息
		stats, err := client.GetIndexStats(ctx, hookOpts.Index)
		assert.NoError(t, err)
		assert.NotNil(t, stats)
	})

	// 测试错误恢复
	t.Run("TestErrorRecovery", func(t *testing.T) {
		// 为错误恢复测试创建新的 Hook
		errorHook, err := elk.NewElkHook(hookOpts)
		assert.NoError(t, err)
		defer errorHook.Close()

		// 创建测试文档
		entry := &logrus.Entry{
			Message: "test recovery message",
			Time:    time.Now(),
			Level:   logrus.InfoLevel,
		}

		// 使用新Hook的Fire方法直接测试
		err = errorHook.Fire(entry)
		assert.NoError(t, err)

		// 模拟错误情况
		errorEntry := &logrus.Entry{
			Message: "test error recovery message",
			Time:    time.Now(),
			Level:   logrus.ErrorLevel,
			Data: logrus.Fields{
				"error": errors.NewError(codes.ELKBulkError, "test error", nil),
			},
		}

		// 测试错误恢复
		err = errorHook.Fire(errorEntry)
		assert.NoError(t, err)

		// 等待处理完成
		time.Sleep(500 * time.Millisecond)

		// 确保所有数据都已写入
		err = errorHook.Close()
		assert.NoError(t, err)

		// 刷新索引以确保所有文档都可搜索
		err = client.RefreshIndex(ctx, hookOpts.Index)
		assert.NoError(t, err)

		// 获取索引统计信息
		stats, err := client.GetIndexStats(ctx, hookOpts.Index)
		assert.NoError(t, err)
		assert.NotNil(t, stats)

		// 验证文档是否写入成功
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"message": "test error recovery message",
				},
			},
		}
		result, err := client.Query(ctx, hookOpts.Index, query)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}
