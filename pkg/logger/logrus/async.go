package logrus

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
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
		aw.start()
	}

	return aw
}

// Write 实现 io.Writer 接口
func (w *AsyncWriter) Write(p []byte) (n int, err error) {
	if !w.config.Enable {
		return w.writer.Write(p)
	}

	// 复制日志内容，避免被修改
	data := make([]byte, len(p))
	copy(data, p)

	select {
	case w.buffer <- data:
		return len(p), nil
	default:
		if w.config.BlockOnFull {
			// 阻塞写入
			w.buffer <- data
			return len(p), nil
		} else if w.config.DropOnFull {
			// 丢弃日志
			atomic.AddInt64(&w.dropCount, 1)
			return len(p), nil
		}
		// 同步写入
		return w.writer.Write(p)
	}
}

// start 启动异步写入
func (w *AsyncWriter) start() {
	w.wg.Add(1)
	go w.run()
}

// Stop 停止异步写入
func (w *AsyncWriter) Stop() error {
	close(w.done)
	if w.config.FlushOnExit {
		return w.Flush()
	}
	w.wg.Wait()
	return nil
}

// Flush 刷新缓冲区
func (w *AsyncWriter) Flush() error {
	if !w.config.Enable {
		return nil
	}

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("flush timeout")
	}
}

// run 运行异步写入循环
func (w *AsyncWriter) run() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.FlushInterval)
	defer ticker.Stop()

	buffer := make([][]byte, 0, w.config.BufferSize)

	flushBuffer := func() {
		if len(buffer) == 0 {
			return
		}

		// 合并多条日志一次写入
		totalSize := 0
		for _, b := range buffer {
			totalSize += len(b)
		}

		data := make([]byte, 0, totalSize)
		for _, b := range buffer {
			data = append(data, b...)
		}

		if _, err := w.writer.Write(data); err != nil {
			fmt.Printf("async write error: %v\n", err)
		}

		// 清空缓冲区
		buffer = buffer[:0]
	}

	for {
		select {
		case <-w.done:
			flushBuffer()
			return
		case <-ticker.C:
			flushBuffer()
		case data := <-w.buffer:
			buffer = append(buffer, data)
			if len(buffer) >= w.config.BufferSize {
				flushBuffer()
			}
		}
	}
}

// GetDropCount 获取丢弃的日志数量
func (w *AsyncWriter) GetDropCount() int64 {
	return atomic.LoadInt64(&w.dropCount)
}
