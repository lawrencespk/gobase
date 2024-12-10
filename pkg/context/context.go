package context

import (
	"context"
	"gobase/pkg/context/types"
	"sync"

	"github.com/gin-gonic/gin"
)

// Context 定义了自定义上下文
type Context struct {
	context.Context
	metadata map[string]interface{}
	values   map[string]interface{}
	mu       sync.RWMutex
}

// NewContext 创建新的上下文
func NewContext(parent context.Context) types.Context {
	if parent == nil {
		parent = context.Background()
	}
	return &Context{
		Context:  parent,
		metadata: make(map[string]interface{}),
		values:   make(map[string]interface{}),
	}
}

// Clone 创建当前上下文的副本
func (c *Context) Clone() types.Context {
	c.mu.RLock()
	defer c.mu.RUnlock()

	newCtx := &Context{
		Context:  c.Context,
		metadata: make(map[string]interface{}),
		values:   make(map[string]interface{}),
	}

	// 复制 metadata
	for k, v := range c.metadata {
		newCtx.metadata[k] = v
	}

	// 复制 values
	for k, v := range c.values {
		newCtx.values[k] = v
	}

	return newCtx
}

// GetMetadata 获取所有元数据
func (c *Context) GetMetadata() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	copy := make(map[string]interface{})
	for k, v := range c.metadata {
		copy[k] = v
	}
	return copy
}

// SetMetadata 设置元数据
func (c *Context) SetMetadata(data map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range data {
		c.metadata[k] = v
	}
}

// DeleteMetadata 删除指定键的元数据
func (c *Context) DeleteMetadata(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.metadata, key)
}

// Metadata 返回元数据的副本
func (c *Context) Metadata() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	copy := make(map[string]interface{})
	for k, v := range c.metadata {
		copy[k] = v
	}
	return copy
}

// GetValue 获取指定键的值
func (c *Context) GetValue(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.values[key]
}

// SetValue 设置指定键的值
func (c *Context) SetValue(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

// DeleteValue 删除指定键的值
func (c *Context) DeleteValue(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.values, key)
}

// 上下文键常量
const (
	KeyUserID    = "user_id"
	KeyUserName  = "user_name"
	KeyRequestID = "request_id"
	KeyClientIP  = "client_ip"
	KeyTraceID   = "trace_id"
	KeySpanID    = "span_id"
	KeyError     = "error"
)

// FromGinContext 从gin.Context中获取自定义上下文
func FromGinContext(c *gin.Context) types.Context {
	if ctx, exists := c.Get(types.ContextKey); exists {
		if customCtx, ok := ctx.(types.Context); ok {
			return customCtx
		}
	}

	// 如果获取失败，创建新的上下文
	ctx := types.DefaultNewContext(c.Request.Context())
	c.Set(types.ContextKey, ctx)
	return ctx
}
