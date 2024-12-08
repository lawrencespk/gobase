package mock

import (
	"sync"

	"gobase/pkg/errors"

	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/mock"
)

// MockConfigClient 模拟Nacos配置客户端
type MockConfigClient struct {
	mu       sync.RWMutex
	configs  map[string]string
	watches  map[string]func(namespace, group, dataId, data string)
	hasError bool
	mock.Mock
}

// NewMockConfigClient 创建新的mock客户端
func NewMockConfigClient() *MockConfigClient {
	return &MockConfigClient{
		configs: make(map[string]string),
		watches: make(map[string]func(namespace, group, dataId, data string)),
	}
}

// SetError 设置错误标志
func (m *MockConfigClient) SetError(hasError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hasError = hasError
}

// GetConfig 获取配置
func (m *MockConfigClient) GetConfig(param vo.ConfigParam) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.hasError {
		return "", errors.NewConfigError("mock error", nil)
	}

	content, ok := m.configs[param.DataId]
	if !ok {
		return "", errors.NewConfigError("config not found", nil)
	}

	return content, nil
}

// PublishConfig 发布配置
func (m *MockConfigClient) PublishConfig(param vo.ConfigParam) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.hasError {
		return false, errors.NewConfigError("mock error", nil)
	}

	m.configs[param.DataId] = param.Content

	if watch, ok := m.watches[param.DataId]; ok {
		go watch("", param.Group, param.DataId, param.Content)
	}
	return true, nil
}

// ListenConfig 监听配置
func (m *MockConfigClient) ListenConfig(param vo.ConfigParam) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.watches[param.DataId] = param.OnChange
	return nil
}

// CancelListenConfig 取消配置监听
func (m *MockConfigClient) CancelListenConfig(param vo.ConfigParam) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.watches, param.DataId)
	return nil
}

// SetConfig 设置mock配置内容
func (m *MockConfigClient) SetConfig(dataId string, content string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.configs[dataId] = content

	if watch, ok := m.watches[dataId]; ok {
		go watch("", "", dataId, content)
	}
}

// DeleteConfig 删除配置
func (m *MockConfigClient) DeleteConfig(param vo.ConfigParam) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.configs, param.DataId)
	return true, nil
}

// PublishAggr 发布聚合配置
func (m *MockConfigClient) PublishAggr(param vo.ConfigParam) (bool, error) {
	// 在mock环境中，我们可以简单地调用普通的发布方法
	return m.PublishConfig(param)
}

// SearchConfig 搜索配置
func (m *MockConfigClient) SearchConfig(param vo.SearchConfigParam) (*model.ConfigPage, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 创建一个简单的响应
	return &model.ConfigPage{
		TotalCount: 0,
		PageItems:  []model.ConfigItem{},
	}, nil
}

// CloseClient 实现关闭方法
func (m *MockConfigClient) CloseClient() {
	m.Called()
}
