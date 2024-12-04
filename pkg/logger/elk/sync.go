package elk

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/olivere/elastic/v7"
)

// logBuffer 日志缓冲区
type logBuffer struct {
	buffer       []map[string]interface{} // 缓冲区
	client       *elastic.Client          // ES客户端
	indexPrefix  string                   // 索引前缀
	mutex        sync.Mutex               // 互斥锁
	batchSize    int                      // 批量大小
	flushChan    chan struct{}            // 刷新通道
	done         chan struct{}            // 完成通道
	errorHandler func(error)              // 错误处理函数
}

// newLogBuffer 创建日志缓冲区
func newLogBuffer(client *elastic.Client, config *ElasticConfig) *logBuffer {
	b := &logBuffer{
		buffer:      make([]map[string]interface{}, 0, config.BatchSize), // 缓冲区
		client:      client,                                              // ES客户端
		indexPrefix: config.IndexPrefix,                                  // 索引前缀
		batchSize:   config.BatchSize,                                    // 批量大小
		flushChan:   make(chan struct{}),                                 // 刷新通道
		done:        make(chan struct{}),                                 // 完成通道
	}

	go b.flushRoutine(config.FlushInterval) // 启动刷新协程
	return b
}

// Add 添加日志
func (b *logBuffer) Add(entry map[string]interface{}) {
	b.mutex.Lock()                     // 加锁
	b.buffer = append(b.buffer, entry) // 添加日志
	if len(b.buffer) >= b.batchSize {  // 如果缓冲区满了
		b.flushChan <- struct{}{} // 发送刷新信号
	}
	b.mutex.Unlock() // 解锁
}

// flushRoutine 刷新日志
func (b *logBuffer) flushRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-b.done: // 完成通道
			b.flush()
			return
		case <-b.flushChan: // 刷新通道
			b.flush()
		case <-ticker.C: // 定时器
			b.flush()
		}
	}
}

// flush 刷新日志
func (b *logBuffer) flush() {
	b.mutex.Lock()          // 加锁
	if len(b.buffer) == 0 { // 如果缓冲区为空
		b.mutex.Unlock() // 解锁
		return
	}

	// 创建批量请求
	bulk := b.client.Bulk()                                                           // 创建批量请求
	indexName := fmt.Sprintf("%s-%s", b.indexPrefix, time.Now().Format("2006.01.02")) // 获取当前索引名称

	// 添加所有文档到批量请求
	for _, doc := range b.buffer {
		req := elastic.NewBulkIndexRequest(). // 创建索引请求
							Index(indexName). // 设置索引
							Doc(doc)          // 设置文档
		bulk.Add(req) // 添加请求
	}

	// 清空缓冲区
	b.buffer = b.buffer[:0] // 清空缓冲区
	b.mutex.Unlock()        // 解锁

	// 执行批量请求
	ctx := context.Background() // 创建上下文
	resp, err := bulk.Do(ctx)   // 执行批量请求
	if err != nil {             // 如果执行失败
		if b.errorHandler != nil { // 如果错误处理函数不为空
			b.errorHandler(fmt.Errorf("bulk index failed: %w", err)) // 调用错误处理函数
		}
		return
	}

	// 处理失败的项目
	if resp.Errors { // 如果批量请求失败
		var failedItems []string             // 失败的项目
		for _, item := range resp.Failed() { // 遍历失败的项目
			failedItems = append(failedItems, item.Error.Reason) // 添加失败的项目
		}
		if b.errorHandler != nil { // 如果错误处理函数不为空
			b.errorHandler(fmt.Errorf("bulk index partial failure: %v", failedItems)) // 调用错误处理函数
		}
	}
}

// Close 关闭缓冲区
func (b *logBuffer) Close() error {
	close(b.done)
	return nil
}

// SetErrorHandler 设置错误处理函数
func (b *logBuffer) SetErrorHandler(handler func(error)) {
	b.mutex.Lock()           // 加锁
	b.errorHandler = handler // 设置错误处理函数
	b.mutex.Unlock()         // 解锁
}
