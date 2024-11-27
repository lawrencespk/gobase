package errors

import (
	"fmt"
	"gobase/pkg/errors/codes"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerOptions 错误处理中间件选项
type ErrorHandlerOptions struct {
	// 是否在响应中包含堆栈信息
	IncludeStack bool
	// 是否在响应中包含详情信息
	IncludeDetails bool
	// 是否记录错误日志
	EnableLogging bool
	// 自定义错误响应格式化函数
	ResponseFormatter func(c *gin.Context, err error) interface{}
}

// DefaultErrorHandlerOptions 默认错误处理选项
var DefaultErrorHandlerOptions = &ErrorHandlerOptions{
	IncludeStack:   false, // 不包含堆栈信息
	IncludeDetails: true,  // 包含详情信息
	EnableLogging:  true,  // 记录错误日志
}

// ErrorHandler 创建错误处理中间件
func ErrorHandler(opts ...*ErrorHandlerOptions) gin.HandlerFunc {
	// 使用默认选项
	options := DefaultErrorHandlerOptions
	if len(opts) > 0 && opts[0] != nil {
		options = opts[0]
	}

	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 处理panic
				err := fmt.Errorf("panic recovered: %v\n%s", r, debug.Stack())
				handleError(c, err, options)
				c.Abort()
			}
		}()

		c.Next()

		// 处理错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err, options)
		}
	}
}

// handleError 处理错误
func handleError(c *gin.Context, err error, opts *ErrorHandlerOptions) {
	// 构造响应
	var response interface{}
	if opts.ResponseFormatter != nil {
		response = opts.ResponseFormatter(c, err)
	} else {
		response = buildErrorResponse(err, opts)
	}

	// 获取HTTP状态码
	status := GetHTTPStatus(err)

	// 记录错误日志
	if opts.EnableLogging {
		logError(c, err, status)
	}

	// 发送响应
	c.JSON(status, response)
}

// buildErrorResponse 构建错误响应
func buildErrorResponse(err error, opts *ErrorHandlerOptions) gin.H {
	response := gin.H{
		"code":    GetErrorCode(err),
		"message": GetErrorMessage(err),
	}

	if opts.IncludeDetails {
		if details := GetErrorDetails(err); len(details) > 0 {
			response["details"] = details
		}
	}

	if opts.IncludeStack {
		if stack := GetErrorStack(err); len(stack) > 0 {
			response["stack"] = stack
		}
	}

	return response
}

// logError 记录错误日志
func logError(c *gin.Context, err error, status int) {
	// TODO: 集成日志系统后完善此函数
	fmt.Printf("[ERROR] %s %s - %d - %v\n",
		c.Request.Method,
		c.Request.URL.Path,
		status,
		err)
}

// NewMiddlewareError 创建中间件错误
func NewMiddlewareError(message string, cause error) error {
	return NewError(codes.MiddlewareError, message, cause)
}

// NewAuthMiddlewareError 创建认证中间件错误
func NewAuthMiddlewareError(message string, cause error) error {
	return NewError(codes.AuthError, message, cause)
}

// NewRateLimitError 创建限流中间件错误
func NewRateLimitError(message string, cause error) error {
	return NewError(codes.RateLimitError, message, cause)
}

// NewTimeoutMiddlewareError 创建超时中间件错误
func NewTimeoutMiddlewareError(message string, cause error) error {
	return NewError(codes.TimeoutMWError, message, cause)
}

// NewCORSError 创建跨域中间件错误
func NewCORSError(message string, cause error) error {
	return NewError(codes.CORSError, message, cause)
}

// NewTracingError 创建追踪中间件错误
func NewTracingError(message string, cause error) error {
	return NewError(codes.TracingError, message, cause)
}

// NewMetricsError 创建指标中间件错误
func NewMetricsError(message string, cause error) error {
	return NewError(codes.MetricsError, message, cause)
}

// NewLoggingError 创建日志中间件错误
func NewLoggingError(message string, cause error) error {
	return NewError(codes.LoggingError, message, cause)
}
