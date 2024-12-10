package types

// 定义上下文键的常量，从 values.go 移动到这里
const (
	// ContextKey 用于在gin.Context中存储自定义上下文的键
	ContextKey = "custom_context"

	// 用户相关键
	KeyUserID    = "user_id"    // 用户ID
	KeyUserName  = "user_name"  // 用户名
	KeyRequestID = "request_id" // 请求ID
	KeyClientIP  = "client_ip"  // 客户端IP
	KeyTraceID   = "trace_id"   // 追踪ID
	KeySpanID    = "span_id"    // SpanID
	KeyError     = "error"      // 错误信息
)
