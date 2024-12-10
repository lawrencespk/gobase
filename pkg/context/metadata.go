package context

import (
	"context"
	"gobase/pkg/context/types"
	"time"
)

// GetUserID 获取用户ID
func (c *Context) GetUserID() string {
	if val := c.GetValue(KeyUserID); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// SetUserID 设置用户ID
func (c *Context) SetUserID(userID string) {
	c.SetValue(KeyUserID, userID)
}

// GetUserName 获取用户名
func (c *Context) GetUserName() string {
	if val := c.GetValue(KeyUserName); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// SetUserName 设置用户名
func (c *Context) SetUserName(userName string) {
	c.SetValue(KeyUserName, userName)
}

// GetRequestID 获取请求ID
func (c *Context) GetRequestID() string {
	if val := c.GetValue(KeyRequestID); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// SetRequestID 设置请求ID
func (c *Context) SetRequestID(requestID string) {
	c.SetValue(KeyRequestID, requestID)
}

// GetClientIP 获取客户端IP
func (c *Context) GetClientIP() string {
	if val := c.GetValue(KeyClientIP); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// SetClientIP 设置客户端IP
func (c *Context) SetClientIP(clientIP string) {
	c.SetValue(KeyClientIP, clientIP)
}

// GetTraceID 获取追踪ID
func (c *Context) GetTraceID() string {
	if val := c.GetValue(KeyTraceID); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// SetTraceID 设置追踪ID
func (c *Context) SetTraceID(traceID string) {
	c.SetValue(KeyTraceID, traceID)
}

// GetSpanID 获取跨度ID
func (c *Context) GetSpanID() string {
	if val := c.GetValue(KeySpanID); val != nil {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}

// SetSpanID 设置跨度ID
func (c *Context) SetSpanID(spanID string) {
	c.SetValue(KeySpanID, spanID)
}

// GetError 获取错误信息
func (c *Context) GetError() error {
	if val := c.GetValue(KeyError); val != nil {
		if err, ok := val.(error); ok {
			return err
		}
	}
	return nil
}

// SetError 设置错误信息
func (c *Context) SetError(err error) {
	c.SetValue(KeyError, err)
}

// WithCancel 创建可取消的上下文
func (c *Context) WithCancel() (types.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c.Context)
	newCtx := c.Clone()
	newCtx.(*Context).Context = ctx
	return newCtx, cancel
}

// WithTimeout 创建带超时的上下文
func (c *Context) WithTimeout(timeout time.Duration) (types.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.Context, timeout)
	newCtx := c.Clone()
	newCtx.(*Context).Context = ctx
	return newCtx, cancel
}

// WithDeadline 创建带截止时间的上下文
func (c *Context) WithDeadline(deadline time.Time) (types.Context, context.CancelFunc) {
	ctx, cancel := context.WithDeadline(c.Context, deadline)
	newCtx := c.Clone()
	newCtx.(*Context).Context = ctx
	return newCtx, cancel
}
