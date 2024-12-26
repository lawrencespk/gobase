package core

import (
	"context"
	"time"
)

// Limiter 定义限流器接口
type Limiter interface {
	// Allow 检查是否允许请求通过
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error)

	// AllowN 检查是否允许N个请求通过
	AllowN(ctx context.Context, key string, n int64, limit int64, window time.Duration) (bool, error)

	// Wait 等待直到允许请求通过或超时
	Wait(ctx context.Context, key string, limit int64, window time.Duration) error

	// Reset 重置限流器状态
	Reset(ctx context.Context, key string) error
}
