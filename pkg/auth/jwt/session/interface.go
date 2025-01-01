package session

import (
	"context"
)

// Store 会话存储接口
type Store interface {
	// Save 保存会话
	Save(ctx context.Context, session *Session) error

	// Get 获取会话
	Get(ctx context.Context, sessionID string) (*Session, error)

	// Delete 删除会话
	Delete(ctx context.Context, sessionID string) error

	// DeleteByUserID 删除用户的所有会话
	DeleteByUserID(ctx context.Context, userID string) error

	// ListByUserID 获取用户的所有会话
	ListByUserID(ctx context.Context, userID string) ([]*Session, error)

	// Count 获取用户的会话数量
	Count(ctx context.Context, userID string) (int, error)

	// Close 关闭存储
	Close() error
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	// 设备ID
	ID string
	// 设备类型
	Type string
	// 设备名称
	Name string
	// 操作系统
	OS string
	// 浏览器
	Browser string
	// IP地址
	IPAddress string
}
