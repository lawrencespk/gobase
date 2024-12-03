package unit

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"gobase/pkg/logger/logrus"
)

// ErrorWriter 用于测试错误情况的写入器
type ErrorWriter struct {
	mu       sync.Mutex
	written  []byte
	failNext bool // 下一次写入是否失败
}

// NewErrorWriter 创建一个用于测试错误情况的写入器
func NewErrorWriter() *ErrorWriter {
	return &ErrorWriter{}
}

// Write 实现 logrus.Writer 接口
func (w *ErrorWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.failNext {
		w.failNext = false
		return 0, errors.New("simulated write error")
	}

	w.written = append(w.written, p...)
	return len(p), nil
}

// GetWritten 获取已写入的数据
func (w *ErrorWriter) GetWritten() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.written
}

// SetFailNext 设置下一次写入失败
func (w *ErrorWriter) SetFailNext() {
	w.mu.Lock()
	w.failNext = true
	w.mu.Unlock()
}

// TestWriteQueue 测试写入队列
func TestWriteQueue(t *testing.T) {
	t.Run("Basic Queue Operations", func(t *testing.T) {
		writer := NewMockWriter()
		config := logrus.QueueConfig{
			MaxSize:       1000,
			BatchSize:     10,
			FlushInterval: time.Millisecond * 100,
			Workers:       1,
		}

		queue, err := logrus.NewWriteQueue(writer, config)
		if err != nil {
			t.Fatalf("Failed to create write queue: %v", err)
		}
		defer queue.Close(context.Background())

		testData := []byte("test data")
		n, err := queue.Write(testData)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if n != len(testData) {
			t.Errorf("Expected write length %d, got %d", len(testData), n)
		}

		// 等待数据被处理
		time.Sleep(config.FlushInterval * 2)

		if !bytes.Equal(writer.GetWritten(), testData) {
			t.Errorf("Expected %v, got %v", testData, writer.GetWritten())
		}
	})

	t.Run("Queue_Full_Behavior", func(t *testing.T) {
		writer := NewMockWriter()
		writer.delay = time.Millisecond * 10

		config := logrus.QueueConfig{
			MaxSize:       1,
			BatchSize:     1,
			FlushInterval: time.Second,
			Workers:       1,
		}

		queue, err := logrus.NewWriteQueue(writer, config)
		if err != nil {
			t.Fatalf("Failed to create write queue: %v", err)
		}
		defer queue.Close(context.Background())

		var writeErrors int
		for i := 0; i < 100; i++ {
			_, err := queue.Write([]byte("test"))
			if err != nil {
				if !logrus.IsQueueFullError(err) {
					t.Errorf("Expected queue full error, got %v", err)
				}
				writeErrors++
			}
		}

		if writeErrors == 0 {
			t.Error("Expected some writes to fail with ErrQueueFull")
		}
	})

	t.Run("Batch Processing", func(t *testing.T) {
		writer := NewMockWriter()
		config := logrus.QueueConfig{
			MaxSize:       1000,
			BatchSize:     10,
			FlushInterval: time.Millisecond * 100,
			Workers:       1,
		}

		queue, err := logrus.NewWriteQueue(writer, config)
		if err != nil {
			t.Fatalf("Failed to create write queue: %v", err)
		}
		defer queue.Close(context.Background())

		// 写入多条数据
		for i := 0; i < 50; i++ {
			queue.Write([]byte("test"))
		}

		// 等待批处理完成
		time.Sleep(config.FlushInterval * 2)

		written := writer.GetWritten()
		expectedLen := 50 * 4 // "test" 的长度是 4
		if len(written) != expectedLen {
			t.Errorf("Expected total written length %d, got %d", expectedLen, len(written))
		}
	})

	t.Run("Error Recovery", func(t *testing.T) {
		writer := NewErrorWriter()
		config := logrus.QueueConfig{
			MaxSize:       1000,
			BatchSize:     10,
			FlushInterval: time.Millisecond * 100,
			Workers:       1,
		}

		queue, err := logrus.NewWriteQueue(writer, config)
		if err != nil {
			t.Fatalf("Failed to create write queue: %v", err)
		}
		defer queue.Close(context.Background())

		// 设置下一次写入失败
		writer.SetFailNext()

		testData := []byte("test data")
		queue.Write(testData)

		// 等待重试
		time.Sleep(config.FlushInterval * 4)

		// 验证数据最终被写入
		if !bytes.Equal(writer.GetWritten(), testData) {
			t.Error("Expected data to be written after recovery")
		}
	})

	t.Run("Graceful Shutdown", func(t *testing.T) {
		writer := NewMockWriter()
		config := logrus.QueueConfig{
			MaxSize:       1000,
			BatchSize:     10,
			FlushInterval: time.Second,
			Workers:       2,
		}

		queue, err := logrus.NewWriteQueue(writer, config)
		if err != nil {
			t.Fatalf("Failed to create write queue: %v", err)
		}

		// 写入一些数据
		testData := []byte("shutdown test")
		queue.Write(testData)

		// 优雅关闭
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err = queue.Close(ctx)
		if err != nil {
			t.Errorf("Expected no error during shutdown, got %v", err)
		}

		// 验证数据被完全写入
		if !bytes.Equal(writer.GetWritten(), testData) {
			t.Error("Expected all data to be written before shutdown")
		}
	})
}

// BenchmarkWriteQueue 测试写入队列的性能
func BenchmarkWriteQueue(b *testing.B) {
	writer := NewMockWriter()
	config := logrus.QueueConfig{
		MaxSize:       1000000,
		BatchSize:     1000,
		FlushInterval: time.Millisecond * 100,
		Workers:       4,
	}

	queue, err := logrus.NewWriteQueue(writer, config)
	if err != nil {
		b.Fatalf("Failed to create write queue: %v", err)
	}
	defer queue.Close(context.Background())

	data := []byte("benchmark test data")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			queue.Write(data)
		}
	})
}
