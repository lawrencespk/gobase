package elk

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"gobase/pkg/errors"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
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

	for _, doc := range documents {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": index,
			},
		}

		if err := json.NewEncoder(&buf).Encode(meta); err != nil {
			return errors.NewELKBulkError("failed to encode metadata", err)
		}

		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			return errors.NewELKBulkError("failed to encode document", err)
		}
	}

	req := esapi.BulkRequest{
		Body:       bytes.NewReader(buf.Bytes()),
		Refresh:    "false",
		FilterPath: []string{"errors", "items.*.error"},
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return errors.NewELKBulkError("failed to execute bulk request", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.NewELKBulkError(
			fmt.Sprintf("bulk operation failed: %s", res.String()),
			nil,
		)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return errors.NewELKBulkError("failed to parse bulk response", err)
	}

	if raw["errors"].(bool) {
		return errors.NewELKBulkError("some documents failed to index", nil)
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

// Close 关闭客户端连接
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
