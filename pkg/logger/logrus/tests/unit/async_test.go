package unit

import (
	"testing"
	"time"

	"gobase/pkg/logger/logrus"

	"github.com/stretchr/testify/assert"
)

func TestAsyncWriter(t *testing.T) {
	t.Run("正常写入测试", func(t *testing.T) {
		mock := NewMockWriter()
		config := logrus.AsyncConfig{
			Enable:        true,
			BufferSize:    1024,
			FlushInterval: time.Millisecond * 100,
			BlockOnFull:   true,

			FlushOnExit: true,
		}

		writer := logrus.NewAsyncWriter(mock, config)
		testData := []byte("测试数据")

		n, err := writer.Write(testData)
		assert.NoError(t, err)
		assert.Equal(t, len(testData), n)

		// 等待异步写入完成
		time.Sleep(config.FlushInterval * 2)
		assert.Equal(t, testData, mock.GetWritten())
	})

	t.Run("缓冲区满时阻塞测试", func(t *testing.T) {
		mock := NewMockWriter()
		config := logrus.AsyncConfig{
			Enable:        true,
			BufferSize:    10, // 设置较小的缓冲区
			FlushInterval: time.Second,
			BlockOnFull:   true,
			FlushOnExit:   true,
		}

		writer := logrus.NewAsyncWriter(mock, config)
		testData := []byte("测试数据")

		done := make(chan bool)
		go func() {
			for i := 0; i < 100; i++ {
				_, err := writer.Write(testData)
				assert.NoError(t, err)
			}
			done <- true
		}()

		select {
		case <-done:
			t.Log("写入完成")
		case <-time.After(time.Second * 2):
			t.Log("写入被阻塞")
		}
	})

	t.Run("缓冲区满时丢弃测试", func(t *testing.T) {
		mock := NewMockWriter()
		config := logrus.AsyncConfig{
			Enable:        true,
			BufferSize:    10,
			FlushInterval: time.Second,
			BlockOnFull:   false,
			DropOnFull:    true,
			FlushOnExit:   true,
		}

		writer := logrus.NewAsyncWriter(mock, config)
		testData := []byte("测试数据")

		for i := 0; i < 100; i++ {
			n, err := writer.Write(testData)
			if err != nil {
				assert.Equal(t, "buffer full", err.Error())
				break
			}
			assert.Equal(t, len(testData), n)
		}
	})

	t.Run("写入错误处理测试", func(t *testing.T) {
		mock := NewMockWriter()
		mock.SetShouldFail(true)
		config := logrus.AsyncConfig{
			Enable:        true,
			BufferSize:    1024,
			FlushInterval: time.Millisecond * 100,
			BlockOnFull:   true,
			FlushOnExit:   true,
		}

		writer := logrus.NewAsyncWriter(mock, config)
		testData := []byte("测试数据")

		_, err := writer.Write(testData)
		assert.NoError(t, err) // 异步写入应该不会立即返回错误

		// 等待错误处理
		time.Sleep(config.FlushInterval * 2)
	})

	t.Run("优雅关闭测试", func(t *testing.T) {
		mock := NewMockWriter()
		mock.SetAppendMode(false)
		config := logrus.AsyncConfig{
			Enable:        true,
			BufferSize:    1024,
			FlushInterval: time.Millisecond * 100,
			BlockOnFull:   true,
			FlushOnExit:   true,
		}

		writer := logrus.NewAsyncWriter(mock, config)
		testData := []byte("测试数据")

		for i := 0; i < 10; i++ {
			_, err := writer.Write(testData)
			assert.NoError(t, err)
		}

		err := writer.Stop()
		assert.NoError(t, err)
		assert.Equal(t, testData, mock.GetWritten())
	})
}

func BenchmarkAsyncWriter(b *testing.B) {
	mock := NewMockWriter()
	config := logrus.AsyncConfig{

		Enable:        true,
		BufferSize:    1024 * 1024, // 1MB buffer
		FlushInterval: time.Millisecond * 100,
		BlockOnFull:   false,
		DropOnFull:    true,
		FlushOnExit:   true,
	}

	writer := logrus.NewAsyncWriter(mock, config)
	defer writer.Stop()

	data := []byte("benchmark test data")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			writer.Write(data)
		}
	})
}

func BenchmarkAsyncWriter_LargeData(b *testing.B) {
	mock := NewMockWriter()
	config := logrus.AsyncConfig{
		Enable:        true,
		BufferSize:    1024 * 1024 * 10, // 10MB buffer
		FlushInterval: time.Millisecond * 100,
		BlockOnFull:   true,
		FlushOnExit:   true,
	}

	writer := logrus.NewAsyncWriter(mock, config)
	defer writer.Stop()

	data := make([]byte, 1024*1024) // 1MB data
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			writer.Write(data)
		}
	})
}
