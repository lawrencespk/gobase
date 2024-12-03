package unit

import (
	"errors"
	"testing"
	"time"

	"gobase/pkg/logger/logrus"

	"github.com/stretchr/testify/assert"
)

// 扩展现有的 MockWriter 结构体
func (w *MockWriter) SetShouldFail(fail bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.err = errors.New("模拟的写入失败")
	w.failNow = fail
}

func (w *MockWriter) SetPanic(panic bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.shouldPanic = panic
}

func TestRecoveryWriter_Write(t *testing.T) {
	t.Run("正常写入测试", func(t *testing.T) {
		mockWriter := NewMockWriter()
		config := logrus.RecoveryConfig{
			Enable:        true,
			MaxRetries:    3,
			RetryInterval: time.Millisecond * 10,
		}

		recoveryWriter := logrus.NewRecoveryWriter(mockWriter, config)
		testData := []byte("测试日志")

		n, err := recoveryWriter.Write(testData)

		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)
		assert.Equal(t, testData, mockWriter.GetWritten())
	})

	t.Run("写入失败重试测试", func(t *testing.T) {
		mockWriter := NewMockWriter()
		mockWriter.SetShouldFail(true)
		config := logrus.RecoveryConfig{
			Enable:        true,
			MaxRetries:    3,
			RetryInterval: time.Millisecond * 10,
		}

		recoveryWriter := logrus.NewRecoveryWriter(mockWriter, config)
		testData := []byte("测试日志")

		_, err := recoveryWriter.Write(testData)

		// 首次写入应该返回错误，但不会阻塞
		assert.NoError(t, err)

		// 等待所有重试完成
		recoveryWriter.WaitForRetries()

		// 验证重试次数
		assert.Equal(t, 4, mockWriter.writeCount) // 1次初始 + 3次重试
	})

	t.Run("Panic恢复测试", func(t *testing.T) {
		var panicCalled bool
		mockWriter := NewMockWriter()
		mockWriter.SetPanic(true)
		config := logrus.RecoveryConfig{
			Enable:        true,
			MaxRetries:    3,
			RetryInterval: time.Millisecond * 10,
			PanicHandler: func(r interface{}, stack []byte) {
				panicCalled = true
			},
		}

		recoveryWriter := logrus.NewRecoveryWriter(mockWriter, config)
		testData := []byte("测试日志")

		_, err := recoveryWriter.Write(testData)

		assert.Error(t, err)
		assert.True(t, panicCalled)
	})

	t.Run("禁用恢复功能测试", func(t *testing.T) {
		mockWriter := NewMockWriter()
		mockWriter.SetShouldFail(true)
		config := logrus.RecoveryConfig{
			Enable: false,
		}

		recoveryWriter := logrus.NewRecoveryWriter(mockWriter, config)
		testData := []byte("测试日志")

		_, err := recoveryWriter.Write(testData)

		assert.Error(t, err)
		assert.Equal(t, 1, mockWriter.writeCount)
	})
}

func TestRecoveryWriter_CleanupRetries(t *testing.T) {
	mockWriter := NewMockWriter()
	mockWriter.SetShouldFail(true)
	config := logrus.RecoveryConfig{
		Enable:        true,
		MaxRetries:    3,
		RetryInterval: time.Millisecond * 10,
	}

	recoveryWriter := logrus.NewRecoveryWriter(mockWriter, config)
	testData := []byte("测试日志")

	// 触发一次写入失败
	recoveryWriter.Write(testData)

	// 等待足够长的时间
	time.Sleep(time.Millisecond * 50)

	// 清理重试记录
	recoveryWriter.CleanupRetries()
}
