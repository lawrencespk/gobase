package integration

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/logger/elk"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorHandling(t *testing.T) {
	// 创建一个带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Run("连接错误测试", func(t *testing.T) {
		// 配置一个错误的地址
		invalidConfig := &elk.ElkConfig{
			Addresses: []string{"http://invalid-host:9200"},
			Timeout:   5 * time.Second,
		}

		client := elk.NewElkClient()
		err := client.Connect(invalidConfig)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "lookup invalid-host")
	})

	t.Run("索引错误测试", func(t *testing.T) {
		client := setupTestClient(t)
		defer client.Close()

		// 尝试创建无效的索引映射
		invalidMapping := &elk.IndexMapping{
			Settings: map[string]interface{}{
				"invalid_setting": "invalid_value",
			},
		}

		err := client.CreateIndex(ctx, "test-invalid-index", invalidMapping)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "index creation")
	})

	t.Run("批量处理错误测试", func(t *testing.T) {
		client := setupTestClient(t)
		defer client.Close()

		// 创建一个配置错误的批处理器
		invalidConfig := &elk.BulkProcessorConfig{
			BatchSize:    0, // 无效的批处理大小
			FlushBytes:   1,
			Interval:     time.Millisecond,
			RetryCount:   1,
			RetryWait:    time.Millisecond,
			CloseTimeout: time.Second,
		}

		processor := elk.NewBulkProcessor(client, invalidConfig)
		assert.Nil(t, processor, "应该返回nil，因为配置无效")
	})

	t.Run("文档大小限制测试", func(t *testing.T) {
		client := setupTestClient(t)
		defer client.Close()

		// 创建一个有效的批处理器但设置很小的文档大小限制
		config := &elk.BulkProcessorConfig{
			BatchSize:    1,
			FlushBytes:   10, // 非常小的文档大小限制
			Interval:     time.Second,
			RetryCount:   1,
			RetryWait:    time.Millisecond,
			CloseTimeout: time.Second,
		}

		processor := elk.NewBulkProcessor(client, config)
		require.NotNil(t, processor)
		defer processor.Close()

		// 尝试添加一个大文档
		largeDoc := map[string]interface{}{
			"field": string(make([]byte, 100)), // 创建一个超过限制的文档
		}

		err := processor.Add(ctx, "test-index", largeDoc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "document size exceeds")
	})

	t.Run("重试机制测试", func(t *testing.T) {
		client := setupTestClient(t)
		defer client.Close()

		retryConfig := elk.RetryConfig{
			MaxRetries:  2,
			InitialWait: time.Millisecond,
			MaxWait:     time.Millisecond * 10,
		}

		operationCount := 0
		operation := func() error {
			operationCount++
			if operationCount <= retryConfig.MaxRetries {
				return errors.NewELKConnectionError("模拟连接错误", nil)
			}
			return nil
		}

		logger := logrus.New()
		err := elk.WithRetry(ctx, retryConfig, operation, logger)
		assert.NoError(t, err)
		assert.Equal(t, retryConfig.MaxRetries+1, operationCount)
	})

	t.Run("关闭超时测试", func(t *testing.T) {
		client := setupTestClient(t)
		defer client.Close()

		config := &elk.BulkProcessorConfig{
			BatchSize:    100,
			FlushBytes:   1024,
			Interval:     time.Second,
			RetryCount:   1,
			RetryWait:    time.Millisecond,
			CloseTimeout: time.Nanosecond, // 极短的超时时间
		}

		processor := elk.NewBulkProcessor(client, config)
		require.NotNil(t, processor)

		// 添加一些文档以确保有待处理的操作
		for i := 0; i < 10; i++ { // 减少文档数量，避免内存问题
			doc := map[string]interface{}{
				"field": i,
			}
			err := processor.Add(ctx, "test-index", doc)
			require.NoError(t, err)
		}

		err := processor.Close()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timed out")
	})
}

// setupTestClient 创建测试用的ELK客户端
func setupTestClient(t *testing.T) elk.Client {
	config := &elk.ElkConfig{
		Addresses: []string{"http://localhost:9200"},
		Timeout:   5 * time.Second,
	}

	client := elk.NewElkClient()
	err := client.Connect(config)
	require.NoError(t, err, "Failed to connect to Elasticsearch")

	return client
}
