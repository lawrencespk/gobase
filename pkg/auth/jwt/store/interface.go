package store

import (
	"context"
	"time"
)

// Store 定义JWT存储接口
type Store interface {
	// Set 存储JWT令牌
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Get 获取JWT令牌
	Get(ctx context.Context, key string) (string, error)

	// Delete 删除JWT令牌
	Delete(ctx context.Context, key string) error

	// Close 关闭存储连接
	Close() error
}

// Options 存储配置选项
type Options struct {
	// Redis配置
	Redis *RedisOptions

	// 监控指标
	EnableMetrics bool

	// 链路追踪
	EnableTracing bool

	// 前缀
	KeyPrefix string

	// CleanupInterval 清理过期会话的时间间隔
	// 如果设置为0或负值，则不进行自动清理
	// 仅用于内存存储模式
	CleanupInterval time.Duration
}

// RedisOptions Redis配置选项
type RedisOptions struct {
	Addr     string
	Password string
	DB       int
}
