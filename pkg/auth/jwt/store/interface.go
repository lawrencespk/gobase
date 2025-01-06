package store

import (
	"context"
	"gobase/pkg/auth/jwt"
	"time"
)

// Store 定义JWT存储接口
type Store interface {
	// Set 存储JWT令牌
	Set(ctx context.Context, key string, value *jwt.TokenInfo, expiration time.Duration) error

	// Get 获取JWT令牌
	Get(ctx context.Context, key string) (*jwt.TokenInfo, error)

	// Delete 删除JWT令牌
	Delete(ctx context.Context, key string) error

	// Close 关闭存储连接
	Close() error
}

// StoreType 存储类型
type StoreType string

const (
	// TypeMemory 内存存储类型
	TypeMemory StoreType = "memory"
	// TypeRedis Redis存储类型
	TypeRedis StoreType = "redis"
)

// Config 存储配置
type Config struct {
	// Type 存储类型
	Type StoreType `json:"type" yaml:"type"`
	// Host 主机地址
	Host string `json:"host" yaml:"host"`
	// Port 端口号
	Port int `json:"port" yaml:"port"`
	// Password 密码
	Password string `json:"password" yaml:"password"`
	// DB 数据库
	DB int `json:"db" yaml:"db"`
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

// IsValidType 检查存储类型是否有效
func IsValidType(t string) bool {
	switch StoreType(t) {
	case TypeMemory, TypeRedis:
		return true
	default:
		return false
	}
}
