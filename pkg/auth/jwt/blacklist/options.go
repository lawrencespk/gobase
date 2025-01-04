package blacklist

import (
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/types"
)

// Option 定义选项函数类型
type Option func(*Options) error

// Options 定义黑名单选项
type Options struct {
	// DefaultExpiration 默认过期时间
	DefaultExpiration time.Duration

	// CleanupInterval 清理间隔
	CleanupInterval time.Duration

	// Logger 日志记录器
	Logger types.Logger

	// EnableMetrics 是否启用指标收集
	EnableMetrics bool

	// KeyPrefix Redis key前缀
	KeyPrefix string

	// Log 日志记录器 (兼容旧版本)
	Log types.Logger
}

// DefaultOptions 返回默认选项
func DefaultOptions() *Options {
	return &Options{
		DefaultExpiration: time.Hour * 24,
		CleanupInterval:   time.Hour,
		EnableMetrics:     true,
		KeyPrefix:         "blacklist:", // 设置默认前缀
	}
}

// WithLogger 设置日志记录器
func WithLogger(logger types.Logger) Option {
	return func(o *Options) error {
		if logger == nil {
			return errors.NewError(codes.InvalidParams, "logger cannot be nil", nil)
		}
		o.Logger = logger
		o.Log = logger
		return nil
	}
}

// WithDefaultExpiration 设置默认过期时间
func WithDefaultExpiration(d time.Duration) Option {
	return func(o *Options) error {
		if d <= 0 {
			return errors.NewError(codes.InvalidParams, "default expiration must be positive", nil)
		}
		o.DefaultExpiration = d
		return nil
	}
}

// WithCleanupInterval 设置清理间隔
func WithCleanupInterval(d time.Duration) Option {
	return func(o *Options) error {
		if d <= 0 {
			return errors.NewError(codes.InvalidParams, "cleanup interval must be positive", nil)
		}
		o.CleanupInterval = d
		return nil
	}
}

// WithMetrics 设置是否启用指标收集
func WithMetrics(enable bool) Option {
	return func(o *Options) error {
		o.EnableMetrics = enable
		return nil
	}
}

// ApplyOptions 应用选项
func ApplyOptions(opts *Options, options ...Option) error {
	for _, opt := range options {
		if err := opt(opts); err != nil {
			return err
		}
	}
	return nil
}

// Validate 验证选项
func (o *Options) Validate() error {
	if o.DefaultExpiration <= 0 {
		return errors.NewError(codes.InvalidParams, "default expiration must be positive", nil)
	}
	if o.CleanupInterval <= 0 {
		return errors.NewError(codes.InvalidParams, "cleanup interval must be positive", nil)
	}
	// 确保有默认的key前缀
	if o.KeyPrefix == "" {
		o.KeyPrefix = "blacklist:"
	}
	// 确保日志记录器兼容性
	if o.Logger == nil {
		o.Logger = o.Log
	}
	if o.Log == nil {
		o.Log = o.Logger
	}
	return nil
}
