package elk

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

// 错误定义
var (
	ErrProcessorClosed   = errors.NewError(codes.ELKBulkError, "bulk processor is closed", nil)
	ErrDocumentTooLarge  = errors.NewError(codes.ELKBulkError, "document size exceeds flush bytes limit", nil)
	ErrRetryExhausted    = errors.NewError(codes.ELKBulkError, "retry attempts exhausted", nil)
	ErrCloseTimeout      = errors.NewError(codes.ELKTimeoutError, "close operation timed out", nil)
	ErrInvalidBatchSize  = errors.NewError(codes.ELKBulkError, "batch size must be greater than 0", nil)
	ErrInvalidFlushBytes = errors.NewError(codes.ELKBulkError, "flush bytes must be greater than 0", nil)
	ErrNilConfig         = errors.NewError(codes.ELKBulkError, "configuration cannot be nil", nil)
)

// BulkStats 统计信息
type BulkStats struct {
	TotalDocuments int64     // 处理的总文档数
	TotalBytes     int64     // 处理的总字节数
	FlushCount     int64     // 刷新次数
	ErrorCount     int64     // 错误次数
	LastError      error     // 最后一次错误
	LastFlushTime  time.Time // 最后一次刷新时间
}

// BulkProcessor 处理批量操作的接口
type BulkProcessor interface {
	Add(ctx context.Context, index string, doc interface{}) error
	Flush(ctx context.Context) error
	Close() error
	Stats() *BulkStats // 新增：获取统计信息
	MaxDocSize() int64 // 新增：获取单个文档最大大小
}

// BulkProcessorConfig 批量处理器的配置
type BulkProcessorConfig struct {
	BatchSize    int           // 触发刷新的文档数量
	FlushBytes   int64         // 触发刷新的字节数
	Interval     time.Duration // 自动刷新的时间间隔
	DefaultIndex string        // 默认索引名称
	RetryCount   int           // 重试次数
	RetryWait    time.Duration // 初始重试等待时间
	MaxWait      time.Duration // 最大重试等待时间
	CloseTimeout time.Duration // 关闭超时时间
}

// indexedDocument 包装文档和其目标索引
type indexedDocument struct {
	index string
	doc   interface{}
}

// bulkProcessor 实现 BulkProcessor 接口
type bulkProcessor struct {
	client    Client
	config    *BulkProcessorConfig
	documents []indexedDocument
	bytesSize int64
	mu        sync.Mutex
	closed    bool
	flushChan chan struct{}
	done      chan struct{}
	closeOnce sync.Once // 新增：确保只关闭一次

	// 统计信息，使用原子操作
	totalDocs  atomic.Int64
	totalBytes atomic.Int64
	flushCount atomic.Int64
	errorCount atomic.Int64
	lastError  error
	lastFlush  atomic.Value // 存储 time.Time
}

// calculateDocumentSize 计算文档的字节大小
func calculateDocumentSize(doc interface{}) (int64, error) {
	data, err := json.Marshal(doc)
	if err != nil {
		return 0, errors.Wrap(err, "failed to marshal document")
	}
	return int64(len(data)), nil
}

// NewBulkProcessor 创建新的批量处理器
func NewBulkProcessor(client Client, config *BulkProcessorConfig) BulkProcessor {
	if config == nil {
		return nil
	}

	// 验证必要的配置参数
	if config.BatchSize <= 0 {
		return nil
	}
	if config.FlushBytes <= 0 {
		return nil
	}

	// 设置默认值
	if config.DefaultIndex == "" {
		config.DefaultIndex = "default"
	}
	if config.RetryCount <= 0 {
		config.RetryCount = 3 // 默认重试3次
	}
	if config.RetryWait <= 0 {
		config.RetryWait = time.Second // 默认等待1秒
	}
	if config.CloseTimeout <= 0 {
		config.CloseTimeout = 30 * time.Second // 默认30秒超时
	}

	bp := &bulkProcessor{
		client:    client,
		config:    config,
		documents: make([]indexedDocument, 0, config.BatchSize),
		bytesSize: 0,
		flushChan: make(chan struct{}, 1),
		done:      make(chan struct{}),
	}

	// 初始化最后刷新时间
	bp.lastFlush.Store(time.Now())

	go bp.flushRoutine()
	return bp
}

// withRetry 包装需要重试的操作
func (bp *bulkProcessor) withRetry(ctx context.Context, operation func(context.Context) error) error {
	var lastErr error
	maxAttempts := bp.config.RetryCount + 1

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// 如果不是第一次尝试，则等待
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return errors.WrapWithCode(ctx.Err(), codes.ELKTimeoutError, "operation cancelled during retry wait")
			case <-time.After(bp.config.RetryWait):
			}
		}

		err := operation(ctx)
		if err == nil {
			return nil
		}

		lastErr = err
		bp.errorCount.Add(1)
		bp.lastError = err

		// 如果是最后一次尝试，返回错误
		if attempt == maxAttempts-1 {
			return errors.WrapWithCode(lastErr, codes.ELKBulkError, "max retry attempts reached")
		}
	}

	return lastErr
}

