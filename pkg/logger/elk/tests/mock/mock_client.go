package mock

import (
	"context"
	"errors"
	"sync"

	"gobase/pkg/logger/elk"
)

type mockElkClient struct {
	mu            sync.Mutex
	connected     bool
	documents     []interface{}
	shouldFailOps bool
	indexes       map[string]*elk.IndexMapping
}

func NewMockElkClient() *mockElkClient {
	return &mockElkClient{
		connected: true,
		documents: make([]interface{}, 0),
		indexes:   make(map[string]*elk.IndexMapping),
	}
}

func (m *mockElkClient) Connect(config *elk.ElkConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFailOps {
		m.connected = false
		return errors.New("mock connection failure")
	}

	m.connected = true
	return nil
}

func (m *mockElkClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFailOps {
		return errors.New("mock close failure")
	}

	m.connected = false
	return nil
}

func (m *mockElkClient) IndexDocument(ctx context.Context, index string, document interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFailOps {
		return errors.New("mock index failure")
	}

	if !m.connected {
		return errors.New("client not connected")
	}

	m.documents = append(m.documents, document)
	return nil
}

func (m *mockElkClient) BulkIndexDocuments(ctx context.Context, index string, documents []interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFailOps {
		return errors.New("mock bulk index failure")
	}

	m.documents = append(m.documents, documents...)
	return nil
}

func (m *mockElkClient) Query(ctx context.Context, index string, query interface{}) (interface{}, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFailOps {
		return nil, errors.New("mock query failure")
	}

	if !m.connected {
		return nil, errors.New("client not connected")
	}

	// 返回模拟的查询结果
	return map[string]interface{}{
		"hits": map[string]interface{}{
			"total": len(m.documents),
			"hits":  m.documents,
		},
	}, nil
}

func (m *mockElkClient) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.connected
}

// 测试辅助方法
func (m *mockElkClient) SetShouldFailOps(should bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFailOps = should
}

func (m *mockElkClient) GetDocuments() []interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.documents
}

func (m *mockElkClient) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.documents = make([]interface{}, 0)
	m.connected = false
	m.shouldFailOps = false
}

// 添加类型断言辅助方法
func AsMockClient(client elk.Client) *mockElkClient {
	if mock, ok := client.(*mockElkClient); ok {
		return mock
	}
	return nil
}

func (m *mockElkClient) CreateIndex(ctx context.Context, index string, mapping *elk.IndexMapping) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFailOps {
		return errors.New("mock create index failure")
	}

	m.indexes[index] = mapping
	return nil
}

func (m *mockElkClient) DeleteIndex(ctx context.Context, index string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFailOps {
		return errors.New("mock delete index failure")
	}

	delete(m.indexes, index)
	return nil
}

func (m *mockElkClient) IndexExists(ctx context.Context, index string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFailOps {
		return false, errors.New("mock index exists failure")
	}

	_, exists := m.indexes[index]
	return exists, nil
}

func (m *mockElkClient) GetIndexMapping(ctx context.Context, index string) (*elk.IndexMapping, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shouldFailOps {
		return nil, errors.New("mock get index mapping failure")
	}

	mapping, exists := m.indexes[index]
	if !exists {
		return nil, errors.New("index does not exist")
	}

	return mapping, nil
}

// 确保类型实现了接口
var _ elk.Client = (*mockElkClient)(nil)
