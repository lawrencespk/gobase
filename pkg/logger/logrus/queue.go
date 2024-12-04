package logrus

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// 定义错误变量
var (
	errQueueFull = errors.New("write queue is full")
)

// QueueConfig 队列配置
type QueueConfig struct {
	MaxSize         int           // 最大队列大小
	BatchSize       int           // 批处理大小
	FlushInterval   time.Duration // 刷新间隔
	Workers         int           // 工作协程数量
	RetryCount      int           // 重试次数
	RetryInterval   time.Duration // 重试间隔
	MaxBatchWait    time.Duration // 最大批处理等待时间
	ShutdownTimeout time.Duration // 关闭超时时间
}

// WriteQueue 写入队列
type WriteQueue struct {
	writer       Writer         // 写入器
	config       QueueConfig    // 配置
	queue        chan []byte    // 队列
	done         chan struct{}  // 结束信号
	wg           sync.WaitGroup // 等待组
	running      int32          // 使用原子操作
	errorHandler ErrorHandler   // 错误处理
}

// validateConfig 验证配置
func validateConfig(config QueueConfig) error {
	if config.MaxSize <= 0 {
		return errors.New("maxSize must be positive") // 最大队列大小必须为正数
	}
	if config.BatchSize <= 0 {
		return errors.New("batchSize must be positive") // 批处理大小必须为正数
	}
	if config.Workers <= 0 {
		return errors.New("workers must be positive") // 工作协程数量必须为正数
	}
	if config.FlushInterval <= 0 {
		return errors.New("flushInterval must be positive") // 刷新间隔必须为正数
	}
	return nil
}

// NewWriteQueue 创建新的写入队列
func NewWriteQueue(writer Writer, config QueueConfig) (*WriteQueue, error) {
	// 验证配置
	if err := validateConfig(config); err != nil { // 验证配置
		return nil, err // 返回错误
	}

	q := &WriteQueue{
		writer: writer,                            // 写入器
		config: config,                            // 配置
		queue:  make(chan []byte, config.MaxSize), // 队列
		done:   make(chan struct{}),               // 结束信号
	}

	atomic.StoreInt32(&q.running, 1)

	// 启动工作协程
	for i := 0; i < config.Workers; i++ {
		q.wg.Add(1)   // 增加等待组
		go q.worker() // 启动工作协程
	}

	return q, nil
}

// Write 写入数据
func (q *WriteQueue) Write(p []byte) (n int, err error) {
	// 检查队列是否正在运行
	if atomic.LoadInt32(&q.running) == 0 { // 检查队列是否正在运行
		return 0, errors.New("queue not running") // 返回错误
	}

	// 复制数据
	data := make([]byte, len(p)) // 创建数据副本
	copy(data, p)                // 复制数据

	// 使用带超时的 select 来平衡响应速度和队列满的检测
	select {
	// 数据写入
	case q.queue <- data:
		return len(p), nil
	// 极短的超时，用于快速检测队列是否已满
	case <-time.After(time.Microsecond):
		// 极短的超时，用于快速检测队列是否已满
		return 0, errQueueFull
	}
}

// worker 工作协程
func (q *WriteQueue) worker() {
	defer q.wg.Done()

	buffer := make([]byte, 0, q.config.BatchSize*2)  // 创建缓冲区
	ticker := time.NewTicker(q.config.FlushInterval) // 创建定时器
	defer ticker.Stop()                              // 停止定时器

	flush := func() bool {
		if len(buffer) > 0 {
			// 创建数据副本进行写入
			data := make([]byte, len(buffer)) // 创建数据副本
			copy(data, buffer)                // 复制数据

			if !q.writeWithRetry(data) { // 写入数据
				return false
			}
			buffer = buffer[:0] // 清空缓冲区
		}
		return true
	}

	for {
		select {
		case <-q.done: // 结束信号
			flush()
			return
		case data, ok := <-q.queue: // 从队列中读取数据
			if !ok {
				flush()
				return
			}
			buffer = append(buffer, data...)       // 添加数据到缓冲区
			if len(buffer) >= q.config.BatchSize { // 如果缓冲区大小达到批处理大小
				if !flush() { // 写入数据
					return
				}
			}
		case <-ticker.C: // 定时器
			if !flush() { // 写入数据
				return
			}
		}
	}
}

// writeWithRetry 带重试的写入操作
func (q *WriteQueue) writeWithRetry(data []byte) bool {
	retries := q.config.RetryCount // 重试次数
	if retries <= 0 {
		retries = 3 // 默认重试次数
	}

	interval := q.config.RetryInterval // 重试间隔
	if interval <= 0 {
		interval = time.Millisecond * 100 // 默认重试间隔
	}

	for i := 0; i < retries; i++ {
		if _, err := q.writer.Write(data); err == nil { // 写入数据
			return true // 写入成功
		} else if q.errorHandler != nil { // 如果错误处理器不为空
			q.errorHandler(err)
		}

		if atomic.LoadInt32(&q.running) == 0 { // 检查队列是否正在运行
			return false // 返回失败
		}

		time.Sleep(interval * time.Duration(i+1)) // 等待重试
	}
	return false // 返回失败
}

// Close 关闭队列
func (q *WriteQueue) Close(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&q.running, 1, 0) { // 检查队列是否正在运行
		return nil // 返回成功
	}

	close(q.queue)

	done := make(chan struct{})
	go func() {
		q.wg.Wait() // 等待工作协程完成
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done(): // 上下文结束
		close(q.done)    // 关闭结束信号
		return ctx.Err() // 返回错误
	}
}

// SetErrorHandler 设置错误处理器
func (q *WriteQueue) SetErrorHandler(handler ErrorHandler) {
	q.errorHandler = handler
}

// ErrorHandler 错误处理函数类型
type ErrorHandler func(error)

// IsQueueFullError 检查错误是否为队列已满错误
func IsQueueFullError(err error) bool {
	return err == errQueueFull
}
