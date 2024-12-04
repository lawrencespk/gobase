package unit

import (
	"testing"
	"time"

	"gobase/pkg/logger/logrus"

	"github.com/stretchr/testify/assert"
)

// TestRecoveryWriter_Write 测试恢复写入器的写入功能
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
		recoveryWriter.WaitForRetries() // 等待所有重试完成

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
		recoveryWriter.WaitForRetries()

		assert.NoError(t, err)
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
		recoveryWriter.WaitForRetries()

		assert.Error(t, err)
		assert.True(t, panicCalled)
	})

	t.Run("禁用恢复功能测试", func(t *testing.T) {
		mockWriter := NewMockWriter()
		config := logrus.RecoveryConfig{
			Enable:        false,
			MaxRetries:    3,
			RetryInterval: time.Millisecond * 10,
		}

		recoveryWriter := logrus.NewRecoveryWriter(mockWriter, config)
		testData := []byte("测试日志")

		n, err := recoveryWriter.Write(testData)
		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)
		assert.Equal(t, testData, mockWriter.GetWritten())
		assert.Equal(t, 1, mockWriter.writeCount) // 只写入一次
	})
}

// TestRecoveryWriter_CleanupRetries 测试清理重试记录
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
