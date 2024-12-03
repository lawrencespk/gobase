package elk

import (
	"sync"
	"time"
)

// logBuffer 日志缓冲区
type logBuffer struct {
	buffer    []map[string]interface{} // 缓冲区
	mutex     sync.Mutex               // 互斥锁
	batchSize int                      // 批量大小
	flushChan chan struct{}            // 刷新通道
}

// newLogBuffer 创建日志缓冲区
func newLogBuffer(batchSize int, flushInterval time.Duration) *logBuffer {
	b := &logBuffer{
		buffer:    make([]map[string]interface{}, 0, batchSize), // 创建缓冲区
		batchSize: batchSize,                                    // 设置批量大小
		flushChan: make(chan struct{}),                          // 创建刷新通道
	}

	go b.flushRoutine(flushInterval) // 启动刷新协程
	return b
}

// Add 添加日志
func (b *logBuffer) Add(entry map[string]interface{}) {
	b.mutex.Lock()                     // 锁定
	b.buffer = append(b.buffer, entry) // 添加日志
	if len(b.buffer) >= b.batchSize {  // 如果缓冲区满
		b.flushChan <- struct{}{} // 发送刷新信号
	}
	b.mutex.Unlock() // 解锁
}

// flushRoutine 刷新日志
func (b *logBuffer) flushRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval) // 创建定时器
	defer ticker.Stop()                // 停止定时器

	for {
		select {
		case <-b.flushChan: // 收到刷新信号
			b.flush() // 刷新日志
		case <-ticker.C: // 定时器触发
			b.flush() // 刷新日志
		}
	}
}

// flush 刷新日志
func (b *logBuffer) flush() {
	b.mutex.Lock()         // 锁定
	defer b.mutex.Unlock() // 解锁

	if len(b.buffer) == 0 {
		return
	}

	// TODO: 实现批量写入到 Elasticsearch
	b.buffer = b.buffer[:0] // 清空缓冲区
}
