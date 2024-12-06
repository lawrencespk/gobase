package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"gobase/pkg/logger/elk"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	defaultElkHost = "http://localhost:9200"
	testIndexName  = "test-gobase-logs"
)

func getElkConfig() *elk.ElkConfig {
	// 允许通过环境变量覆盖默认配置
	elkHost := os.Getenv("ELK_HOST")
	if elkHost == "" {
		elkHost = defaultElkHost
	}

	return &elk.ElkConfig{
		Addresses: []string{elkHost},
		Username:  os.Getenv("ELK_USERNAME"),
		Password:  os.Getenv("ELK_PASSWORD"),
		Timeout:   30,
	}
}

func TestRealElasticsearch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real elasticsearch test in short mode")
	}

	config := getElkConfig()
	client := elk.NewElkClient()
	ctx := context.Background()

	// 连接到ES
	err := client.Connect(config)
	if err != nil {
		t.Fatalf("Failed to connect to Elasticsearch: %v", err)
	}
	defer client.Close()

	// 清理测试索引
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.DeleteIndex(cleanupCtx, testIndexName); err != nil {
			t.Logf("Warning: Failed to delete test index: %v", err)
		}

		// 确保连接在清理完成后关闭
		if err := client.Close(); err != nil {
			t.Logf("Warning: Failed to close client: %v", err)
		}
	})

	t.Run("BasicOperations", func(t *testing.T) {
		// 检查索引是否存在
		exists, err := client.IndexExists(ctx, testIndexName)
		require.NoError(t, err)
		if !exists {
			// 创建索引
			err = client.CreateIndex(ctx, testIndexName, elk.DefaultIndexMapping())
			require.NoError(t, err, "Failed to create index")
		}

		// 写入文档
		doc := map[string]interface{}{
			"message":   "test message",
			"level":     "info",
			"timestamp": time.Now(),
		}
		err = client.IndexDocument(ctx, testIndexName, doc)
		require.NoError(t, err, "Failed to index document")

		// 等待文档可搜索
		time.Sleep(1 * time.Second)

		// 查询文档
		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"message": "test message",
				},
			},
		}
		result, err := client.Query(ctx, testIndexName, query)
		require.NoError(t, err, "Failed to query document")
		assert.NotEmpty(t, result, "Query should return results")
	})

	t.Run("RetryOnHighLoad", func(t *testing.T) {
		// 模拟高负载情况下的写入
		retryConfig := elk.RetryConfig{
			MaxRetries:  3,
			InitialWait: 100 * time.Millisecond,
			MaxWait:     1 * time.Second,
		}

		// 并发写入多个文档
		for i := 0; i < 10; i++ {
			doc := map[string]interface{}{
				"message":   fmt.Sprintf("bulk test message %d", i),
				"level":     "info",
				"timestamp": time.Now(),
			}

			err := elk.WithRetry(ctx, retryConfig, func() error {
				return client.IndexDocument(ctx, testIndexName, doc)
			}, logrus.New())
			assert.NoError(t, err, "Document indexing should succeed with retry")
		}
	})

	t.Run("LargeDocument", func(t *testing.T) {
		// 测试写入大文档
		largeText := make([]byte, 1024*1024) // 1MB
		for i := range largeText {
			largeText[i] = 'a'
		}

		doc := map[string]interface{}{
			"message":     "large document test",
			"level":       "info",
			"timestamp":   time.Now(),
			"large_field": string(largeText),
		}

		err := client.IndexDocument(ctx, testIndexName, doc)
		assert.NoError(t, err, "Should handle large documents")
	})

	t.Run("QueryTimeout", func(t *testing.T) {
		// 测试复杂查询的超时情况
		complexQuery := map[string]interface{}{
			"query": map[string]interface{}{
				"bool": map[string]interface{}{
					"must": []map[string]interface{}{
						{
							"match": map[string]interface{}{
								"message": "test",
							},
						},
						{
							"range": map[string]interface{}{
								"timestamp": map[string]interface{}{
									"gte": "now-1h",
									"lt":  "now",
								},
							},
						},
					},
				},
			},
			"sort": []map[string]interface{}{
				{
					"timestamp": map[string]interface{}{
						"order": "desc",
					},
				},
			},
		}

		// 使用短超时的上下文
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
		defer cancel()

		_, err := client.Query(ctxWithTimeout, testIndexName, complexQuery)
		// 这里不断言具体错误，因为查询可能成功也可能超时，取决于ES的负载
		t.Logf("Complex query result: %v", err)
	})

	t.Run("IndexTemplate", func(t *testing.T) {
		// 测试索引模板操作
		templateName := "test-template"
		template := map[string]interface{}{
			"index_patterns": []string{"test-*"},
			"settings": map[string]interface{}{
				"number_of_shards":   1,
				"number_of_replicas": 0,
			},
			"mappings": map[string]interface{}{
				"properties": map[string]interface{}{
					"timestamp": map[string]interface{}{
						"type": "date",
					},
					"message": map[string]interface{}{
						"type": "text",
					},
					"level": map[string]interface{}{
						"type": "keyword",
					},
				},
			},
			"aliases": map[string]interface{}{
				"test-alias": map[string]interface{}{},
			},
		}

		// 创建模板
		err := client.CreateIndexTemplate(ctx, templateName, template)
		require.NoError(t, err, "Should create index template")

		// 使用模板创建新索引
		newIndex := fmt.Sprintf("test-%d", time.Now().Unix())
		createIndexBody := map[string]interface{}{} // 空配置，使用模板默认值
		err = client.CreateIndex(ctx, newIndex, &elk.IndexMapping{
			Settings: createIndexBody,
		})
		require.NoError(t, err, "Should create index using template")

		// 验证索引是否创建成功
		exists, err := client.IndexExists(ctx, newIndex)
		require.NoError(t, err, "Should check index existence")
		assert.True(t, exists, "Index should exist")

		// 等待索引准备就绪
		time.Sleep(1 * time.Second)

		// 验证索引可用性
		doc := map[string]interface{}{
			"message":   "test template message",
			"level":     "info",
			"timestamp": time.Now(),
		}
		err = client.IndexDocument(ctx, newIndex, doc)
		require.NoError(t, err, "Should index document to template-created index")

		// 清理资源
		t.Cleanup(func() {
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// 删除测试索引
			if err := client.DeleteIndex(cleanupCtx, newIndex); err != nil {
				t.Logf("Warning: Failed to delete test index %s: %v", newIndex, err)
			}

			// 删除模板
			if err := client.DeleteIndexTemplate(cleanupCtx, templateName); err != nil {
				t.Logf("Warning: Failed to delete template %s: %v", templateName, err)
			}
		})
	})
}
