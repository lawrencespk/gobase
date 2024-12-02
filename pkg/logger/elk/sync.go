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

	go b.flushRoutine(flushInterval)
	return b
}

// Add 添加日志
func (b *logBuffer) Add(entry map[string]interface{}) {
	b.mutex.Lock()
	b.buffer = append(b.buffer, entry)
	if len(b.buffer) >= b.batchSize {
		b.flushChan <- struct{}{}
	}
	b.mutex.Unlock()
}

// flushRoutine 刷新日志
func (b *logBuffer) flushRoutine(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-b.flushChan:
			b.flush()
		case <-ticker.C:
			b.flush()
		}
	}
}

// flush 刷新日志
func (b *logBuffer) flush() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if len(b.buffer) == 0 {
		return
	}

	// TODO: 实现批量写入到 Elasticsearch
	b.buffer = b.buffer[:0]
}
