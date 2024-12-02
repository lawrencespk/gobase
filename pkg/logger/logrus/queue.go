package logrus

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// 定义错误变量 (修改错误信息格式)
var (
	errQueueFull           = errors.New("write queue is full")
	errQueueAlreadyRunning = errors.New("queue already running")
	errQueueNotRunning     = errors.New("queue not running")
	errShutdownTimeout     = errors.New("shutdown timeout")
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
	queue        chan []byte    // 队列
	done         chan struct{}  // 结束信号
	config       QueueConfig    // 配置
	isRunning    atomic.Bool    // 运行状态
	metrics      *QueueMetrics  // 指标
	wg           sync.WaitGroup // 等待组
	errorHandler func(error)    // 错误处理
	workers      []*QueueWorker // 工作协程池
}

// QueueMetrics 队列指标
type QueueMetrics struct {
	enqueuedCount  atomic.Int64 // 入队消息数
	dequeuedCount  atomic.Int64 // 出队消息数
	droppedCount   atomic.Int64 // 丢弃消息数
	errorCount     atomic.Int64 // 错误次数
	queueLength    atomic.Int64 // 当前队列长度
	processingTime atomic.Int64 // 处理时间
}

// NewWriteQueue 创建写入队列
func NewWriteQueue(writer Writer, config QueueConfig) (*WriteQueue, error) {
	if writer == nil {
		return nil, fmt.Errorf("writer cannot be nil")
	}
	if err := validateQueueConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid queue config: %w", err)
	}

	q := &WriteQueue{
		queue:   make(chan []byte, config.MaxSize),
		config:  config,
		metrics: &QueueMetrics{},
		done:    make(chan struct{}),
		writer:  writer,
	}

	// 初始化工作协程池
	q.workers = make([]*QueueWorker, config.Workers)
	for i := 0; i < config.Workers; i++ {
		q.workers[i] = &QueueWorker{
			id:    i,
			queue: q,
		}
	}

	return q, nil
}

// Write 实现 io.Writer 接口
func (q *WriteQueue) Write(p []byte) (n int, err error) {
	if !q.isRunning.Load() {
		return 0, errQueueNotRunning
	}

	data := make([]byte, len(p))
	copy(data, p)

	select {
	case q.queue <- data:
		q.metrics.enqueuedCount.Add(1)
		q.metrics.queueLength.Add(1)
		return len(p), nil
	default:
		q.metrics.droppedCount.Add(1)
		return 0, errQueueFull
	}
}

// Close 关闭队列
func (q *WriteQueue) Close(ctx context.Context) error {
	close(q.done)

	// 等待所有工作协程退出
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Len 获取队列长度
func (q *WriteQueue) Len() int {
	return len(q.queue)
}

// IsEmpty 检查队列是否为空
func (q *WriteQueue) IsEmpty() bool {
	return q.Len() == 0
}

// 错误处理器
type ErrorHandler func(error)

// 启动队列
func (q *WriteQueue) Start() error {
	if !q.isRunning.CompareAndSwap(false, true) {
		return errQueueAlreadyRunning
	}

	// 启动所有工作协程
	for _, worker := range q.workers {
		q.wg.Add(1)
		go func(w *QueueWorker) {
			defer q.wg.Done()
			w.run()
		}(worker)
	}

	return nil
}

// 优雅关闭
func (q *WriteQueue) Stop() error {
	if !q.isRunning.CompareAndSwap(true, false) {
		return errQueueNotRunning
	}

	// 关闭队列通道
	close(q.queue)

	// 等待所有工作协程完成
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	// 等待超时或完成
	select {
	case <-done:
		return nil
	case <-time.After(q.config.ShutdownTimeout):
		return errShutdownTimeout
	}
}

// 定义 QueueWorker
type QueueWorker struct {
	id    int         // 工作协程ID
	queue *WriteQueue // 所属队列
}

// 添加配置验证函数
func validateQueueConfig(config *QueueConfig) error {
	if config.MaxSize <= 0 {
		return errors.New("maxSize must be positive")
	}
	if config.BatchSize <= 0 {
		return errors.New("batchSize must be positive")
	}
	if config.Workers <= 0 {
		return errors.New("workers must be positive")
	}
	if config.FlushInterval <= 0 {
		return errors.New("flushInterval must be positive")
	}
	return nil
}

func (w *QueueWorker) run() {
	for {
		select {
		case data, ok := <-w.queue.queue:
			if !ok {
				return
			}
			if err := w.processData(data); err != nil {
				w.queue.metrics.errorCount.Add(1)
				if w.queue.errorHandler != nil {
					w.queue.errorHandler(err)
				}
			}
		case <-w.queue.done:
			return
		}
	}
}

func (w *QueueWorker) processData(data []byte) error {
	// 处理单条数据
	start := time.Now()

	if _, err := w.queue.writer.Write(data); err != nil {
		return err
	}

	w.queue.metrics.processingTime.Add(time.Since(start).Microseconds())
	w.queue.metrics.dequeuedCount.Add(1)
	return nil
}

// IsQueueFullError 检查是否为队列已满错误
func IsQueueFullError(err error) bool {
	return err == errQueueFull
}
