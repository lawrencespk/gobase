package jwt

import (
	"gobase/pkg/middleware/jwt/extractor"
	"gobase/pkg/middleware/jwt/validator"
)

// Option 定义中间件选项函数类型
type Option func(*Options)

// Options 中间件配置选项
type Options struct {
	// 是否启用追踪
	EnableTracing bool
	// 是否启用指标收集
	EnableMetrics bool
	// 超时时间(秒)
	Timeout int
	// Token提取器
	Extractor extractor.TokenExtractor
	// Token验证器
	Validator validator.TokenValidator
}

// NewOptions 创建默认选项
func NewOptions(opts ...Option) *Options {
	// 创建默认的Claims验证器
	defaultValidator := validator.NewClaimsValidator()

	options := &Options{
		EnableTracing: false,
		EnableMetrics: false,
		Timeout:       60,
		// 使用 Authorization header
		Extractor: extractor.NewHeaderExtractor("Authorization", "Bearer "),
		// 使用默认的Claims验证器
		Validator: defaultValidator,
	}

	for _, opt := range opts {
		opt(options)
	}

	return options
}

// WithTracing 设置是否启用追踪
func WithTracing(enable bool) Option {
	return func(o *Options) {
		o.EnableTracing = enable
	}
}

// WithMetrics 设置是否启用指标收集
func WithMetrics(enable bool) Option {
	return func(o *Options) {
		o.EnableMetrics = enable
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout int) Option {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithExtractor 设置Token提取器
func WithExtractor(extractor extractor.TokenExtractor) Option {
	return func(o *Options) {
		o.Extractor = extractor
	}
}

// WithValidator 设置Token验证器
func WithValidator(validator validator.TokenValidator) Option {
	return func(o *Options) {
		o.Validator = validator
	}
}
