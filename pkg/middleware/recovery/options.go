package recovery

import (
	"gobase/pkg/logger/types"
)

// Options Recovery中间件的配置选项
type Options struct {
	// 日志记录器
	Logger types.Logger

	// 是否打印堆栈信息
	PrintStack bool

	// 自定义错误处理函数
	ErrorHandler func(error)

	// 是否禁用错误响应
	DisableErrorResponse bool
}

// Option 定义配置选项函数类型
type Option func(*Options)

// WithLogger 设置日志记录器
func WithLogger(logger types.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// WithPrintStack 设置是否打印堆栈信息
func WithPrintStack(print bool) Option {
	return func(o *Options) {
		o.PrintStack = print
	}
}

// WithErrorHandler 设置自定义错误处理函数
func WithErrorHandler(handler func(error)) Option {
	return func(o *Options) {
		o.ErrorHandler = handler
	}
}

// WithDisableErrorResponse 设置是否禁用错误响应
func WithDisableErrorResponse(disable bool) Option {
	return func(o *Options) {
		o.DisableErrorResponse = disable
	}
}

// defaultOptions 返回默认配置
func defaultOptions() *Options {
	return &Options{
		Logger:               nil,   // 默认logger
		PrintStack:           true,  // 默认打印堆栈信息
		DisableErrorResponse: false, // 默认不禁用错误响应
	}
}
