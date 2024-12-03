package logrus

import (
	"fmt"
	"runtime"
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
	config   RecoveryConfig         // 配置
	writer   Writer                 // 写入器
	mu       sync.Mutex             // 互斥锁
	retryMap map[string]*retryState // 记录每个日志的重试状态
}

type retryState struct {
	retries   int       // 重试次数
	lastRetry time.Time // 最后一次重试时间
	content   []byte    // 日志内容
}

// NewRecoveryWriter 创建错误恢复写入器
func NewRecoveryWriter(w Writer, config RecoveryConfig) *RecoveryWriter {
	// 验证配置
	if config.PanicHandler == nil {
		config.PanicHandler = defaultPanicHandler
	}
	// 创建错误恢复写入器
	return &RecoveryWriter{
		config:   config,                       // 配置
		writer:   w,                            // 写入器
		retryMap: make(map[string]*retryState), // 重试状态映射
	}
}

// Write 实现 io.Writer 接口
func (w *RecoveryWriter) Write(p []byte) (n int, err error) {
	// 检查是否启用错误恢复
	if !w.config.Enable {
		return w.writer.Write(p)
	}

	// 使用 defer 恢复 panic
	defer func() {
		// 恢复 panic
		if r := recover(); r != nil {
			stack := make([]byte, w.config.MaxStackSize)    // 创建堆栈
			stack = stack[:runtime.Stack(stack, false)]     // 获取堆栈
			w.config.PanicHandler(r, stack)                 // 调用 panic 处理函数
			err = fmt.Errorf("recovered from panic: %v", r) // 设置错误
		}
	}()

	// 尝试写入
	n, err = w.writer.Write(p)

	// 如果写入失败，处理错误
	if err != nil {
		return w.handleWriteError(p)
	}

	return n, nil
}

// handleWriteError 处理写入错误
func (w *RecoveryWriter) handleWriteError(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	key := string(p)                 // 使用日志内容作为key
	state, exists := w.retryMap[key] // 获取重试状态

	// 如果重试状态不存在，创建新的重试状态
	if !exists {
		state = &retryState{content: p} // 创建新的重试状态
		w.retryMap[key] = state
	}

	// 检查是否超过最大重试次数
	if state.retries >= w.config.MaxRetries {
		delete(w.retryMap, key)
		return 0, fmt.Errorf("max retries exceeded")
	}

	// 检查重试间隔
	if time.Since(state.lastRetry) < w.config.RetryInterval {
		return 0, fmt.Errorf("retry too frequent")
	}

	// 增加重试次数并更新时间
	state.retries++
	state.lastRetry = time.Now()

	// 异步重试
	go func() {
		time.Sleep(w.config.RetryInterval)
		if _, err := w.writer.Write(p); err == nil {
			w.mu.Lock()
			delete(w.retryMap, key)
			w.mu.Unlock()
		}
	}()

	return len(p), nil
}

// defaultPanicHandler 默认panic处理函数
func defaultPanicHandler(r interface{}, stack []byte) {
	fmt.Printf("Recovered from panic: %v\nStack trace:\n%s\n", r, stack)
}

// CleanupRetries 清理过期的重试记录
func (w *RecoveryWriter) CleanupRetries() {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 遍历重试记录
	for key, state := range w.retryMap {
		// 检查是否超过重试间隔
		if time.Since(state.lastRetry) > w.config.RetryInterval*time.Duration(w.config.MaxRetries) {
			delete(w.retryMap, key)
		}
	}
}
