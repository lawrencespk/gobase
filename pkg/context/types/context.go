package types

import (
	"context"
	"time"
)

// Context 扩展标准库的 context.Context 接口
type Context interface {
	context.Context // 继承标准库的 Context 接口

	// 元数据相关
	SetMetadata(key string, value interface{})  // 设置元数据
	GetMetadata(key string) (interface{}, bool) // 获取元数据
	Metadata() map[string]interface{}           // 获取所有元数据
	DeleteMetadata(key string)                  // 删除元数据
	SetMetadataMap(data map[string]interface{}) // 设置多个元数据

	// 用户信息相关
	SetUserID(userID string)     // 设置用户ID
	GetUserID() string           // 获取用户ID
	SetUserName(userName string) // 设置用户名
	GetUserName() string         // 获取用户名

	// 请求相关
	SetRequestID(requestID string) // 设置请求ID
	GetRequestID() string          // 获取请求ID
	SetClientIP(clientIP string)   // 设置客户端IP
	GetClientIP() string           // 获取客户端IP

	// 追踪相关
	SetTraceID(traceID string) // 设置追踪ID
	GetTraceID() string        // 获取追踪ID
	SetSpanID(spanID string)   // 设置SpanID
	GetSpanID() string         // 获取SpanID

	// 超时控制
	WithTimeout(timeout time.Duration) (Context, context.CancelFunc) // 设置超时时间
	WithDeadline(deadline time.Time) (Context, context.CancelFunc)   // 设置截止时间
	WithCancel() (Context, context.CancelFunc)                       // 创建一个可取消的上下文

	// 错误处理
	SetError(err error) // 设置错误信息
	GetError() error    // 获取错误信息
	HasError() bool     // 检查是否存在错误

	// 克隆当前上下文
	Clone() Context
}
