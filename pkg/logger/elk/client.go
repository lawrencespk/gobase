package elk

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gobase/pkg/errors"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/elastic/go-elasticsearch/v8/esutil"
)

// Client 定义了 ElkClient 的接口
type Client interface {
	Connect(config *ElkConfig) error
	Close() error
	IndexDocument(ctx context.Context, index string, document interface{}) error
	BulkIndexDocuments(ctx context.Context, index string, documents []interface{}) error
	Query(ctx context.Context, index string, query interface{}) (interface{}, error)
	IsConnected() bool

	// 索引管理方法（实现在 index.go 中）
	CreateIndex(ctx context.Context, index string, mapping *IndexMapping) error
	DeleteIndex(ctx context.Context, index string) error
	IndexExists(ctx context.Context, index string) (bool, error)
	GetIndexMapping(ctx context.Context, index string) (*IndexMapping, error)
	CreateIndexTemplate(ctx context.Context, templateName string, template map[string]interface{}) error
	DeleteIndexTemplate(ctx context.Context, templateName string) error
}

// ElkClient 实现 Client 接口
type ElkClient struct {
	config      *ElkConfig
	client      *elasticsearch.Client
	isConnected bool
}

// NewElkClient 创建新的 ELK 客户端
func NewElkClient() *ElkClient {
	return &ElkClient{
		isConnected: false,
	}
}

