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
	ErrProcessorClosed  = errors.NewError(codes.ELKBulkError, "bulk processor is closed", nil)
	ErrDocumentTooLarge = errors.NewError(codes.ELKBulkError, "document size exceeds flush bytes limit", nil)
	ErrRetryExhausted   = errors.NewError(codes.ELKBulkError, "retry attempts exhausted", nil)
	ErrCloseTimeout     = errors.NewError(codes.ELKTimeoutError, "close operation timed out", nil)
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
}

// BulkProcessorConfig 批量处理器的配置
type BulkProcessorConfig struct {
	BatchSize    int           // 触发刷新的文档数量
	FlushBytes   int64         // 触发刷新的字节数
	Interval     time.Duration // 自动刷新的时间间隔
	DefaultIndex string        // 默认索引名称
	RetryCount   int           // 新增：重试次数
	RetryWait    time.Duration // 新增：重试等待时间
	CloseTimeout time.Duration // 新增：关闭超时时间
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
	for attempt := 0; attempt <= bp.config.RetryCount; attempt++ {
		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "context cancelled during retry")
		default:
			if err := operation(ctx); err == nil {
				return nil
			} else {
				lastErr = errors.Wrap(err, "operation failed during retry")
				if attempt < bp.config.RetryCount {
					waitTime := bp.config.RetryWait * time.Duration(attempt+1)
					select {
					case <-ctx.Done():
						return errors.Wrap(ctx.Err(), "context cancelled during retry wait")
					case <-time.After(waitTime):
						continue
					}
				}
			}
		}
	}
	return errors.Wrap(lastErr, ErrRetryExhausted.Error())
}

func (bp *bulkProcessor) Add(ctx context.Context, index string, doc interface{}) error {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if bp.closed {
		return ErrProcessorClosed
	}

	docSize, err := calculateDocumentSize(doc)
	if err != nil {
		return err
	}

	if docSize > bp.config.FlushBytes {
		if err := bp.flushLocked(ctx); err != nil {
			return err
		}
	}

	bp.documents = append(bp.documents, indexedDocument{index: index, doc: doc})
	bp.bytesSize += docSize
	bp.totalDocs.Add(1)
	bp.totalBytes.Add(docSize)

	if len(bp.documents) >= bp.config.BatchSize || bp.bytesSize >= bp.config.FlushBytes {
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

	indexedDocs := make(map[string][]interface{})
	for _, idoc := range bp.documents {
		indexedDocs[idoc.index] = append(indexedDocs[idoc.index], idoc.doc)
	}

	err := bp.withRetry(ctx, func(ctx context.Context) error {
		for index, docs := range indexedDocs {
			if err := bp.client.BulkIndexDocuments(ctx, index, docs); err != nil {
				bp.errorCount.Add(1)
				bp.lastError = err
				return err
			}
		}
		return nil
	})

	if err != nil {
		// 记录错误但不返回，让处理器继续工作
		bp.errorCount.Add(1)
		bp.lastError = err
		// 清空文档列表，防止重复处理失败的文档
		bp.documents = bp.documents[:0]
		bp.bytesSize = 0
		return nil
	}

	bp.documents = bp.documents[:0]
	bp.bytesSize = 0
	bp.flushCount.Add(1)
	bp.lastFlush.Store(time.Now())
	return nil
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
		return errors.Wrap(err, "error during graceful shutdown")
	case <-closeCtx.Done():
		if closeCtx.Err() == context.DeadlineExceeded {
			return ErrCloseTimeout
		}
		return errors.Wrap(closeCtx.Err(), "context cancelled during graceful shutdown")
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
