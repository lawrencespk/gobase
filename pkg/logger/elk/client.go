package elk

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/mock"
)

// IndicesExistsServiceInterface 定义索引存在检查服务的接口
type IndicesExistsServiceInterface interface {
	Do(ctx context.Context) (bool, error) // 检查索引是否存在
}

// IndicesCreateServiceInterface 定义索引创建服务的接口
type IndicesCreateServiceInterface interface {
	BodyJson(mapping map[string]interface{}) IndicesCreateServiceInterface // 设置索引配置
	Do(ctx context.Context) (interface{}, error)                           // 创建索引
}

// MockElasticClientInterface 定义模拟客户端需要实现的接口
type MockElasticClientInterface interface {
	IndexExists(indices ...string) IndicesExistsServiceInterface // 检查索引是否存在
	CreateIndex(name string) IndicesCreateServiceInterface
	// 其他需要的方法
}

// ElasticClientInterface 统一的客户端接口
type ElasticClientInterface interface {
	MockElasticClientInterface                    // 继承模拟客户端接口
	BulkProcessor() *elastic.BulkProcessorService // 批处理器
	// 添加其他需要的方法
}

// MockElasticClient 用于测试的模拟客户端
type MockElasticClient struct {
	mock.Mock // 继承mock.Mock
}

// IndexExists 实现接口方法
func (m *MockElasticClient) IndexExists(indices ...string) IndicesExistsServiceInterface {
	args := m.Called(indices)                          // 调用mock.Called
	return args.Get(0).(IndicesExistsServiceInterface) // 返回接口
}

// CreateIndex 实现接口方法
func (m *MockElasticClient) CreateIndex(name string) IndicesCreateServiceInterface {
	args := m.Called(name)                             // 调用mock.Called
	return args.Get(0).(IndicesCreateServiceInterface) // 返回接口
}

// BulkProcessor 实现接口方法
func (m *MockElasticClient) BulkProcessor() *elastic.BulkProcessorService {
	args := m.Called()                                 // 调用mock.Called
	return args.Get(0).(*elastic.BulkProcessorService) // 返回接口
}

// Do 实现接口方法
func (m *MockElasticClient) Do(ctx context.Context) (bool, error) {
	args := m.Called(ctx)              // 调用mock.Called
	return args.Bool(0), args.Error(1) // 返回布尔值和错误
}

// ElasticClientAdapter 适配器结构体
type ElasticClientAdapter struct {
	*elastic.Client // 继承elastic.Client
}

// IndexExists 实现接口方法
func (a *ElasticClientAdapter) IndexExists(indices ...string) IndicesExistsServiceInterface {
	return &IndicesExistsServiceAdapter{a.Client.IndexExists(indices...)} // 返回适配器
}

// CreateIndex 实现接口方法
func (a *ElasticClientAdapter) CreateIndex(name string) IndicesCreateServiceInterface {
	return &IndicesCreateServiceAdapter{
		IndicesCreateService: a.Client.CreateIndex(name), // 返回适配器
	}
}

// BulkProcessor 实现接口方法
func (a *ElasticClientAdapter) BulkProcessor() *elastic.BulkProcessorService {
	return a.Client.BulkProcessor() // 返回批处理器
}

// IndicesExistsServiceAdapter 适配器
type IndicesExistsServiceAdapter struct {
	*elastic.IndicesExistsService // 继承elastic.IndicesExistsService
}

// IndicesCreateServiceAdapter 适配器
type IndicesCreateServiceAdapter struct {
	*elastic.IndicesCreateService // 继承elastic.IndicesCreateService
}

// BodyJson 实现接口方法
func (a *IndicesCreateServiceAdapter) BodyJson(mapping map[string]interface{}) IndicesCreateServiceInterface {
	// 调用底层服务的 BodyJson 方法，并返回包装后的适配器
	return &IndicesCreateServiceAdapter{
		IndicesCreateService: a.IndicesCreateService.BodyJson(mapping), // 返回适配器
	}
}

// Do 实现接口方法
func (a *IndicesCreateServiceAdapter) Do(ctx context.Context) (interface{}, error) {
	return a.IndicesCreateService.Do(ctx) // 返回接口
}

// ElkClient Elasticsearch客户端
type ElkClient struct {
	client    ElasticClientInterface // Elasticsearch客户端
	config    *ElasticConfig         // 配置
	buffer    *logBuffer             // 日志缓冲区
	processor *elastic.BulkProcessor // 批处理器
	mu        sync.RWMutex           // 互斥锁
	closed    bool                   // 是否关闭
}

// SetClient 设置 Elasticsearch 客户端 (用于测试)
func (c *ElkClient) SetClient(client interface{}) {
	c.mu.Lock()         // 加锁
	defer c.mu.Unlock() // 解锁

	if esClient, ok := client.(*elastic.Client); ok { // 检查是否为elastic.Client
		c.client = &ElasticClientAdapter{esClient} // 设置适配器
	} else if mockClient, ok := client.(ElasticClientInterface); ok { // 检查是否为ElasticClientInterface
		c.client = mockClient // 设置模拟客户端
	}
}

