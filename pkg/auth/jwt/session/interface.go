package session

import (
	"context"
	"gobase/pkg/logger/types"
	"time"
)

// Store 定义了会话存储的接口
type Store interface {
	// Save 保存会话
	Save(ctx context.Context, session *Session) error
	// Get 获取会话
	Get(ctx context.Context, tokenID string) (*Session, error)
	// Delete 删除会话
	Delete(ctx context.Context, tokenID string) error
	// Refresh 刷新会话过期时间
	Refresh(ctx context.Context, tokenID string, newExpiration time.Time) error
	// Close 关闭存储连接
	Close(ctx context.Context) error
	// Ping 检查存储是否可用
	Ping(ctx context.Context) error
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
