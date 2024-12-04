package logrus

import (
	"fmt"
	"io"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

// RecoveryConfig 错误恢复配置
type RecoveryConfig struct {
	Enable           bool                      // 是否启用错误恢复
	MaxRetries       int                       // 最大重试次数
	RetryInterval    time.Duration             // 重试间隔
	PanicHandler     func(interface{}, []byte) // panic处理函数
	EnableStackTrace bool                      // 是否启用堆栈跟踪
	MaxStackSize     int                       // 最大堆栈大小
}

// RecoveryWriter 错误恢复写入器
type RecoveryWriter struct {
	writer    io.Writer      // 写入器
	config    RecoveryConfig // 配置
	retryDone chan struct{}  // 用于通知重试完成
	wg        sync.WaitGroup // 添加等待组
	mu        sync.Mutex     // 添加互斥锁
	retrying  bool           // 用于标记是否正在重试
}

// NewRecoveryWriter 创建错误恢复写入器
func NewRecoveryWriter(writer io.Writer, config RecoveryConfig) *RecoveryWriter {
	return &RecoveryWriter{
		writer:    writer,              // 写入器
		config:    config,              // 配置
		retryDone: make(chan struct{}), // 用于通知重试完成
	}
}

// Write 实现 io.Writer 接口
func (w *RecoveryWriter) Write(p []byte) (n int, err error) {
	if !w.config.Enable {
		return w.writer.Write(p)
	}

	// 包装首次写入以处理panic
	func() {
		defer func() {
			if r := recover(); r != nil {
				if w.config.PanicHandler != nil {
					w.config.PanicHandler(r, debug.Stack())
				}
				err = fmt.Errorf("panic during write: %v", r)
			}
		}()
		n, err = w.writer.Write(p)
	}()

	// 如果首次写入成功，直接返回
	if err == nil {
		return n, nil
	}

	// 如果是panic导致的错误，不进行重试
	if err != nil && strings.Contains(err.Error(), "panic during write") {
		return 0, err
	}

	// 准备重试数据
	data := make([]byte, len(p))
	copy(data, p)

	w.mu.Lock()
	w.retrying = true
	w.mu.Unlock()

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer func() {
			w.mu.Lock()
			w.retrying = false
			w.mu.Unlock()

			if r := recover(); r != nil {
				if w.config.PanicHandler != nil {
					w.config.PanicHandler(r, debug.Stack())
				}
			}
		}()

		// 重试写入
		for i := 0; i < w.config.MaxRetries; i++ {
			if _, err := w.writer.Write(data); err == nil {
				return
			}
			time.Sleep(w.config.RetryInterval)
		}
	}()

	return len(p), nil
}

// WaitForRetries 等待所有重试完成
func (w *RecoveryWriter) WaitForRetries() {
	w.wg.Wait()
}

func (w *RecoveryWriter) CleanupRetries() {
	w.WaitForRetries()
}

// 实现 io.Closer 接口
func (w *RecoveryWriter) Close() error {
	w.mu.Lock()
	if w.retrying { // 如果正在重试
		w.mu.Unlock()
		w.WaitForRetries() // 等待重试完成
		return nil
	}
	w.mu.Unlock()

	if closer, ok := w.writer.(io.Closer); ok { // 如果writer实现了io.Closer接口
		return closer.Close() // 关闭writer
	}
	return nil
}
