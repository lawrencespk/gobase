package logger

import (
	"sync"

	"github.com/gin-gonic/gin"
)

var bodyWriterPool = sync.Pool{
	New: func() interface{} {
		return &bodyWriter{
			body: make([]byte, 0, 1024), // 预分配1KB空间
		}
	},
}

// bodyWriter 响应体Writer实现
type bodyWriter struct {
	gin.ResponseWriter
	body      []byte  // 响应体
	buffer    *Buffer // 缓冲区
	size      int     // 响应大小
	status    int     // 状态码
	committed bool    // 是否已提交
}

// 优化newBodyWriter创建
func newBodyWriter(c *gin.Context, config *Config) *bodyWriter {
	w := bodyWriterPool.Get().(*bodyWriter)
	w.ResponseWriter = c.Writer
	w.body = w.body[:0] // 重置
	w.size = 0
	w.status = 0
	w.committed = false

	// 如果启用缓冲，创建buffer
	if config.Buffer.Enable {
		w.buffer = NewBuffer(config.Buffer, w)
	}

	return w
}

// Write 实现Write接口
func (w *bodyWriter) Write(b []byte) (int, error) {
	if w.buffer != nil {
		n, err := w.buffer.Write(b)
		if err != nil {
			return n, newLogWriteError("failed to write to buffer", err)
		}
		w.size += n
		w.body = append(w.body, b...)
		return n, nil
	}

	n, err := w.ResponseWriter.Write(b)
	if err != nil {
		return n, newLogWriteError("failed to write response", err)
	}
	w.size += n
	w.body = append(w.body, b...)
	return n, nil
}

// WriteHeader 实现WriteHeader接口
func (w *bodyWriter) WriteHeader(status int) {
	if !w.committed {
		w.status = status
		w.ResponseWriter.WriteHeader(status)
		w.committed = true
	}
}

// Status 获取状态码
func (w *bodyWriter) Status() int {
	return w.status
}

// Size 获取响应大小
func (w *bodyWriter) Size() int {
	return w.size
}

// Body 获取响应体
func (w *bodyWriter) Body() []byte {
	return w.body
}

// Close 关闭writer
func (w *bodyWriter) Close() error {
	if w.buffer != nil {
		if err := w.buffer.Close(); err != nil {
			return newLogBufferError("failed to close buffer", err)
		}
	}

	// 重置并放回池
	w.ResponseWriter = nil
	w.body = w.body[:0]
	w.size = 0
	w.status = 0
	w.committed = false
	w.buffer = nil
	bodyWriterPool.Put(w)

	return nil
}
