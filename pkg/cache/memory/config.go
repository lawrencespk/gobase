package memory

import (
	"time"

	"gobase/pkg/errors"
)

// Config 内存缓存配置
type Config struct {
	// 最大条目数
	MaxEntries int

	// 清理间隔
	CleanupInterval time.Duration

	// 默认过期时间
	DefaultTTL time.Duration
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c == nil {
		return errors.NewConfigInvalidError("config cannot be nil", nil)
	}

	if c.MaxEntries <= 0 {
		return errors.NewConfigInvalidError("max entries must be positive", nil)
	}

	if c.CleanupInterval <= 0 {
		return errors.NewConfigInvalidError("cleanup interval must be positive", nil)
	}

	if c.DefaultTTL <= 0 {
		return errors.NewConfigInvalidError("default TTL must be positive", nil)
	}

	return nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxEntries:      10000,
		CleanupInterval: time.Minute,
		DefaultTTL:      time.Hour,
	}
}
