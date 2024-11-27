package context

import (
	"context"
	"sync"
	"time"

	"gobase/pkg/context/types"
)

// baseContext 实现 types.Context 接口
type baseContext struct {
	context.Context
	metadata map[string]interface{}
	mu       sync.RWMutex
}

// NewContext 创建一个新的 Context
func NewContext(ctx context.Context) types.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return &baseContext{
		Context:  ctx,
		metadata: make(map[string]interface{}),
	}
}

// SetMetadata 设置元数据
func (c *baseContext) SetMetadata(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metadata[key] = value
}

// GetMetadata 获取元数据
func (c *baseContext) GetMetadata(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.metadata[key]
	return value, ok
}

// Metadata 获取所有元数据
func (c *baseContext) Metadata() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	metadata := make(map[string]interface{}, len(c.metadata))
	for k, v := range c.metadata {
		metadata[k] = v
	}
	return metadata
}

// SetUserID 设置用户ID
func (c *baseContext) SetUserID(userID string) {
	c.SetMetadata(types.KeyUserID, userID)
}

// GetUserID 获取用户ID
func (c *baseContext) GetUserID() string {
	if v, ok := c.GetMetadata(types.KeyUserID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// SetUserName 设置用户名
func (c *baseContext) SetUserName(userName string) {
	c.SetMetadata(types.KeyUserName, userName)
}

// GetUserName 获取用户名
func (c *baseContext) GetUserName() string {
	if v, ok := c.GetMetadata(types.KeyUserName); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// SetRequestID 设置请求ID
func (c *baseContext) SetRequestID(requestID string) {
	c.SetMetadata(types.KeyRequestID, requestID)
}

// GetRequestID 获取请求ID
func (c *baseContext) GetRequestID() string {
	if v, ok := c.GetMetadata(types.KeyRequestID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// SetClientIP 设置客户端IP
func (c *baseContext) SetClientIP(clientIP string) {
	c.SetMetadata(types.KeyClientIP, clientIP)
}

// GetClientIP 获取客户端IP
func (c *baseContext) GetClientIP() string {
	if v, ok := c.GetMetadata(types.KeyClientIP); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// SetTraceID 设置追踪ID
func (c *baseContext) SetTraceID(traceID string) {
	c.SetMetadata(types.KeyTraceID, traceID)
}

// GetTraceID 获取追踪ID
func (c *baseContext) GetTraceID() string {
	if v, ok := c.GetMetadata(types.KeyTraceID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// SetSpanID 设置Span ID
func (c *baseContext) SetSpanID(spanID string) {
	c.SetMetadata(types.KeySpanID, spanID)
}

// GetSpanID 获取Span ID
func (c *baseContext) GetSpanID() string {
	if v, ok := c.GetMetadata(types.KeySpanID); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// WithTimeout 创建一个带超时的新上下文
func (c *baseContext) WithTimeout(timeout time.Duration) (types.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.Context, timeout)
	return &baseContext{
		Context:  ctx,
		metadata: c.Metadata(), // 复制元数据
	}, cancel
}

// WithDeadline 创建一个带截止时间的新上下文
func (c *baseContext) WithDeadline(deadline time.Time) (types.Context, context.CancelFunc) {
	ctx, cancel := context.WithDeadline(c.Context, deadline)
	return &baseContext{
		Context:  ctx,
		metadata: c.Metadata(), // 复制元数据
	}, cancel
}

// WithCancel 创建一个可取消的新上下文
func (c *baseContext) WithCancel() (types.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c.Context)
	return &baseContext{
		Context:  ctx,
		metadata: c.Metadata(), // 复制元数据
	}, cancel
}

// Clone 克隆当前上下文
func (c *baseContext) Clone() types.Context {
	return &baseContext{
		Context:  c.Context,
		metadata: c.Metadata(), // 复制元数据
	}
}

// DeleteMetadata 删除元数据
func (c *baseContext) DeleteMetadata(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.metadata, key)
}

// SetMetadataMap 设置多个元数据
func (c *baseContext) SetMetadataMap(data map[string]interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range data {
		c.metadata[k] = v
	}
}

// SetError 设置错误信息
func (c *baseContext) SetError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err == nil {
		delete(c.metadata, types.KeyError) // 当 err 为 nil 时，删除错误
	} else {
		c.metadata[types.KeyError] = err
	}
}

// GetError 获取错误信息
func (c *baseContext) GetError() error {
	if v, ok := c.GetMetadata(types.KeyError); ok {
		if err, ok := v.(error); ok {
			return err
		}
	}
	return nil
}

// HasError 检查是否存在错误
func (c *baseContext) HasError() bool {
	return c.GetError() != nil
}
