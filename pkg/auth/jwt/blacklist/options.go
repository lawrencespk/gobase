package blacklist

import (
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
)

// Options 黑名单配置选项
type Options struct {
	// Redis配置
	Redis *RedisOptions

	// 清理间隔
	CleanupInterval time.Duration

	// 监控指标
	EnableMetrics bool

	// 链路追踪
	EnableTracing bool

	// 日志记录器
	Log types.Logger
}

// RedisOptions Redis配置选项
type RedisOptions struct {
	Addr     string
	Password string
	DB       int
}

// Validate 验证配置选项的有效性
func (o *Options) Validate() error {
	// 验证必填项
	if o == nil {
		return errors.NewConfigInvalidError("options cannot be nil", nil)
	}

	// 验证日志记录器
	if o.Log == nil {
		return errors.NewConfigInvalidError("logger is required", nil)
	}

	// 验证清理间隔
	if o.CleanupInterval < 0 {
		return errors.NewConfigInvalidError("cleanup interval must be non-negative", nil)
	}

	return nil
}
