package elk

import (
	"context"
	"encoding/json"
	"strings"

	"gobase/pkg/errors"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// IndexMapping 定义索引映射结构
type IndexMapping struct {
	Settings map[string]interface{} `json:"settings,omitempty"`
	Mappings map[string]interface{} `json:"mappings,omitempty"`
}

// DefaultIndexMapping 返回默认的索引映射
func DefaultIndexMapping() *IndexMapping {
	return &IndexMapping{
		Settings: map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 1,
		},
		Mappings: map[string]interface{}{
			"properties": map[string]interface{}{
				"@timestamp": map[string]interface{}{
					"type": "date",
				},
				"level": map[string]interface{}{
					"type": "keyword",
				},
				"message": map[string]interface{}{
					"type": "text",
				},
				"fields": map[string]interface{}{
					"type": "object",
				},
			},
		},
	}
}

// DeleteIndex 删除索引
func (c *ElkClient) DeleteIndex(ctx context.Context, index string) error {
	if !c.isConnected {
		return errors.NewELKConnectionError("client is not connected", nil)
	}

	if strings.TrimSpace(index) == "" {
		return errors.NewELKIndexError("index name cannot be empty", nil)
	}

	req := esapi.IndicesDeleteRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return errors.NewELKIndexError("failed to delete index", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.NewELKIndexError(
			"failed to delete index: "+res.String(),
			nil,
		)
	}

	return nil
}

// IndexExists 检查索引是否存在
func (c *ElkClient) IndexExists(ctx context.Context, index string) (bool, error) {
	if !c.isConnected {
		return false, errors.NewELKConnectionError("client is not connected", nil)
	}

	if strings.TrimSpace(index) == "" {
		return false, errors.NewELKIndexError("index name cannot be empty", nil)
	}

	req := esapi.IndicesExistsRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return false, errors.NewELKIndexError("failed to check index existence", err)
	}
	defer res.Body.Close()

	return !res.IsError(), nil
}

// GetIndexMapping 获取索引映射
func (c *ElkClient) GetIndexMapping(ctx context.Context, index string) (*IndexMapping, error) {
	if !c.isConnected {
		return nil, errors.NewELKConnectionError("client is not connected", nil)
	}

	if strings.TrimSpace(index) == "" {
		return nil, errors.NewELKIndexError("index name cannot be empty", nil)
	}

	req := esapi.IndicesGetMappingRequest{
		Index: []string{index},
	}

	res, err := req.Do(ctx, c.client)
	if err != nil {
		return nil, errors.NewELKIndexError("failed to get index mapping", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, errors.NewELKIndexError(
			"failed to get index mapping: "+res.String(),
			nil,
		)
	}

	var mapping IndexMapping
	if err := json.NewDecoder(res.Body).Decode(&mapping); err != nil {
		return nil, errors.NewELKIndexError("failed to decode mapping response", err)
	}

	return &mapping, nil
}
