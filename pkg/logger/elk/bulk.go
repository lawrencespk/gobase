package elk

import (
	"bytes"
	"context"
	"encoding/json"
	"sync"
	"time"

	"gobase/pkg/config"
	"gobase/pkg/errors"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// BulkProcessor 处理批量操作
type BulkProcessor struct {
	client      *ElkClient
	buffer      *bytes.Buffer
	bufferMutex sync.Mutex
	batchSize   int           // 批量大小
	flushBytes  int           // 刷新字节数阈值
	flushTimer  *time.Timer   // 定时刷新计时器
	interval    time.Duration // 刷新间隔
}

// BulkProcessorConfig 批量处理器配置
type BulkProcessorConfig struct {
	BatchSize  int           // 批量大小
	FlushBytes int           // 刷新字节数阈值
	Interval   time.Duration // 刷新间隔
}

// DefaultBulkProcessorConfig 返回默认配置
func DefaultBulkProcessorConfig() *BulkProcessorConfig {
	conf := config.GetConfig()
	if conf == nil {
		return &BulkProcessorConfig{
			BatchSize:  1000,
			FlushBytes: 5 * 1024 * 1024,
			Interval:   30 * time.Second,
		}
	}

	interval, err := time.ParseDuration(conf.ELK.Bulk.Interval)
	if err != nil {
		interval = 30 * time.Second
	}

	return &BulkProcessorConfig{
		BatchSize:  conf.ELK.Bulk.BatchSize,
		FlushBytes: conf.ELK.Bulk.FlushBytes,
		Interval:   interval,
	}
}

// NewBulkProcessor 创建新的批量处理器
func NewBulkProcessor(client *ElkClient, config *BulkProcessorConfig) *BulkProcessor {
	if config == nil {
		config = DefaultBulkProcessorConfig()
	}

	b := &BulkProcessor{
		client:     client,
		buffer:     &bytes.Buffer{},
		batchSize:  config.BatchSize,
		flushBytes: config.FlushBytes,
		interval:   config.Interval,
	}

	// 启动定时刷新
	b.startTimer()

	return b
}

// Add 添加文档到批量操作
func (b *BulkProcessor) Add(ctx context.Context, index string, document interface{}) error {
	if !b.client.isConnected {
		return errors.NewELKConnectionError("client is not connected", nil)
	}

	// 构建批量操作元数据
	meta := map[string]interface{}{
		"index": map[string]interface{}{
			"_index": index,
		},
	}

	// 序列化元数据和文档
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return errors.NewELKBulkError("failed to marshal metadata", err)
	}

	docBytes, err := json.Marshal(document)
	if err != nil {
		return errors.NewELKBulkError("failed to marshal document", err)
	}

	b.bufferMutex.Lock()
	defer b.bufferMutex.Unlock()

	// 写入批量操作格式
	b.buffer.Write(metaBytes)
	b.buffer.WriteString("\n")
	b.buffer.Write(docBytes)
	b.buffer.WriteString("\n")

	// 检查是否需要刷新
	if b.buffer.Len() >= b.flushBytes {
		return b.Flush(ctx)
	}

	return nil
}

// Flush 刷新批量操作
func (b *BulkProcessor) Flush(ctx context.Context) error {
	b.bufferMutex.Lock()
	defer b.bufferMutex.Unlock()

	if b.buffer.Len() == 0 {
		return nil
	}

	req := esapi.BulkRequest{
		Body: bytes.NewReader(b.buffer.Bytes()),
	}

	res, err := req.Do(ctx, b.client.client)
	if err != nil {
		return errors.NewELKBulkError("failed to execute bulk operation", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.NewELKBulkError(
			"bulk operation failed: "+res.String(),
			nil,
		)
	}

	// 解析响应
	var bulkResponse struct {
		Errors bool `json:"errors"`
		Items  []struct {
			Index struct {
				Status int    `json:"status"`
				Error  string `json:"error,omitempty"`
			} `json:"index"`
		} `json:"items"`
	}

	if err := json.NewDecoder(res.Body).Decode(&bulkResponse); err != nil {
		return errors.NewELKBulkError("failed to decode bulk response", err)
	}

	// 检查是否有错误
	if bulkResponse.Errors {
		// 这里可以添加更详细的错误处理逻辑
		return errors.NewELKBulkError("some documents failed to index", nil)
	}

	// 清空缓冲区
	b.buffer.Reset()

	return nil
}

// Close 关闭批量处理器
func (b *BulkProcessor) Close() error {
	if b.flushTimer != nil {
		b.flushTimer.Stop()
	}

	// 尝试最后一次刷新
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return b.Flush(ctx)
}

// startTimer 启动定时刷新
func (b *BulkProcessor) startTimer() {
	b.flushTimer = time.NewTimer(b.interval)
	go func() {
		for range b.flushTimer.C {
			ctx := context.Background()
			if err := b.Flush(ctx); err != nil {
				// 处理错误
				switch {
				case errors.Is(err, errors.NewELKConnectionError("", nil)):
					// 连接错误，可能需要重试
					if log != nil {
						log.Error(errors.Wrap(err, "ELK bulk processor connection error, retrying"))
					}
					time.Sleep(time.Second)

				case errors.Is(err, errors.NewELKBulkError("", nil)):
					// 批量操作错误
					if log != nil {
						log.Error(errors.Wrap(err, "ELK bulk operation failed"))
					}

				default:
					// 其他错误
					if log != nil {
						log.Error(errors.Wrap(err, "Unknown error during ELK bulk flush"))
					}
				}
			}
			b.flushTimer.Reset(b.interval)
		}
	}()
}

// SetLogger 设置日志实例
func SetLogger(logger Logger) {
	log = logger
}