// Connect 连接到 Elasticsearch 服务器
func (c *ElkClient) Connect(cfg *ElkConfig) error {
	if err := c.validateConfig(cfg); err != nil {
		return err
	}

	c.config = cfg
	// cfg := elasticsearch.Config{
	// 	Addresses: config.Addresses,
	// 	Username:  config.Username,
	// 	Password:  config.Password,
	//
	esCfg := elasticsearch.Config{
		Addresses: cfg.Addresses,
		Username:  cfg.Username,
		Password:  cfg.Password,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		RetryOnStatus: []int{502, 503, 504},
		MaxRetries:    3,
		RetryBackoff: func(i int) time.Duration {
			return time.Duration(i) * time.Second
		},
		RetryOnError: func(req *http.Request, err error) bool {
			return err != nil
		},
	}
	//client, err := elasticsearch.NewClient(cfg)
	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return errors.NewELKConnectionError("failed to create elasticsearch client", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
	defer cancel()

	_, err = client.Info(client.Info.WithContext(ctx))
	if err != nil {
		return errors.NewELKConnectionError("failed to connect to elasticsearch", err)
	}

	c.client = client
	//c.config = config
	c.config = cfg
	c.isConnected = true

	return nil
}

// validateConfig 验证配置
func (c *ElkClient) validateConfig(config *ElkConfig) error {
	if config == nil {
		return errors.NewELKConfigError("configuration is nil", nil)
	}

	if len(config.Addresses) == 0 {
		return errors.NewELKConfigError("no elasticsearch addresses provided", nil)
	}

	if config.Timeout <= 0 {
		return errors.NewELKConfigError("invalid timeout value", nil)
	}

	return nil
}

// IndexDocument 索引单个文档
func (c *ElkClient) IndexDocument(ctx context.Context, index string, document interface{}) error {
	if !c.isConnected {
		return errors.NewELKConnectionError("client is not connected", nil)
	}

	payload, err := json.Marshal(document)
	if err != nil {
		return errors.NewELKIndexError("failed to marshal document", err)
	}

	req := esapi.IndexRequest{
		Index:      index,
		Body:       bytes.NewReader(payload),
		Refresh:    "false",
		FilterPath: []string{"result"},
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return errors.NewELKIndexError("failed to execute index request", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.NewELKIndexError(
			fmt.Sprintf("failed to index document: %s", res.String()),
			nil,
		)
	}

	return nil
}

// BulkIndexDocuments 批量索引文档
func (c *ElkClient) BulkIndexDocuments(ctx context.Context, index string, documents []interface{}) error {
	if !c.isConnected {
		return errors.NewELKConnectionError("client is not connected", nil)
	}

	if len(documents) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for i, doc := range documents {
		// 写入元数据行
		metaLine := fmt.Sprintf(`{"index":{"_index":"%s"}}`, index)
		if _, err := buf.WriteString(metaLine + "\n"); err != nil {
			return errors.NewELKBulkError(fmt.Sprintf("failed to write metadata for doc %d: %v", i, err), err)
		}

		// 处理文档内容
		docBytes, err := json.Marshal(doc)
		if err != nil {
			return errors.NewELKBulkError(fmt.Sprintf("failed to marshal document %d: %v", i, err), err)
		}

		// 写入文档
		if _, err := buf.Write(docBytes); err != nil {
			return errors.NewELKBulkError(fmt.Sprintf("failed to write document %d: %v", i, err), err)
		}

		// 每个文档后添加换行符
		if _, err := buf.WriteString("\n"); err != nil {
			return errors.NewELKBulkError(fmt.Sprintf("failed to write newline after document %d: %v", i, err), err)
		}
	}

	// 发送批量请求
	req := esapi.BulkRequest{
		Body:    bytes.NewReader(buf.Bytes()),
		Refresh: "true",
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return errors.NewELKBulkError("failed to execute bulk request", err)
	}
	defer res.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return errors.NewELKBulkError("failed to read response body", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return errors.NewELKBulkError(
			fmt.Sprintf("failed to parse response: %s", string(respBody)),
			err,
		)
	}

	// 检查响应中的错误
	if res.IsError() {
		return errors.NewELKBulkError(
			fmt.Sprintf("批量索引失败: %s", string(respBody)),
			nil,
		)
	}

	return nil
}

// Query 查询文档
func (c *ElkClient) Query(ctx context.Context, index string, query interface{}) (interface{}, error) {
	if !c.isConnected {
		return nil, errors.NewELKConnectionError("client is not connected", nil)
	}

	payload, err := json.Marshal(query)
	if err != nil {
		return nil, errors.NewELKQueryError("failed to marshal query", err)
	}

	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  bytes.NewReader(payload),
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, errors.NewELKQueryError("failed to execute search request", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.NewELKQueryError(
			fmt.Sprintf("search request failed: %s", res.String()),
			nil,
		)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, errors.NewELKQueryError("failed to parse search response", err)
	}

	return result, nil
}

// Close 闭客户端连接
func (c *ElkClient) Close() error {
	if !c.isConnected {
		return nil
	}

	c.client = nil
	c.config = nil
	c.isConnected = false

	return nil
}

// IsConnected 检查客户端是否已连接
func (c *ElkClient) IsConnected() bool {
	return c.isConnected
}

// CreateIndexTemplate 创建索引模板
func (c *ElkClient) CreateIndexTemplate(ctx context.Context, templateName string, template map[string]interface{}) error {
	if !c.isConnected {
		return errors.NewELKConnectionError("client is not connected", nil)
	}

	req := esapi.IndicesPutTemplateRequest{
		Name: templateName,
		Body: esutil.NewJSONReader(template),
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return errors.NewELKIndexError("failed to create index template", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.NewELKIndexError("failed to create index template: "+res.String(), nil)
	}

	return nil
}

// DeleteIndexTemplate 删除索引模板
func (c *ElkClient) DeleteIndexTemplate(ctx context.Context, templateName string) error {
	if !c.isConnected {
		return errors.NewELKConnectionError("client is not connected", nil)
	}

	req := esapi.IndicesDeleteTemplateRequest{
		Name: templateName,
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return errors.NewELKIndexError("failed to delete index template", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.NewELKIndexError("failed to delete index template: "+res.String(), nil)
	}

	return nil
}

// 添加的方法到 ElkClient
func (c *ElkClient) HealthCheck(ctx context.Context) error {
	resp, err := c.client.Cluster.Health(
		c.client.Cluster.Health.WithContext(ctx),
	)
	if err != nil {
		return errors.NewELKConnectionError("health check failed", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return errors.NewELKConnectionError("health check response error", nil)
	}

	return nil
}

func (c *ElkClient) CreateIndex(ctx context.Context, indexName string, mapping *IndexMapping) error {
	body, err := json.Marshal(mapping)
	if err != nil {
		return errors.NewELKIndexError("failed to marshal index mapping", err)
	}

	resp, err := c.client.Indices.Create(
		indexName,
		c.client.Indices.Create.WithBody(strings.NewReader(string(body))),
		c.client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return errors.NewELKIndexError("failed to create index", err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return errors.NewELKIndexError("index creation failed", nil)
	}

	return nil
}

// RefreshIndex 刷新指定的索引
func (c *ElkClient) RefreshIndex(ctx context.Context, index string) error {
	req := esapi.IndicesRefreshRequest{
		Index: []string{index},
	}
	res, err := req.Do(ctx, c.client)
	if err != nil {
		return errors.NewELKIndexError("failed to refresh index", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.NewELKIndexError("index refresh failed", nil)
	}
	return nil
}

// GetIndexStats 获取索引的统计信息
func (c *ElkClient) GetIndexStats(ctx context.Context, index string) (map[string]interface{}, error) {
	req := esapi.IndicesStatsRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, errors.NewELKIndexError("failed to get index stats", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.NewELKIndexError("failed to get index stats", nil)
	}

	var stats map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return nil, errors.NewELKIndexError("failed to parse index stats", err)
	}

	return stats, nil
}