// Add 添加文档到批处理器
func (bp *bulkProcessor) Add(ctx context.Context, index string, doc interface{}) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if bp.closed {
		return errors.NewError(codes.ELKBulkError, "processor is closed", nil)
	}

	// 计算文档大小
	docSize, err := calculateDocumentSize(doc)
	if err != nil {
		return errors.WrapWithCode(err, codes.SerializationError, "failed to calculate document size")
	}

	// 检查单个文档大小是否超过限制
	if docSize > bp.config.FlushBytes {
		return errors.NewError(codes.ELKBulkError, "document size exceeds maximum allowed size", nil)
	}

	// 如果当前批次大小加上新文档会超过限制，先刷新
	if bp.bytesSize+docSize > bp.config.FlushBytes {
		if err := bp.flushLocked(ctx); err != nil {
			return errors.WrapWithCode(err, codes.ELKBulkError, "failed to flush before adding document")
		}
	}

	// 添加文档到缓冲区
	bp.documents = append(bp.documents, indexedDocument{
		index: index,
		doc:   doc,
	})
	bp.bytesSize += docSize
	bp.totalDocs.Add(1)
	bp.totalBytes.Add(docSize)

	// 如果达到批次大小，触发刷新
	if len(bp.documents) >= bp.config.BatchSize {
		return bp.flushLocked(ctx)
	}

	return nil
}

func (bp *bulkProcessor) Flush(ctx context.Context) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return bp.flushLocked(ctx)
}

func (bp *bulkProcessor) flushLocked(ctx context.Context) error {
	if len(bp.documents) == 0 {
		return nil
	}

	// 复制需要刷新的文档
	docsToFlush := make([]indexedDocument, len(bp.documents))
	copy(docsToFlush, bp.documents)

	// 清空缓冲区
	bp.documents = make([]indexedDocument, 0, bp.config.BatchSize)
	bp.bytesSize = 0

	// 在释放锁后执行刷新操作
	return bp.withRetry(ctx, func(ctx context.Context) error {
		// 按索引分组文档
		indexedDocs := make(map[string][]interface{})
		for _, idoc := range docsToFlush {
			if idoc.index == "" {
				idoc.index = bp.config.DefaultIndex
			}
			indexedDocs[idoc.index] = append(indexedDocs[idoc.index], idoc.doc)
		}

		// 只在成功时添加文档
		for index, docs := range indexedDocs {
			if err := bp.client.BulkIndexDocuments(ctx, index, docs); err != nil {
				return err
			}
		}
		return nil
	})
}

func (bp *bulkProcessor) Close() error {
	var closeErr error
	bp.closeOnce.Do(func() {
		bp.mu.Lock()
		if bp.closed {
			bp.mu.Unlock()
			return
		}
		bp.closed = true
		bp.mu.Unlock()

		// 执行优雅关闭
		closeErr = bp.gracefulClose(context.Background())
	})
	return closeErr
}

// gracefulClose 优雅关闭处理
func (bp *bulkProcessor) gracefulClose(ctx context.Context) error {
	closeCtx, cancel := context.WithTimeout(ctx, bp.config.CloseTimeout)
	defer cancel()

	close(bp.done)

	doneChan := make(chan error, 1)
	go func() {
		bp.mu.Lock()
		defer bp.mu.Unlock()
		doneChan <- bp.flushLocked(closeCtx)
	}()

	select {
	case err := <-doneChan:
		if err != nil {
			return errors.WrapWithCode(err, codes.ELKBulkError, "error during graceful shutdown")
		}
		return nil
	case <-closeCtx.Done():
		if closeCtx.Err() == context.DeadlineExceeded {
			return errors.NewError(codes.ELKTimeoutError, "shutdown timeout exceeded", nil)
		}
		return errors.WrapWithCode(closeCtx.Err(), codes.ELKTimeoutError, "context cancelled during shutdown")
	}
}

func (bp *bulkProcessor) flushRoutine() {
	ticker := time.NewTicker(bp.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), bp.config.CloseTimeout)
			if err := bp.Flush(ctx); err != nil {
				bp.errorCount.Add(1)
				bp.lastError = errors.Wrap(err, "failed to flush in background routine")
			}
			cancel()
		case <-bp.done:
			return
		}
	}
}

// Stats 返回当前统计信息的快照
func (bp *bulkProcessor) Stats() *BulkStats {
	return &BulkStats{
		TotalDocuments: bp.totalDocs.Load(),
		TotalBytes:     bp.totalBytes.Load(),
		FlushCount:     bp.flushCount.Load(),
		ErrorCount:     bp.errorCount.Load(),
		LastError:      bp.lastError,
		LastFlushTime:  bp.lastFlush.Load().(time.Time),
	}
}

// MaxDocSize 返回单个文档的最大大小限制
func (bp *bulkProcessor) MaxDocSize() int64 {
	// 使用配置中的 FlushBytes 作为单个文档的大小限制
	return bp.config.FlushBytes
}
