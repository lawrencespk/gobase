package unit

import (
	"bytes"
	"sync"
	"testing"
	"time"

	"gobase/pkg/logger/logrus"
)

// MockWriter 用于测试的模拟写入器
type MockWriter struct {
	mu      sync.Mutex
	written []byte
	delay   time.Duration // 模拟写入延迟
	err     error         // 模拟写入错误
}

func NewMockWriter() *MockWriter {
	return &MockWriter{}
}

func (w *MockWriter) Write(p []byte) (n int, err error) {
	if w.delay > 0 {
		time.Sleep(w.delay)
	}
	if w.err != nil {
		return 0, w.err
	}
	w.mu.Lock()
	w.written = append(w.written, p...)
	w.mu.Unlock()
	return len(p), nil
}

func (w *MockWriter) GetWritten() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.written
}

// 测试AsyncWriter
func TestAsyncWriter(t *testing.T) {
	t.Run("Basic Write Operations", func(t *testing.T) {
		mock := NewMockWriter()
		config := logrus.AsyncConfig{
			Enable:        true,
			BufferSize:    1024,
			FlushInterval: time.Millisecond * 100,
			BlockOnFull:   false,
			DropOnFull:    false,
			FlushOnExit:   true,
		}

		writer := logrus.NewAsyncWriter(mock, config)
		defer writer.Stop()

		testData := []byte("test data")
		n, err := writer.Write(testData)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if n != len(testData) {
			t.Errorf("Expected write length %d, got %d", len(testData), n)
		}

		// 等待异步写入完成
		time.Sleep(config.FlushInterval * 2)

		if !bytes.Equal(mock.GetWritten(), testData) {
			t.Errorf("Expected %v, got %v", testData, mock.GetWritten())
		}
	})

	t.Run("Buffer Full Behavior", func(t *testing.T) {
		mock := NewMockWriter()
		mock.delay = time.Millisecond * 10 // 添加写入延迟

		config := logrus.AsyncConfig{
			Enable:        true,
			BufferSize:    1, // 很小的缓冲区以测试满载情况
			FlushInterval: time.Second,
			BlockOnFull:   false,
			DropOnFull:    true,
			FlushOnExit:   true,
		}

		writer := logrus.NewAsyncWriter(mock, config)
		defer writer.Stop()

		// 快速写入多条数据
		for i := 0; i < 100; i++ {
			writer.Write([]byte("test"))
		}

		// 检查是否有日志被丢弃
		if writer.GetDropCount() == 0 {
			t.Error("Expected some logs to be dropped")
		}
	})

	t.Run("Flush On Exit", func(t *testing.T) {
		mock := NewMockWriter()
		config := logrus.AsyncConfig{
			Enable:        true,
			BufferSize:    1024,
			FlushInterval: time.Second,
			BlockOnFull:   false,
			DropOnFull:    false,
			FlushOnExit:   true,
		}

		writer := logrus.NewAsyncWriter(mock, config)
		testData := []byte("test data")
		writer.Write(testData)

		// 立即停止并检查数据是否被刷新
		err := writer.Stop()
		if err != nil {
			t.Errorf("Expected no error on stop, got %v", err)
		}

		if !bytes.Equal(mock.GetWritten(), testData) {
			t.Errorf("Expected data to be flushed on exit")
		}
	})

	t.Run("Concurrent Writes", func(t *testing.T) {
		mock := NewMockWriter()
		config := logrus.AsyncConfig{
			Enable:        true,
			BufferSize:    1024,
			FlushInterval: time.Millisecond * 100,
			BlockOnFull:   true,
			DropOnFull:    false,
			FlushOnExit:   true,
		}

		writer := logrus.NewAsyncWriter(mock, config)
		defer writer.Stop()

		var wg sync.WaitGroup
		workers := 10
		iterations := 100

		wg.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					writer.Write([]byte("test"))
				}
			}()
		}
		wg.Wait()

		// 等待所有数据写入
		time.Sleep(config.FlushInterval * 2)

		expectedLen := workers * iterations * 4 // "test" 的长度是 4
		if len(mock.GetWritten()) != expectedLen {
			t.Errorf("Expected total written length %d, got %d", expectedLen, len(mock.GetWritten()))
		}
	})
}

// 测试AsyncWriter的性能
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
