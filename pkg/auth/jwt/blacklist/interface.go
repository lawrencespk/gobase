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
