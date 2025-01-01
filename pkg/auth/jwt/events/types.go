package events

import (
	"time"
)

// EventType 事件类型
type EventType string

const (
	// TokenRevoked Token吊销事件
	TokenRevoked EventType = "token_revoked"
	// TokenExpired Token过期事件
	TokenExpired EventType = "token_expired"
	// KeyRotated 密钥轮换事件
	KeyRotated EventType = "key_rotated"
	// SessionCreated 会话创建事件
	SessionCreated EventType = "session_created"
	// SessionDestroyed 会话销毁事件
	SessionDestroyed EventType = "session_destroyed"
)

// Event JWT相关事件
type Event struct {
	// 事件ID
	ID string `json:"id"`
	// 事件类型
	Type EventType `json:"type"`
	// 事件时间
	Timestamp time.Time `json:"timestamp"`
	// 用户ID
	UserID string `json:"user_id,omitempty"`
	// Token ID
	TokenID string `json:"token_id,omitempty"`
	// 会话ID
	SessionID string `json:"session_id,omitempty"`
	// 密钥ID
	KeyID string `json:"key_id,omitempty"`
	// 事件负载
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// EventHandler 事件处理器
type EventHandler func(event *Event) error
