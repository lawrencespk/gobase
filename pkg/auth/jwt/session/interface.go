package session

import (
	"context"
	"gobase/pkg/logger/types"
	"time"
)

// Store 会话存储接口
type Store interface {
	// Get 获取会话数据
	Get(ctx context.Context, key string) (string, error)

	// Set 设置会话数据
	Set(ctx context.Context, key string, value string, expiration time.Duration) error

	// Delete 删除会话数据
	Delete(ctx context.Context, key string) error

	// Close 关闭存储连接
	Close() error
}

// Options 存储配置选项
type Options struct {
	// Redis配置
	Redis *RedisOptions
	// 键前缀
	KeyPrefix string
	// 是否启用监控
	EnableMetrics bool
	// 日志记录器
	Log types.Logger
}

// RedisOptions Redis配置
type RedisOptions struct {
	Addr     string
	Password string
	DB       int
}
