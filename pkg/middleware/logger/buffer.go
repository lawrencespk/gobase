package logger

import (
	"bytes"
	"sync"
	"time"
)

// Buffer 日志缓冲器
type Buffer struct {
	buf           *bytes.Buffer
	mutex         sync.Mutex
	flushInterval time.Duration
	maxSize       int
	flushOnError  bool
	writer        BodyWriter
	flushChan     chan struct{}
	closeChan     chan struct{}
}

// NewBuffer 创建缓冲器
func NewBuffer(config BufferConfig, writer BodyWriter) *Buffer {
	if !config.Enable {
		return nil
	}

	b := &Buffer{
		buf:           bytes.NewBuffer(make([]byte, 0, config.Size)),
		flushInterval: time.Duration(config.FlushInterval) * time.Millisecond,
		maxSize:       config.Size,
		flushOnError:  config.FlushOnError,
		writer:        writer,
		flushChan:     make(chan struct{}, 1),
		closeChan:     make(chan struct{}),
	}

	// 启动定时刷新
	go b.flushLoop()

	return b
}

// Write 实现io.Writer接口
func (b *Buffer) Write(p []byte) (n int, err error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// 如果缓冲区已满，先刷新
	if b.buf.Len()+len(p) > b.maxSize {
		if err := b.flush(); err != nil {
			return 0, newLogBufferFullError("buffer is full", err)
		}
	}

	n, err = b.buf.Write(p)
	if err != nil {
		return n, newLogWriteError("failed to write to buffer", err)
	}

	return n, nil
}

// flush 刷新缓冲区
func (b *Buffer) flush() error {
	if b.buf.Len() == 0 {
		return nil
	}

	// 写入底层writer
	_, err := b.writer.Write(b.buf.Bytes())
	if err != nil {
		return newLogFlushError("failed to flush buffer", err)
	}

	// 清空缓冲区
	b.buf.Reset()
	return nil
}

// FlushOnError 错误时刷新
func (b *Buffer) FlushOnError() error {
	if !b.flushOnError {
		return nil
	}
	return b.Flush()
}

// Flush 手动刷新
func (b *Buffer) Flush() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.flush()
}

// flushLoop 定时刷新循环
func (b *Buffer) flushLoop() {
	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = b.Flush()
		case <-b.flushChan:
			_ = b.Flush()
		case <-b.closeChan:
			_ = b.Flush()
			return
		}
	}
}

// Close 关闭缓冲器
func (b *Buffer) Close() error {
	close(b.closeChan)
	return b.Flush()
}
