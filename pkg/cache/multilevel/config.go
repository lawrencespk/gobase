package multilevel

import (
	"time"

	"gobase/pkg/errors"
)

// Config 多级缓存配置
type Config struct {
	// L1缓存配置
	L1Config *L1Config

	// L2缓存配置
	L2Config *L2Config

	// L1缓存TTL
	L1TTL time.Duration

	// 是否启用自动预热
	EnableAutoWarmup bool

	// 预热间隔
	WarmupInterval time.Duration

	// 预热并发数
	WarmupConcurrency int
}

// L1Config 一级缓存配置
type L1Config struct {
	// 最大条目数
	MaxEntries int

	// 清理间隔
	CleanupInterval time.Duration
}

// L2Config 二级缓存配置
type L2Config struct {
	// Redis地址
	RedisAddr string

	// Redis密码
	RedisPassword string

	// Redis数据库
	RedisDB int
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.L1Config == nil {
		return errors.NewConfigInvalidError("L1 config is required", nil)
	}
	if c.L2Config == nil {
		return errors.NewConfigInvalidError("L2 config is required", nil)
	}
	if c.L1TTL <= 0 {
		return errors.NewConfigInvalidError("L1 TTL must be positive", nil)
	}
	if c.EnableAutoWarmup && c.WarmupInterval <= 0 {
		return errors.NewConfigInvalidError("warmup interval must be positive", nil)
	}
	return nil
}
