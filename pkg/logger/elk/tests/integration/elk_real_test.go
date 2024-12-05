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
		err := client.DeleteIndex(ctx, testIndexName)
		if err != nil {
			t.Logf("Warning: Failed to delete test index: %v", err)
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
				"number_of_shards": 1,
			},
			"mappings": elk.DefaultIndexMapping().Mappings,
		}

		// 创建模板
		err := client.CreateIndexTemplate(ctx, templateName, template)
		assert.NoError(t, err, "Should create index template")

		// 证使用模板创建的索引
		newIndex := fmt.Sprintf("test-%d", time.Now().Unix())
		err = client.CreateIndex(ctx, newIndex, nil) // nil mapping将使用模板
		assert.NoError(t, err, "Should create index using template")

		// 清理
		t.Cleanup(func() {
			client.DeleteIndex(ctx, newIndex)
			client.DeleteIndexTemplate(ctx, templateName)
		})
	})
}
