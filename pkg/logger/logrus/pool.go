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
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

// Get 从池中获取一个缓冲区
func (p *BufferPool) Get() *bytes.Buffer {
	return p.pool.Get().(*bytes.Buffer)
}

// Put 将缓冲区放回池中
func (p *BufferPool) Put(buf *bytes.Buffer) {
	buf.Reset()
	p.pool.Put(buf)
}

// WritePool 是一个写入器缓冲池
type WritePool struct {
	pool sync.Pool
}

// NewWritePool 创建一个新的写入器池
func NewWritePool(size int) *WritePool {
	return &WritePool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 0, size)
			},
		},
	}
}

// Get 从池中获取一个写入缓冲区
func (p *WritePool) Get() []byte {
	return p.pool.Get().([]byte)
}

// Put 将写入缓冲区放回池中
func (p *WritePool) Put(buf []byte) {
	p.pool.Put(&buf)
}
