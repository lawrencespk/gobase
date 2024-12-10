package types

import (
	"context"
	"time"
)

// Context 定义了自定义上下文接口
type Context interface {
	context.Context // 继承标准库context.Context接口

	// 元数据操作
	GetMetadata() map[string]interface{}
	SetMetadata(metadata map[string]interface{})
	GetValue(key string) interface{}
	SetValue(key string, value interface{})
	DeleteMetadata(key string)
	Metadata() map[string]interface{}

	// 用户信息
	GetUserID() string
	SetUserID(userID string)
	GetUserName() string
	SetUserName(userName string)

	// 请求信息
	GetRequestID() string
	SetRequestID(requestID string)
	GetClientIP() string
	SetClientIP(clientIP string)

	// 追踪信息
	GetTraceID() string
	SetTraceID(traceID string)
	GetSpanID() string
	SetSpanID(spanID string)

	// 错误处理
	GetError() error
	SetError(err error)

	// 上下文控制
	WithCancel() (Context, context.CancelFunc)
	WithTimeout(timeout time.Duration) (Context, context.CancelFunc)
	WithDeadline(deadline time.Time) (Context, context.CancelFunc)

	// 克隆
	Clone() Context

	// 删除值
	DeleteValue(key string)
}

// NewBaseContext 定义创建基础上下文的函数类型
type NewBaseContext func(context.Context) Context

var (
	// DefaultNewContext 默认的上下文创建函数
	DefaultNewContext NewBaseContext
)

// NewContext 创建新的上下文
func NewContext(parent context.Context) Context {
	if DefaultNewContext == nil {
		panic("DefaultNewContext not set")
	}
	if parent == nil {
		parent = context.Background()
	}
	return DefaultNewContext(parent)
}
