package logrus

import (
	"bytes"
	"sync"
)

// BufferPool 是一个内存缓冲池
type BufferPool struct {
	pool sync.Pool
}

// NewBufferPool 创建一个新的缓冲池
func NewBufferPool() *BufferPool {
	// 创建一个新的缓冲池
	return &BufferPool{
		pool: sync.Pool{ // 同步池
			New: func() interface{} {
				return new(bytes.Buffer) // 创建一个新的缓冲区
			},
		},
	}
}

// Get 从池中获取一个缓冲区
func (p *BufferPool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer) // 从池中获取一个缓冲区
}

// Put 将缓冲区放回池中
func (p *BufferPool) Put(buf *bytes.Buffer) {
	buf.Reset()     // 重置缓冲区
	p.pool.Put(buf) // 将缓冲区放回池中
}

// WritePool 是一个写入器缓冲池
type WritePool struct {
	pool sync.Pool // 同步池
}

// NewWritePool 创建一个新的写入器池
func NewWritePool(size int) *WritePool {
	return &WritePool{
		pool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, 0, size) // 创建一个新的缓冲区
				return &buf
			},
		},
	}
}

// Get 从池中获取一个写入缓冲区
func (p *WritePool) Get() []byte {
	return *(p.pool.Get().(*[]byte)) // 从池中获取一个写入缓冲区
}

// Put 将写入缓冲区放回池中
func (p *WritePool) Put(buf []byte) {
	buf = buf[:0]    // 重置缓冲区
	p.pool.Put(&buf) // 将缓冲区放回池中
}
