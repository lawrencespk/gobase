package core

import (
	"context"
	"time"
)

// RedisConfig Redis配置
type RedisConfig struct {
	// 基础配置
	Addresses []string `json:"addresses" yaml:"addresses"` // Redis地址列表
	Username  string   `json:"username" yaml:"username"`   // 用户名
	Password  string   `json:"password" yaml:"password"`   // 密码
	Database  int      `json:"database" yaml:"database"`   // 数据库索引

	// 连接池配置
	PoolSize   int `json:"poolSize" yaml:"poolSize"`     // 连接池大小
	MaxRetries int `json:"maxRetries" yaml:"maxRetries"` // 最大重试次数

	// 超时配置
	DialTimeout  time.Duration `json:"dialTimeout" yaml:"dialTimeout"`   // 连接超时
	ReadTimeout  time.Duration `json:"readTimeout" yaml:"readTimeout"`   // 读取超时
	WriteTimeout time.Duration `json:"writeTimeout" yaml:"writeTimeout"` // 写入超时
}

// Limiter 限流器接口
type Limiter interface {
	// Allow 判断是否允许通过
	// key: 限流标识
	// limit: 限流阈值
	// window: 时间窗口
	Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error)

	// AllowN 判断是否允许N个请求通过
	AllowN(ctx context.Context, key string, n int64, limit int64, window time.Duration) (bool, error)

	// Wait 等待直到允许通过或超时
	Wait(ctx context.Context, key string, limit int64, window time.Duration) error

	// Reset 重置限流器
	Reset(ctx context.Context, key string) error
}

// LimiterOption 限流器配置选项
type LimiterOption func(*LimiterOptions)

// LimiterOptions 限流器配置
type LimiterOptions struct {
	// 限流器名称
	Name string

	// 限流算法类型
	Algorithm string

	// Redis配置(如果使用Redis限流器)
	RedisConfig *RedisConfig

	// 监控配置
	EnableMetrics bool
	MetricsPrefix string

	// 链路追踪配置
	EnableTracing bool
}
