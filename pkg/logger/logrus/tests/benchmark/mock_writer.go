package benchmark

import (
	"sync"
)

// MockWriter 用于测试的模拟写入器
type MockWriter struct {
	mu      sync.Mutex
	written []byte
}

// NewMockWriter 创建一个新的 MockWriter 实例
func NewMockWriter() *MockWriter {
	return &MockWriter{
		written: make([]byte, 0),
	}
}

// Write 实现 io.Writer 接口
func (w *MockWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.written = append(w.written, p...)
	return len(p), nil
}
