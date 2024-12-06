package logrus

import (
	"sync"
	"sync/atomic"
	"time"

	"gobase/pkg/errors"
)

// AsyncConfig 异步写入配置
type AsyncConfig struct {
	Enable        bool          // 是否启用异步写入
	BufferSize    int           // 缓冲区大小
	FlushInterval time.Duration // 定期刷新间隔
	BlockOnFull   bool          // 缓冲区满时是否阻塞
	DropOnFull    bool          // 缓冲区满时是否丢弃
	FlushOnExit   bool          // 退出时是否刷新缓冲区
}

// AsyncWriter 异步写入器
type AsyncWriter struct {
	config    AsyncConfig    // 配置
	buffer    chan []byte    // 缓冲区
	done      chan struct{}  // 退出通道
	wg        sync.WaitGroup // 等待组
	writer    Writer         // 写入器
	dropCount int64          // 丢弃的日志计数
}

// NewAsyncWriter 创建异步写入器
func NewAsyncWriter(w Writer, config AsyncConfig) *AsyncWriter {
	aw := &AsyncWriter{
		config: config,                               // 配置
		buffer: make(chan []byte, config.BufferSize), // 缓冲区
		done:   make(chan struct{}),                  // 退出通道
		writer: w,                                    // 写入器
	}

	if config.Enable {
		aw.start() // 启动异步写入
	}

	return aw
}

// Write 实现 io.Writer 接口
func (w *AsyncWriter) Write(p []byte) (n int, err error) {
	if !w.config.Enable { // 如果未启用异步写入
		return w.writer.Write(p) // 同步写入
	}

	// 复制日志内容，避免被修改
	data := make([]byte, len(p)) // 创建数据副本
	copy(data, p)                // 复制数据

	select {
	case w.buffer <- data: // 写入缓冲区
		return len(p), nil // 返回写入长度
	default:
		if w.config.BlockOnFull {
			// 阻塞写入
			w.buffer <- data   // 写入缓冲区
			return len(p), nil // 返回写入长度
		} else if w.config.DropOnFull {
			// 丢弃日志
			atomic.AddInt64(&w.dropCount, 1) // 增加丢弃计数
			return len(p), nil               // 返回写入长度
		}
		// 同步写入
		return w.writer.Write(p) // 同步写入
	}
}

// start 启动异步写入
func (w *AsyncWriter) start() {
	w.wg.Add(1) // 增加等待组计数
	go w.run()  // 启动异步写入循环
}

// Stop 停止异步写入
func (w *AsyncWriter) Stop() error {
	close(w.done) // 关闭退出通道
	w.wg.Wait()   // 等待异步写入完成

	if w.config.FlushOnExit {
		// 先刷新数据
		if err := w.Flush(); err != nil { // 刷新数据
			return errors.NewOperationFailedError("failed to flush async writer", err)
		}
	}
	return nil
}

// Flush 刷新缓冲区
func (w *AsyncWriter) Flush() error {
	if !w.config.Enable { // 如果未启用异步写入
		return nil
	}

	// 直接处理缓冲区中的数据
	for {
		select {
		case data := <-w.buffer:
			if _, err := w.writer.Write(data); err != nil { // 写入数据
				return errors.NewOperationFailedError("failed to write data in flush", err)
			}
		default:
			// 缓冲区已空
			return nil
		}
	}
}

// run 运行异步写入循环
func (w *AsyncWriter) run() {
	defer w.wg.Done() // 减少等待组计数

	ticker := time.NewTicker(w.config.FlushInterval) // 创建定时器
	defer ticker.Stop()

	buffer := make([][]byte, 0, w.config.BufferSize) // 创建缓冲区

	flushBuffer := func() {
		if len(buffer) == 0 {
			return
		}

		// 合并多条日志一次写入
		totalSize := 0
		for _, b := range buffer {
			totalSize += len(b) // 计算总大小
		}

		data := make([]byte, 0, totalSize) // 创建数据
		for _, b := range buffer {
			data = append(data, b...) // 合并数据
		}

		if _, err := w.writer.Write(data); err != nil { // 写入数据
			// 使用错误处理中间件记录错误
			_ = errors.NewOperationFailedError("async write failed", err)
		}

		// 清空缓冲区
		buffer = buffer[:0] // 清空缓冲区
	}

	for {
		select {
		case <-w.done: // 退出通道触发
			flushBuffer() // 刷新缓冲区
			return
		case <-ticker.C: // 定时器触发
			flushBuffer()
		case data := <-w.buffer: // 缓冲区触发
			buffer = append(buffer, data)           // 添加数据
			if len(buffer) >= w.config.BufferSize { // 如果缓冲区满
				flushBuffer() // 刷新缓冲区
			}
		}
	}
}

// GetDropCount 获取丢弃的日志数量
func (w *AsyncWriter) GetDropCount() int64 {
	return atomic.LoadInt64(&w.dropCount) // 获取丢弃计数
}
