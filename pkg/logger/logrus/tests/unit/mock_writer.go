package unit

import (
	"errors"
	"sync"
	"time"
)

// MockWriter 用于测试的模拟写入器
type MockWriter struct {
	mu          sync.Mutex
	written     []byte
	err         error
	failNow     bool
	shouldPanic bool
	writeCount  int
	delay       time.Duration
	appendMode  bool
}

// NewMockWriter 创建一个新的 MockWriter 实例
func NewMockWriter() *MockWriter {
	return &MockWriter{
		appendMode: true,
	}
}

// SetAppendMode 设置写入模式
func (w *MockWriter) SetAppendMode(append bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.appendMode = append
}

func (w *MockWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.writeCount++

	if w.delay > 0 {
		time.Sleep(w.delay)
	}

	if w.shouldPanic {
		panic("模拟的写入panic")
	}

	if w.failNow {
		return 0, w.err
	}

	if w.appendMode {
		// 追加模式
		w.written = append(w.written, p...)
	} else {
		// 覆盖模式
		w.written = make([]byte, len(p))
		copy(w.written, p)
	}

	return len(p), nil
}

func (w *MockWriter) GetWritten() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.written
}

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

// SetDelay 设置写入延迟
func (w *MockWriter) SetDelay(delay time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.delay = delay
}

// GetWriteCount 获取写入次数
func (w *MockWriter) GetWriteCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.writeCount
}

// Reset 重置 MockWriter 的状态
func (w *MockWriter) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.written = nil
	w.err = nil
	w.failNow = false
	w.shouldPanic = false
	w.writeCount = 0
	w.delay = 0
}
