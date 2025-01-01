package session

import "time"

// Session JWT会话数据
type Session struct {
	// UserID 用户ID
	UserID string `json:"user_id"`

	// TokenID 令牌ID
	TokenID string `json:"token_id"`

	// ExpiresAt 过期时间
	ExpiresAt time.Time `json:"expires_at"`

	// CreatedAt 创建时间
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`

	// Metadata 元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}
