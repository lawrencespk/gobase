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
	if err := validateConfig(config); err != nil {
		return nil, err
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
		q.wg.Add(1)
		go q.worker()
	}

	return q, nil
}

// Write 写入数据
func (q *WriteQueue) Write(p []byte) (n int, err error) {
	// 检查队列是否正在运行
	if atomic.LoadInt32(&q.running) == 0 {
		return 0, errors.New("queue not running")
	}

	// 复制数据
	data := make([]byte, len(p))
	copy(data, p)

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

	buffer := make([]byte, 0, q.config.BatchSize*2)
	ticker := time.NewTicker(q.config.FlushInterval)
	defer ticker.Stop()

	flush := func() bool {
		if len(buffer) > 0 {
			// 创建数据副本进行写入
			data := make([]byte, len(buffer))
			copy(data, buffer)

			if !q.writeWithRetry(data) {
				return false
			}
			buffer = buffer[:0]
		}
		return true
	}

	for {
		select {
		case <-q.done:
			flush()
			return
		case data, ok := <-q.queue:
			if !ok {
				flush()
				return
			}
			buffer = append(buffer, data...)
			if len(buffer) >= q.config.BatchSize {
				if !flush() {
					return
				}
			}
		case <-ticker.C:
			if !flush() {
				return
			}
		}
	}
}

// writeWithRetry 带重试的写入操作
func (q *WriteQueue) writeWithRetry(data []byte) bool {
	retries := q.config.RetryCount
	if retries <= 0 {
		retries = 3 // 默认重试次数
	}

	interval := q.config.RetryInterval
	if interval <= 0 {
		interval = time.Millisecond * 100 // 默认重试间隔
	}

	for i := 0; i < retries; i++ {
		if _, err := q.writer.Write(data); err == nil {
			return true
		} else if q.errorHandler != nil {
			q.errorHandler(err)
		}

		if atomic.LoadInt32(&q.running) == 0 {
			return false
		}

		time.Sleep(interval * time.Duration(i+1))
	}
	return false
}

// Close 关闭队列
func (q *WriteQueue) Close(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&q.running, 1, 0) {
		return nil
	}

	close(q.queue)

	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		close(q.done)
		return ctx.Err()
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
