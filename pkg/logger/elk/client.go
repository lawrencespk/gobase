package elk

import (
	"context"
	"fmt"
	"time"

	"github.com/olivere/elastic/v7"
)

// Options 配置
type Options struct {
	Workers       int           // 工作线程数
	BatchSize     int           // 批量大小
	FlushInterval time.Duration // 刷新间隔
	RetryCount    int           // 重试次数
	RetryTimeout  time.Duration // 重试超时
}

// defaultOptions 默认配置
var defaultOptions = &Options{
	Workers:       2,                // 工作线程数
	BatchSize:     100,              // 批量大小
	FlushInterval: 5 * time.Second,  // 刷新间隔
	RetryCount:    3,                // 重试次数
	RetryTimeout:  30 * time.Second, // 重试超时
}

// ElkClient Elasticsearch客户端
type ElkClient struct {
	client    *elastic.Client        // Elasticsearch客户端
	index     string                 // Elasticsearch索引
	processor *elastic.BulkProcessor // 批处理器
	opts      *Options               // 配置
}

// NewElkClient 创建ElkClient
func NewElkClient(urls []string, index string, opts *Options) (*ElkClient, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(urls...), // 设置Elasticsearch URL
		elastic.SetSniff(false), // 禁用嗅探
		elastic.SetRetrier(elastic.NewBackoffRetrier(elastic.NewExponentialBackoff(
			opts.RetryTimeout,    // 设置重试超时
			opts.RetryTimeout*10, // 设置重试超时
		))), // 设置重试器
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %v", err)
	}

	processor, err := client.BulkProcessor().
		Name("LogProcessor").              // 批处理器名称
		Workers(opts.Workers).             // 工作线程数
		BulkActions(opts.BatchSize).       // 批量操作数
		FlushInterval(opts.FlushInterval). // 刷新间隔
		Do(context.Background())           // 执行

	if err != nil {
		return nil, fmt.Errorf("failed to create bulk processor: %v", err)
	}

	return &ElkClient{
		client:    client,    // Elasticsearch客户端
		index:     index,     // Elasticsearch索引
		processor: processor, // 批处理器
	}, nil
}

// Write 写入日志
func (c *ElkClient) Write(entry map[string]interface{}) error {
	// 创建索引名称（可以按日期分割）
	indexName := fmt.Sprintf("%s-%s", c.index, time.Now().Format("2006.01.02"))

	// 添加到批处理器
	req := elastic.NewBulkIndexRequest(). // 创建批量索引请求
						Index(indexName). // 设置索引名称
						Doc(entry)        // 设置文档

	c.processor.Add(req) // 添加到批处理器
	return nil
}

// Close 关闭批处理器
func (c *ElkClient) Close() error {
	return c.processor.Close()
}
