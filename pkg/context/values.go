package context

import (
	"gobase/pkg/context/types"
	"time"
)

// GetContextValue 从上下文中获取值
func GetContextValue(ctx types.Context, key string) interface{} {
	if v, ok := ctx.GetMetadata(key); ok {
		return v
	}
	return ctx.Value(key)
}

// SetContextValue 设置上下文值
func SetContextValue(ctx types.Context, key string, value interface{}) {
	ctx.SetMetadata(key, value)
}

// GetUserID 获取用户ID
func GetUserID(ctx types.Context) string {
	return ctx.GetUserID()
}

// GetUserName 获取用户名
func GetUserName(ctx types.Context) string {
	return ctx.GetUserName()
}

// GetRequestID 获取请求ID
func GetRequestID(ctx types.Context) string {
	return ctx.GetRequestID()
}

// GetClientIP 获取客户端IP
func GetClientIP(ctx types.Context) string {
	return ctx.GetClientIP()
}

// GetTraceID 获取追踪ID
func GetTraceID(ctx types.Context) string {
	return ctx.GetTraceID()
}

// GetSpanID 获取Span ID
func GetSpanID(ctx types.Context) string {
	return ctx.GetSpanID()
}

// GetStringValue 获取字符串值
func GetStringValue(ctx types.Context, key string) (string, bool) {
	if v, ok := ctx.GetMetadata(key); ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

// GetIntValue 获取整数值
func GetIntValue(ctx types.Context, key string) (int, bool) {
	if v, ok := ctx.GetMetadata(key); ok {
		if i, ok := v.(int); ok {
			return i, true
		}
	}
	return 0, false
}

// GetBoolValue 获取布尔值
func GetBoolValue(ctx types.Context, key string) (bool, bool) {
	if v, ok := ctx.GetMetadata(key); ok {
		if b, ok := v.(bool); ok {
			return b, true
		}
	}
	return false, false
}

// GetFloat64Value 获取float64值
func GetFloat64Value(ctx types.Context, key string) (float64, bool) {
	if v, ok := ctx.GetMetadata(key); ok {
		if f, ok := v.(float64); ok {
			return f, true
		}
	}
	return 0, false
}

// GetTimeValue 获取时间值
func GetTimeValue(ctx types.Context, key string) (time.Time, bool) {
	if v, ok := ctx.GetMetadata(key); ok {
		if t, ok := v.(time.Time); ok {
			return t, true
		}
	}
	return time.Time{}, false
}

// GetError 获取错误信息
func GetError(ctx types.Context) error {
	return ctx.GetError()
}

// HasError 检查是否存在错误
func HasError(ctx types.Context) bool {
	return ctx.HasError()
}
