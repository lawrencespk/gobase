package tests

import (
	"context"
	"fmt"
	"gobase/pkg/logger/elk"
	"testing"

	"github.com/olivere/elastic/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCreateService 完全模拟的创建服务
type MockCreateService struct {
	mock.Mock
	indexName string
}

// NewMockCreateService 构造函数
func NewMockCreateService(indexName string) *MockCreateService {
	return &MockCreateService{
		indexName: indexName,
	}
}

// BodyJson 实现 IndicesCreateServiceInterface 接口
func (m *MockCreateService) BodyJson(mapping map[string]interface{}) elk.IndicesCreateServiceInterface {
	m.Called(mapping)
	return m
}

// Do 实现 IndicesCreateServiceInterface 接口
func (m *MockCreateService) Do(ctx context.Context) (interface{}, error) {
	args := m.Called(ctx)
	return args.Get(0), args.Error(1)
}

// MockExistsService 完全模拟的存在检查服务
type MockExistsService struct {
	mock.Mock
}

func (m *MockExistsService) Do(ctx context.Context) (bool, error) {
	args := m.Called(ctx)
	return args.Bool(0), args.Error(1)
}

// MockClient 完全模拟的客户端
type MockClient struct {
	mock.Mock
}

// IndexExists 实现 ClientInterface 接口
func (m *MockClient) IndexExists(indices ...string) elk.IndicesExistsServiceInterface {
	args := m.Called(indices)
	return args.Get(0).(elk.IndicesExistsServiceInterface)
}

// CreateIndex 实现 ClientInterface 接口
func (m *MockClient) CreateIndex(name string) elk.IndicesCreateServiceInterface {
	args := m.Called(name)
	return args.Get(0).(elk.IndicesCreateServiceInterface)
}

// 添加 BulkProcessor 方法
func (m *MockClient) BulkProcessor() *elastic.BulkProcessorService {
	args := m.Called()
	return args.Get(0).(*elastic.BulkProcessorService)
}

// TestCreateIndex 测试创建索引
func TestCreateIndex(t *testing.T) {
	mockClient := new(MockClient)
	elkClient := elk.NewTestElkClient(mockClient)

	indexName := "test-index"
	fmt.Printf("Testing index creation: %s\n", indexName)

	// 设置 IndexExists 的期望
	mockExistsService := new(MockExistsService)
	mockClient.On("IndexExists", []string{indexName}).Return(mockExistsService)
	mockExistsService.On("Do", mock.Anything).Return(false, nil)

	// 设置 CreateIndex 的期望
	mockCreateService := NewMockCreateService(indexName)
	mockClient.On("CreateIndex", indexName).Return(mockCreateService)

	// 设置 BodyJson 的期望，返回 mockCreateService 自身
	mockCreateService.On("BodyJson", mock.MatchedBy(func(mapping map[string]interface{}) bool {
		settings, ok := mapping["settings"].(map[string]interface{})
		if !ok || settings == nil {
			return false
		}
		_, hasShards := settings["number_of_shards"]
		_, hasReplicas := settings["number_of_replicas"]
		_, hasRefresh := settings["refresh_interval"]
		return hasShards && hasReplicas && hasRefresh
	})).Return(mockCreateService)

	// 设置 Do 的期望
	mockCreateService.On("Do", mock.Anything).Return(&elastic.IndicesCreateResult{
		Acknowledged: true,
		Index:        indexName,
	}, nil)

	err := elkClient.CreateIndex(indexName)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockExistsService.AssertExpectations(t)
	mockCreateService.AssertExpectations(t)
}

// TestCreateIndexWhenIndexExists 测试当索引存在时创建索引
func TestCreateIndexWhenIndexExists(t *testing.T) {
	mockClient := new(MockClient)
	elkClient := elk.NewTestElkClient(mockClient)

	indexName := "existing-index"

	mockExistsService := new(MockExistsService)
	mockExistsService.On("Do", mock.Anything).Return(true, nil)
	mockClient.On("IndexExists", []string{indexName}).Return(mockExistsService)

	err := elkClient.CreateIndex(indexName)
	assert.NoError(t, err)

	mockClient.AssertExpectations(t)
	mockExistsService.AssertExpectations(t)
}

func TestCreateIndexError(t *testing.T) {
	mockClient := new(MockClient)
	elkClient := elk.NewTestElkClient(mockClient)

	indexName := "error-index"

	mockExistsService := new(MockExistsService)
	mockExistsService.On("Do", mock.Anything).Return(false, fmt.Errorf("mock error"))
	mockClient.On("IndexExists", []string{indexName}).Return(mockExistsService)

	err := elkClient.CreateIndex(indexName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check index existence")

	mockClient.AssertExpectations(t)
	mockExistsService.AssertExpectations(t)
}

// TestCreateIndexWithCreateError 测试创建索引时发生错误的情况
func TestCreateIndexWithCreateError(t *testing.T) {
	mockClient := new(MockClient)
	elkClient := elk.NewTestElkClient(mockClient)

	indexName := "error-create-index"
	fmt.Printf("Testing index creation error: %s\n", indexName)

	// 设置 IndexExists 的期望
	mockExistsService := new(MockExistsService)
	mockClient.On("IndexExists", []string{indexName}).Return(mockExistsService)
	mockExistsService.On("Do", mock.Anything).Return(false, nil)

	// 设置 CreateIndex 的期望
	mockCreateService := NewMockCreateService(indexName)
	mockClient.On("CreateIndex", indexName).Return(mockCreateService)

	// 设置 BodyJson 的期望
	mockCreateService.On("BodyJson", mock.Anything).Return(mockCreateService)

	// 设置 Do 方法返回错误
	expectedError := fmt.Errorf("failed to create index: invalid settings")
	mockCreateService.On("Do", mock.Anything).Return(nil, expectedError)

	err := elkClient.CreateIndex(indexName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create index")

	mockClient.AssertExpectations(t)
	mockExistsService.AssertExpectations(t)
	mockCreateService.AssertExpectations(t)
}

// TestCreateIndexWithInvalidSettings 测试无效的索引设置
func TestCreateIndexWithInvalidSettings(t *testing.T) {
	mockClient := new(MockClient)
	config := elk.DefaultElasticConfig()
	config.NumberOfShards = 0 // 无效的分片数

	elkClient := elk.NewTestElkClient(mockClient)
	elkClient.SetConfig(config)

	indexName := "invalid-settings-index"
	fmt.Printf("Testing invalid settings: %s\n", indexName)

	// 由于配置无效，不应该调用任何 mock 方法
	// 直接验证返回的错误
	err := elkClient.CreateIndex(indexName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid settings")

	// 验证没有调用任何 mock 方法
	mockClient.AssertNotCalled(t, "IndexExists")
	mockClient.AssertNotCalled(t, "CreateIndex")
}

// TestCreateIndexWithNilClient 测试客户端为空的情况
func TestCreateIndexWithNilClient(t *testing.T) {
	elkClient := elk.NewTestElkClient(nil)

	indexName := "nil-client-index"
	err := elkClient.CreateIndex(indexName)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "client is not initialized")
}
