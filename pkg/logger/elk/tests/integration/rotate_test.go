package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gobase/pkg/logger/elk"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexRotation(t *testing.T) {
	// 设置测试配置
	config := getElkConfig()

	// 创建客户端
	client := elk.NewElkClient()
	err := client.Connect(config)
	require.NoError(t, err)
	defer client.Close()

	// 创建测试上下文
	ctx := context.Background()

	// 测试索引前缀
	indexPrefix := "test-logs"

	// 测试用例
	testCases := []struct {
		name     string
		docs     int
		interval time.Duration
	}{
		{
			name:     "基本轮转测试",
			docs:     100,
			interval: 1 * time.Second,
		},
		{
			name:     "高频轮转测试",
			docs:     1000,
			interval: 100 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 创建Hook配置
			hookOpts := elk.ElkHookOptions{
				Config: config,
				Index:  fmt.Sprintf("%s-%s", indexPrefix, time.Now().Format("2006.01.02")),
				BatchConfig: &elk.BulkProcessorConfig{
					BatchSize:    50,
					FlushBytes:   1 * 1024 * 1024,
					Interval:     tc.interval,
					RetryCount:   3,
					RetryWait:    time.Second,
					CloseTimeout: 5 * time.Second,
				},
			}

			// 创建Hook
			hook, err := elk.NewElkHook(hookOpts)
			require.NoError(t, err)
			defer hook.Close()

			// 写入测试数据
			for i := 0; i < tc.docs; i++ {
				doc := map[string]interface{}{
					"message":   fmt.Sprintf("test message %d", i),
					"timestamp": time.Now(),
					"level":     "info",
				}
				err := hook.GetBulkProcessor().Add(ctx, hookOpts.Index, doc)
				require.NoError(t, err)

				if i > 0 && i%100 == 0 {
					time.Sleep(tc.interval)
				}
			}

			// 等待数据写入完成
			time.Sleep(2 * tc.interval)

			// 验证索引状态
			stats, err := client.GetIndexStats(ctx, hookOpts.Index)
			require.NoError(t, err)

			// 验证文档数量
			docCount, ok := stats["_all"].(map[string]interface{})["primaries"].(map[string]interface{})["docs"].(map[string]interface{})["count"].(float64)
			require.True(t, ok)
			assert.Equal(t, float64(tc.docs), docCount)

			// 验证索引健康状态
			exists, err := client.IndexExists(ctx, hookOpts.Index)
			require.NoError(t, err)
			assert.True(t, exists)

			// 清理测试索引
			err = client.DeleteIndex(ctx, hookOpts.Index)
			require.NoError(t, err)
		})
	}
}

func TestIndexTemplateRotation(t *testing.T) {
	// 设置测试配置
	config := getElkConfig()

	client := elk.NewElkClient()
	err := client.Connect(config)
	require.NoError(t, err)
	defer client.Close()

	ctx := context.Background()

	// 创建索引模板
	template := map[string]interface{}{
		"index_patterns": []string{"test-logs-*"},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type": "text",
				},
				"timestamp": map[string]interface{}{
					"type": "date",
				},
				"level": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	// 创建模板
	err = client.CreateIndexTemplate(ctx, "test-logs-template", template)
	require.NoError(t, err)
	defer client.DeleteIndexTemplate(ctx, "test-logs-template")

	// 创建多个时间序列索引
	indices := []string{
		fmt.Sprintf("test-logs-%s", time.Now().AddDate(0, 0, -2).Format("2006.01.02")),
		fmt.Sprintf("test-logs-%s", time.Now().AddDate(0, 0, -1).Format("2006.01.02")),
		fmt.Sprintf("test-logs-%s", time.Now().Format("2006.01.02")),
	}

	// 清理测试索引
	defer func() {
		for _, index := range indices {
			_ = client.DeleteIndex(ctx, index)
		}
	}()

	// 为每个索引写入测试数据
	for _, index := range indices {
		mapping := elk.DefaultIndexMapping()
		err = client.CreateIndex(ctx, index, mapping)
		require.NoError(t, err)

		// 写入测试数据
		docs := make([]interface{}, 10)
		for i := range docs {
			docs[i] = map[string]interface{}{
				"message":   fmt.Sprintf("test message %d", i),
				"timestamp": time.Now(),
				"level":     "info",
			}
		}

		err = client.BulkIndexDocuments(ctx, index, docs)
		require.NoError(t, err)
	}

	// 验证所有索引是否正确创建和写入
	for _, index := range indices {
		exists, err := client.IndexExists(ctx, index)
		require.NoError(t, err)
		assert.True(t, exists)

		stats, err := client.GetIndexStats(ctx, index)
		require.NoError(t, err)

		docCount, ok := stats["_all"].(map[string]interface{})["primaries"].(map[string]interface{})["docs"].(map[string]interface{})["count"].(float64)
		require.True(t, ok)
		assert.Equal(t, float64(10), docCount)
	}
}
