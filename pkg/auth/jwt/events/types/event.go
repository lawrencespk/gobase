package types

// EventType 定义事件类型
type EventType string

const (
	// EventTypeTokenRevoked 表示令牌被撤销的事件
	EventTypeTokenRevoked EventType = "token_revoked"
	// EventTypeTokenRefreshed 表示令牌被刷新的事件
	EventTypeTokenRefreshed EventType = "token_refreshed"
)
