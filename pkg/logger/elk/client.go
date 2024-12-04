package elk

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/olivere/elastic/v7"
)

// ElkClient Elasticsearch客户端
type ElkClient struct {
	client    *elastic.Client        // ES客户端
	config    *ElasticConfig         // 配置
	buffer    *logBuffer             // 日志缓冲
	processor *elastic.BulkProcessor // 批处理器
	mu        sync.RWMutex           // 读写锁
	closed    bool                   // 是否已关闭
}

// NewElkClient 创建新的ES客户端
func NewElkClient(config *ElasticConfig) (*ElkClient, error) {
	if config == nil {
		config = DefaultElasticConfig() // 使用默认配置
	}

	// 创建ES客户端选项
	options := []elastic.ClientOptionFunc{
		elastic.SetURL(config.Addresses...),        // 设置URL
		elastic.SetSniff(config.Sniff),             // 设置嗅探
		elastic.SetHealthcheck(config.Healthcheck), // 设置健康检查
		elastic.SetRetrier(elastic.NewBackoffRetrier(elastic.NewExponentialBackoff(time.Millisecond, 5*time.Second))), // 设置重试器
	}

	// 添加认证
	if config.Username != "" && config.Password != "" {
		options = append(options, elastic.SetBasicAuth(config.Username, config.Password)) // 设置基本认证
	}

	// 添加TLS配置
	if config.TLS != nil {
		options = append(options, elastic.SetHttpClient(&http.Client{ // 设置HTTP客户端
			Transport: &http.Transport{
				TLSClientConfig: config.TLS, // 设置TLS配置
			},
			Timeout: config.DialTimeout, // 设置超时时间
		}))
	}

	// 创建客户端
	client, err := elastic.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err) // 创建ES客户端失败
	}

	// 创建ElkClient实例
	elkClient := &ElkClient{
		client: client, // ES客户端
		config: config, // 配置
	}

	// 创建日志缓冲区
	elkClient.buffer = newLogBuffer(client, config)

	// 创建批处理器
	processor, err := client.BulkProcessor().
		Name("ElkLogProcessor").             // 设置名称
		Workers(config.Workers).             // 设置工作线程数
		BulkActions(config.BatchSize).       // 设置批量大小
		FlushInterval(config.FlushInterval). // 设置刷新间隔
		Do(context.Background())             // 执行
	if err != nil {
		return nil, fmt.Errorf("failed to create bulk processor: %w", err) // 创建批处理器失败
	}

	elkClient.processor = processor
	return elkClient, nil
}

// CreateIndex 创建索引
func (c *ElkClient) CreateIndex(indexName string) error {
	ctx := context.Background()                            // 上下文
	exists, err := c.client.IndexExists(indexName).Do(ctx) // 检查索引是否存在
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err) // 检查索引是否存在失败
	}

	if !exists {
		// 创建索引
		createIndex := c.client.CreateIndex(indexName)

		// 设置索引配置
		indexSettings := map[string]interface{}{
			"settings": map[string]interface{}{ // 设置索引配置
				"number_of_shards":   c.config.NumberOfShards,   // 分片数
				"number_of_replicas": c.config.NumberOfReplicas, // 副本数
				"refresh_interval":   c.config.RefreshInterval,  // 刷新间隔
			},
		}

		_, err = createIndex.BodyJson(indexSettings).Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err) // 创建索引失败
		}
	}

	return nil
}

// Write 写入日志
func (c *ElkClient) Write(entry map[string]interface{}) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return fmt.Errorf("client is closed") // 客户端已关闭
	}
	c.mu.RUnlock()

	// 获取当前索引名称
	indexName := fmt.Sprintf("%s-%s", c.config.IndexPrefix, time.Now().Format("2006.01.02"))

	// 确保索引存在
	if err := c.CreateIndex(indexName); err != nil {
		return fmt.Errorf("failed to ensure index exists: %w", err) // 确保索引存在失败
	}

	// 创建索引请求
	req := elastic.NewBulkIndexRequest().
		Index(indexName). // 设置索引
		Doc(entry)        // 设置文档

	// 添加到批处理器
	c.processor.Add(req) // 添加请求

	return nil
}

// Close 关闭客户端
func (c *ElkClient) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	c.mu.Unlock()

	// 关闭缓冲区
	if err := c.buffer.Close(); err != nil {
		return fmt.Errorf("failed to close buffer: %w", err) // 关闭缓冲区失败
	}

	// 关闭批处理器
	if err := c.processor.Close(); err != nil {
		return fmt.Errorf("failed to close bulk processor: %w", err) // 关闭批处理器失败
	}

	return nil
}
