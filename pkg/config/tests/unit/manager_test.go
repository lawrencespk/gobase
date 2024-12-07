package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gobase/pkg/config"
)

// MockLoader 模拟配置加载器
type MockLoader struct {
	mock.Mock
}

func (m *MockLoader) Load() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockLoader) Get(key string) interface{} {
	args := m.Called(key)
	return args.Get(0)
}

func (m *MockLoader) GetString(key string) string {
	args := m.Called(key)
	return args.String(0)
}

func (m *MockLoader) GetInt(key string) int {
	args := m.Called(key)
	return args.Int(0)
}

func (m *MockLoader) GetBool(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockLoader) GetDuration(key string) time.Duration {
	args := m.Called(key)
	return args.Get(0).(time.Duration)
}

func (m *MockLoader) IsSet(key string) bool {
	args := m.Called(key)
	return args.Bool(0)
}

func (m *MockLoader) Watch(key string, callback func(key string, value interface{})) error {
	args := m.Called(key, callback)
	return args.Error(0)
}

// GetFloat64 获取浮点数配置
func (m *MockLoader) GetFloat64(key string) float64 {
	args := m.Called(key)
	return args.Get(0).(float64)
}

// GetStringSlice 获取字符串切片配置
func (m *MockLoader) GetStringSlice(key string) []string {
	args := m.Called(key)
	return args.Get(0).([]string)
}

// GetStringMap 获取字符串映射配置
func (m *MockLoader) GetStringMap(key string) map[string]interface{} {
	args := m.Called(key)
	return args.Get(0).(map[string]interface{})
}

// GetStringMapString 获取字符串映射字符串配置
func (m *MockLoader) GetStringMapString(key string) map[string]string {
	args := m.Called(key)
	return args.Get(0).(map[string]string)
}

// Set 设置配置值
func (m *MockLoader) Set(key string, value interface{}) {
	m.Called(key, value)
}

// MockParser 模拟配置解析器
type MockParser struct {
	mock.Mock
}

func (m *MockParser) Parse(key string, out interface{}) error {
	args := m.Called(key, out)
	return args.Error(0)
}

func TestManager_GetValues_Unit(t *testing.T) {
	// 创建mock
	mockLoader := new(MockLoader)
	mockParser := new(MockParser)

	// 设置预期行为
	mockLoader.On("GetString", "app.server.host").Return("0.0.0.0")
	mockLoader.On("GetInt", "app.server.port").Return(8080)
	mockLoader.On("GetDuration", "app.database.timeout").Return(30 * time.Second)
	mockLoader.On("GetString", "non.existent.key").Return("")

	// 创建管理器实例
	manager := config.NewManagerWithLoader(mockLoader, mockParser)

	t.Run("get string value", func(t *testing.T) {
		value := manager.GetString("app.server.host")
		assert.Equal(t, "0.0.0.0", value)
	})

	t.Run("get int value", func(t *testing.T) {
		value := manager.GetInt("app.server.port")
		assert.Equal(t, 8080, value)
	})

	t.Run("get duration value", func(t *testing.T) {
		value := manager.GetDuration("app.database.timeout")
		assert.Equal(t, 30*time.Second, value)
	})

	t.Run("get non-existent value", func(t *testing.T) {
		value := manager.GetString("non.existent.key")
		assert.Empty(t, value)
	})

	// 验证所有预期的调用都已发生
	mockLoader.AssertExpectations(t)
}

func TestManager_Parse_Unit(t *testing.T) {
	// 创建mock
	mockLoader := new(MockLoader)
	mockParser := new(MockParser)

	// 创建测试数据
	testConfig := struct {
		Host string
		Port int
	}{}

	// 设置预期行为
	mockParser.On("Parse", "app.server", &testConfig).Return(nil)
	mockParser.On("Parse", "non.existent", mock.Anything).Return(assert.AnError)

	// 创建管理器实例
	manager := config.NewManagerWithLoader(mockLoader, mockParser)

	t.Run("parse valid config", func(t *testing.T) {
		err := manager.Parse("app.server", &testConfig)
		require.NoError(t, err)
	})

	t.Run("parse invalid key", func(t *testing.T) {
		err := manager.Parse("non.existent", &testConfig)
		assert.Error(t, err)
	})

	// 验证所有预期的调用都已发生
	mockParser.AssertExpectations(t)
}

func TestManager_Watch_Unit(t *testing.T) {
	// 创建mock
	mockLoader := new(MockLoader)
	mockParser := new(MockParser)

	// 设置预期行为
	callback := func(key string, value interface{}) {}
	mockLoader.On("Watch", "app.server.port", mock.AnythingOfType("func(string, interface {})")).Return(nil)

	// 创建管理器实例
	manager := config.NewManagerWithLoader(mockLoader, mockParser)

	t.Run("watch config change", func(t *testing.T) {
		err := manager.Watch("app.server.port", callback)
		require.NoError(t, err)
	})

	// 验证所有预期的调用都已发生
	mockLoader.AssertExpectations(t)
}
