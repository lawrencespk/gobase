package elk

import "context"

// ElkClientInterface 定义与 Elasticsearch 交互的接口
type ElkClientInterface interface {
	Connect(config *ElkConfig) error
	IndexDocument(ctx context.Context, index string, document interface{}) error
	BulkIndexDocuments(ctx context.Context, index string, documents []interface{}) error
	Query(ctx context.Context, index string, query interface{}) (interface{}, error)
	Close() error
}
