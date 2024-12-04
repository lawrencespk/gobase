package unit

import (
	"context"
	"errors"
	"testing"

	"gobase/pkg/logger/elk"
	"gobase/pkg/logger/logrus"

	slogrus "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// 创建一个模拟的 ElkClient
type mockElkClient struct {
	shouldError bool
}

func (m *mockElkClient) Connect(config *elk.ElkConfig) error {
	return nil
}

func (m *mockElkClient) Close() error {
	return nil
}

func (m *mockElkClient) IndexDocument(ctx context.Context, index string, document interface{}) error {
	if m.shouldError {
		return errors.New("mock error")
	}
	return nil
}

// 添加缺少的方法
func (m *mockElkClient) BulkIndexDocuments(ctx context.Context, index string, documents []interface{}) error {
	if m.shouldError {
		return errors.New("mock error")
	}
	return nil
}

func (m *mockElkClient) Query(ctx context.Context, index string, query interface{}) (interface{}, error) {
	if m.shouldError {
		return nil, errors.New("mock error")
	}
	return nil, nil
}

func (m *mockElkClient) IsConnected() bool {
	return true
}

// TestNewHook 测试Hook的创建
func TestNewHook(t *testing.T) {
	// 使用模拟的客户端
	mockClient := &mockElkClient{}
	hook := logrus.NewHookWithClient(mockClient)
	assert.NotNil(t, hook)
}

// TestHookLevels 测试Hook的日志级别过滤
func TestHookLevels(t *testing.T) {
	tests := []struct {
		name   string
		levels []slogrus.Level
	}{
		{
			name:   "单个级别",
			levels: []slogrus.Level{slogrus.InfoLevel},
		},
		{
			name:   "多个级别",
			levels: []slogrus.Level{slogrus.InfoLevel, slogrus.WarnLevel, slogrus.ErrorLevel},
		},
		{
			name:   "所有级别",
			levels: slogrus.AllLevels,
		},
		{
			name:   "无级别",
			levels: []slogrus.Level{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockElkClient{}
			h := logrus.NewHookWithClient(mockClient)

			// 假设我们在 Hook 中有一个方法可以设置期望的级别
			h.SetLevels(tt.levels) // 需要在 Hook 中实现这个方法

			levels := h.Levels()

			assert.Equal(t, len(tt.levels), len(levels), "日志级别数量不匹配")
			for i, level := range tt.levels {
				assert.Equal(t, level, levels[i], "日志级别不匹配")
			}
		})
	}
}

// TestHookFire 测试Hook的Fire方法
func TestHookFire(t *testing.T) {
	mockClient := &mockElkClient{}
	h := logrus.NewHookWithClient(mockClient)

	entry := &slogrus.Entry{
		Logger:  slogrus.New(),
		Level:   slogrus.InfoLevel,
		Message: "test message",
	}

	err := h.Fire(entry)
	assert.NoError(t, err)
}

// TestHookFireWithFields 测试带字段的Hook Fire
func TestHookFireWithFields(t *testing.T) {
	mockClient := &mockElkClient{}
	h := logrus.NewHookWithClient(mockClient)

	entry := &slogrus.Entry{
		Logger: slogrus.New(),
		Level:  slogrus.InfoLevel,
		Data: slogrus.Fields{
			"key": "value",
		},
		Message: "test message with fields",
	}

	err := h.Fire(entry)
	assert.NoError(t, err)
}

// TestHookFireError 测试Hook Fire的错误处理
func TestHookFireError(t *testing.T) {
	// 创建一个总是返回错误的模拟客户端
	errorClient := &mockElkClient{shouldError: true}
	h := logrus.NewHookWithClient(errorClient)

	entry := &slogrus.Entry{
		Logger:  slogrus.New(),
		Level:   slogrus.InfoLevel,
		Message: "test message",
	}

	err := h.Fire(entry)
	assert.Error(t, err)
}
