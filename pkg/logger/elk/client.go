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
	Do(ctx context.Context) (bool, error)
}

// IndicesCreateServiceInterface 定义索引创建服务的接口
type IndicesCreateServiceInterface interface {
	BodyJson(mapping map[string]interface{}) IndicesCreateServiceInterface
	Do(ctx context.Context) (interface{}, error)
}

// MockElasticClientInterface 定义模拟客户端需要实现的接口
type MockElasticClientInterface interface {
	IndexExists(indices ...string) IndicesExistsServiceInterface
	CreateIndex(name string) IndicesCreateServiceInterface
	// 其他需要的方法
}

// ElasticClientInterface 统一的客户端接口
type ElasticClientInterface interface {
	MockElasticClientInterface
	BulkProcessor() *elastic.BulkProcessorService
	// 添加其他需要的方法
}

// MockElasticClient 用于测试的模拟客户端
type MockElasticClient struct {
	mock.Mock
}

func (m *MockElasticClient) IndexExists(indices ...string) IndicesExistsServiceInterface {
	args := m.Called(indices)
	return args.Get(0).(IndicesExistsServiceInterface)
}

func (m *MockElasticClient) CreateIndex(name string) IndicesCreateServiceInterface {
	args := m.Called(name)
	return args.Get(0).(IndicesCreateServiceInterface)
}

func (m *MockElasticClient) BulkProcessor() *elastic.BulkProcessorService {
	args := m.Called()
	return args.Get(0).(*elastic.BulkProcessorService)
}

func (m *MockElasticClient) Do(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

// ElasticClientAdapter 适配器结构体
type ElasticClientAdapter struct {
	*elastic.Client
}

// IndexExists 实现接口方法
func (a *ElasticClientAdapter) IndexExists(indices ...string) IndicesExistsServiceInterface {
	return &IndicesExistsServiceAdapter{a.Client.IndexExists(indices...)}
}

// CreateIndex 实现接口方法
func (a *ElasticClientAdapter) CreateIndex(name string) IndicesCreateServiceInterface {
	return &IndicesCreateServiceAdapter{
		IndicesCreateService: a.Client.CreateIndex(name),
	}
}

// BulkProcessor 实现接口方法
func (a *ElasticClientAdapter) BulkProcessor() *elastic.BulkProcessorService {
	return a.Client.BulkProcessor()
}

// IndicesExistsServiceAdapter 适配器
type IndicesExistsServiceAdapter struct {
	*elastic.IndicesExistsService
}

// IndicesCreateServiceAdapter 适配器
type IndicesCreateServiceAdapter struct {
	*elastic.IndicesCreateService
}

// BodyJson 实现接口方法
func (a *IndicesCreateServiceAdapter) BodyJson(mapping map[string]interface{}) IndicesCreateServiceInterface {
	// 调用底层服务的 BodyJson 方法，并返回包装后的适配器
	return &IndicesCreateServiceAdapter{
		IndicesCreateService: a.IndicesCreateService.BodyJson(mapping),
	}
}

// Do 实现接口方法
func (a *IndicesCreateServiceAdapter) Do(ctx context.Context) (interface{}, error) {
	return a.IndicesCreateService.Do(ctx)
}

// ElkClient Elasticsearch客户端
type ElkClient struct {
	client    ElasticClientInterface
	config    *ElasticConfig
	buffer    *logBuffer
	processor *elastic.BulkProcessor
	mu        sync.RWMutex
	closed    bool
}

// SetClient 设置 Elasticsearch 客户端 (用于测试)
func (c *ElkClient) SetClient(client interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if esClient, ok := client.(*elastic.Client); ok {
		c.client = &ElasticClientAdapter{esClient}
	} else if mockClient, ok := client.(ElasticClientInterface); ok {
		c.client = mockClient
	}
}

// NewElkClient 创建新的ES客户端
func NewElkClient(config *ElasticConfig) (*ElkClient, error) {
	if config == nil {
		config = DefaultElasticConfig()
	}

	// 创建ES客户端选项
	options := []elastic.ClientOptionFunc{
		elastic.SetURL(config.Addresses...),
		elastic.SetSniff(config.Sniff),
		elastic.SetHealthcheck(config.Healthcheck),
		elastic.SetRetrier(elastic.NewBackoffRetrier(elastic.NewExponentialBackoff(time.Millisecond, 5*time.Second))),
	}

	// 添加认证
	if config.Username != "" && config.Password != "" {
		options = append(options, elastic.SetBasicAuth(config.Username, config.Password))
	}

	// 添加TLS配置
	if config.TLS != nil {
		options = append(options, elastic.SetHttpClient(&http.Client{
			Transport: &http.Transport{
				TLSClientConfig: config.TLS,
			},
			Timeout: config.DialTimeout,
		}))
	}

	// 创建客户端
	esClient, err := elastic.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// 创建ElkClient实例，使用适配器包装elastic.Client
	elkClient := &ElkClient{
		client: &ElasticClientAdapter{esClient},
		config: config,
	}

	// 创建日志缓冲区
	elkClient.buffer = newLogBuffer(esClient, config)

	// 创建批处理器
	processor, err := esClient.BulkProcessor().
		Name("ElkLogProcessor").
		Workers(config.Workers).
		BulkActions(config.BatchSize).
		FlushInterval(config.FlushInterval).
		Do(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create bulk processor: %w", err)
	}

	elkClient.processor = processor
	return elkClient, nil
}

// CreateIndex 创建索引
func (c *ElkClient) CreateIndex(indexName string) error {
	if c.client == nil {
		return fmt.Errorf("client is not initialized")
	}

	// 验证设置
	if c.config.NumberOfShards <= 0 {
		return fmt.Errorf("invalid settings: number_of_shards must be greater than 0")
	}

	fmt.Printf("Creating index: %s\n", indexName)
	ctx := context.Background()
	exists, err := c.client.IndexExists(indexName).Do(ctx)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}

	if !exists {
		createIndex := c.client.CreateIndex(indexName)
		indexSettings := map[string]interface{}{
			"settings": map[string]interface{}{
				"number_of_shards":   c.config.NumberOfShards,
				"number_of_replicas": c.config.NumberOfReplicas,
				"refresh_interval":   c.config.RefreshInterval,
			},
		}

		_, err = createIndex.BodyJson(indexSettings).Do(ctx)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
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

// 为了支持测试，添加一个用于测试的构造函数
func NewTestElkClient(mockClient ...ElasticClientInterface) *ElkClient {
	client := &ElkClient{
		config: DefaultElasticConfig(),
	}

	if len(mockClient) > 0 {
		client.client = mockClient[0]
	} else {
		client.client = new(MockElasticClient)
	}

	return client
}

// SetConfig 设置配置（用于测试）
func (c *ElkClient) SetConfig(config *ElasticConfig) {
	c.config = config
}
