package unit

import (
	"testing"

	"gobase/pkg/logger/elk/tests/mock"
	"gobase/pkg/logger/logrus"

	slogrus "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TestNewHook 测试Hook的创建
func TestNewHook(t *testing.T) {
	mockClient := mock.NewMockElkClient()
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
			mockClient := mock.NewMockElkClient()
			h := logrus.NewHookWithClient(mockClient)

			// 假设我们在 Hook 中有一个��法可以设置期望的级别
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
	mockClient := mock.NewMockElkClient()
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
	mockClient := mock.NewMockElkClient()
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
	mockClient := mock.NewMockElkClient()
	mockClient.SetShouldFailOps(true)
	h := logrus.NewHookWithClient(mockClient)

	entry := &slogrus.Entry{
		Logger:  slogrus.New(),
		Level:   slogrus.InfoLevel,
		Message: "test message",
	}

	err := h.Fire(entry)
	assert.Error(t, err)
}

// TestHook 测试完整的Hook功能
func TestHook(t *testing.T) {
	mockClient := mock.NewMockElkClient()
	hook := logrus.NewHookWithClient(mockClient)

	// 测试Hook基本属性
	assert.NotNil(t, hook)
	assert.NotEmpty(t, hook.Levels())

	// 测试Fire功能
	entry := &slogrus.Entry{
		Logger:  slogrus.New(),
		Level:   slogrus.InfoLevel,
		Message: "test complete hook functionality",
		Data: slogrus.Fields{
			"test_key": "test_value",
		},
	}

	err := hook.Fire(entry)
	assert.NoError(t, err)

	// 验证文档是否被正确索引
	docs := mockClient.GetDocuments()
	assert.NotEmpty(t, docs)
}
