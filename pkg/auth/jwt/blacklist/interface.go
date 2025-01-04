package blacklist

import (
	"context"
	"time"
)

// TokenBlacklist 定义Token黑名单接口
type TokenBlacklist interface {
	// Add 将Token添加到黑名单
	Add(ctx context.Context, token string, expiration time.Duration) error

	// IsBlacklisted 检查Token是否在黑名单中
	IsBlacklisted(ctx context.Context, token string) (bool, error)

	// Remove 从黑名单中移除Token
	Remove(ctx context.Context, token string) error

	// Clear 清空黑名单
	Clear(ctx context.Context) error

	// Close 关闭黑名单
	Close() error
}

// Store 定义黑名单存储接口
type Store interface {
	// Add 添加token到黑名单
	Add(ctx context.Context, tokenID, reason string, expiration time.Duration) error

	// Get 获取token的黑名单原因
	Get(ctx context.Context, tokenID string) (string, error)

	// Remove 从黑名单中移除token
	Remove(ctx context.Context, tokenID string) error

	// Close 关闭存储
	Close() error
}
