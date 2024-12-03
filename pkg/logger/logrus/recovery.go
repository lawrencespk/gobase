package logrus

import (
	"fmt"
	"io"
	"runtime/debug"
	"strings"
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
	if !w.config.Enable { // 如果未启用错误恢复，则直接写入
		return w.writer.Write(p) // 直接写入
	}

	// 复制数据以防后续修改
	data := make([]byte, len(p))
	copy(data, p) // 复制数据

	// 包装首次写入，处理可能的 panic
	var firstWriteErr error
	func() {
		defer func() {
			if r := recover(); r != nil { // 捕获 panic
				if w.config.PanicHandler != nil {
					w.config.PanicHandler(r, debug.Stack()) // 调用 panic 处理函数
				}
				firstWriteErr = fmt.Errorf("panic in write: %v", r) // 设置错误
			}
		}()
		n, firstWriteErr = w.writer.Write(data) // 写入数据
	}()

	// 如果首次写入成功，直接返回
	if firstWriteErr == nil {
		return n, nil
	}

	// 如果是 panic 导致的错误，直接返回错误
	if strings.Contains(firstWriteErr.Error(), "panic in write") {
		return 0, firstWriteErr
	}

	// 启动重试协程
	go func() {
		defer func() {
			if r := recover(); r != nil { // 捕获 panic
				if w.config.PanicHandler != nil {
					w.config.PanicHandler(r, debug.Stack()) // 调用 panic 处理函数
				}
			}
			// 无论成功失败，都通知重试完成
			close(w.retryDone) // 通知重试完成
		}()

		for i := 0; i < w.config.MaxRetries; i++ {
			time.Sleep(w.config.RetryInterval)
			if _, err := w.writer.Write(data); err == nil { // 写入数据
				return // 如果写入成功，返回
			}
		}
	}()

	// 返回成功，因为重试是异步的
	return len(p), nil
}

// WaitForRetries 等待所有重试完成
func (w *RecoveryWriter) WaitForRetries() {
	if w.retryDone != nil {
		<-w.retryDone                     // 等待重试完成
		w.retryDone = make(chan struct{}) // 重置 channel 以供下次使用
	}
}

func (w *RecoveryWriter) CleanupRetries() {
	// 等待所有重试完成
	w.WaitForRetries()
}