// NewElkClient 创建新的ES客户端
func NewElkClient(config *ElasticConfig) (*ElkClient, error) {
	if config == nil {
		config = DefaultElasticConfig() // 设置默认配置
	}

	// 创建ES客户端选项
	options := []elastic.ClientOptionFunc{
		elastic.SetURL(config.Addresses...),        // 设置地址
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
		options = append(options, elastic.SetHttpClient(&http.Client{
			Transport: &http.Transport{ // 设置HTTP客户端
				TLSClientConfig: config.TLS, // 设置TLS配置
			},
			Timeout: config.DialTimeout, // 设置超时时间
		}))
	}

	// 创建客户端
	esClient, err := elastic.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err) // 创建客户端失败
	}

	// 创建ElkClient实例，使用适配器包装elastic.Client
	elkClient := &ElkClient{
		client: &ElasticClientAdapter{esClient}, // 使用适配器包装elastic.Client
		config: config,                          // 设置配置
	}

	// 创建日志缓冲区
	elkClient.buffer = newLogBuffer(esClient, config) // 创建日志缓冲区

	// 创建批处理器
	processor, err := esClient.BulkProcessor(). // 创建批处理器
							Name("ElkLogProcessor").             // 设置名称
							Workers(config.Workers).             // 设置工作线程数
							BulkActions(config.BatchSize).       // 设置批量大小
							FlushInterval(config.FlushInterval). // 设置刷新间隔
							Do(context.Background())             // 执行
	if err != nil {
		return nil, fmt.Errorf("failed to create bulk processor: %w", err) // 创建批处理器失败
	}

	elkClient.processor = processor // 设置批处理器
	return elkClient, nil
}

// CreateIndex 创建索引
func (c *ElkClient) CreateIndex(indexName string) error {
	if c.client == nil {
		return fmt.Errorf("client is not initialized")
	}

	// 验证设置
	if c.config.NumberOfShards <= 0 {
		return fmt.Errorf("invalid settings: number_of_shards must be greater than 0") // 无效的设置
	}

	fmt.Printf("Creating index: %s\n", indexName) // 打印创建索引
	ctx := context.Background()
	exists, err := c.client.IndexExists(indexName).Do(ctx) // 检查索引是否存在
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if !exists {
		createIndex := c.client.CreateIndex(indexName) // 创建索引
		indexSettings := map[string]interface{}{
			"settings": map[string]interface{}{
				"number_of_shards":   c.config.NumberOfShards,   // 设置分片数
				"number_of_replicas": c.config.NumberOfReplicas, // 设置副本数
				"refresh_interval":   c.config.RefreshInterval,  // 设置刷新间隔
			},
		}

		_, err = createIndex.BodyJson(indexSettings).Do(ctx) // 设置索引配置
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err) // 创建索引失败
		}
	}

	return nil
}

// Write 写入日志
func (c *ElkClient) Write(entry map[string]interface{}) error {
	c.mu.RLock() // 加锁
	if c.closed {
		c.mu.RUnlock()                        // 解锁
		return fmt.Errorf("client is closed") // 客户端已关闭
	}
	c.mu.RUnlock() // 解锁

	// 获取当前索引名称
	indexName := fmt.Sprintf("%s-%s", c.config.IndexPrefix, time.Now().Format("2006.01.02")) // 设置索引名称

	// 确保索引存在
	if err := c.CreateIndex(indexName); err != nil {
		return fmt.Errorf("failed to ensure index exists: %w", err) // 确保索引存在失败
	}

	// 创建索引请求
	req := elastic.NewBulkIndexRequest(). // 创建索引请求
						Index(indexName). // 设置索引
						Doc(entry)        // 设置文档

	// 添加到批处理器
	c.processor.Add(req) // 添加请求

	return nil
}

// Close 关闭客户端
func (c *ElkClient) Close() error {
	c.mu.Lock() // 加锁
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true // 设置已关闭
	c.mu.Unlock()   // 解锁

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

// 为了支持测试，添加一个用于测试的构造函数
func NewTestElkClient(mockClient ...ElasticClientInterface) *ElkClient {
	client := &ElkClient{ // 创建ElkClient实例
		config: DefaultElasticConfig(), // 设置默认配置
	}

	if len(mockClient) > 0 {
		client.client = mockClient[0] // 设置模拟客户端
	} else {
		client.client = new(MockElasticClient) // 创建模拟客户端
	}

	return client
}

// SetConfig 设置配置（用于测试）
func (c *ElkClient) SetConfig(config *ElasticConfig) {
	c.config = config // 设置配置
}
