package recovery

import (
	"fmt"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"gobase/pkg/context/types"
	"gobase/pkg/errors"
	"gobase/pkg/logger"
	loggerTypes "gobase/pkg/logger/types"
	contextMiddleware "gobase/pkg/middleware/context"
)

// Recovery 创建一个Recovery中间件
func Recovery(opts ...Option) gin.HandlerFunc {
	// 合并配置选项
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	// 如果没有设置logger，使用默认logger
	if options.Logger == nil {
		options.Logger = logger.GetLogger()
	}

	// 返回Recovery中间件
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取自定义上下文
				ctx := contextMiddleware.GetContextFromGin(c)

				// 构建错误信息
				stack := string(debug.Stack())
				errMsg := fmt.Sprintf("panic recovered: %v", err)

				// 创建系统错误
				sysErr := errors.NewSystemError(errMsg, nil)

				// 记录错误到上下文
				ctx.SetError(sysErr)

				// 记录日志
				logError(ctx, sysErr, stack, options)

				// 调用自定义错误处理函数
				if options.ErrorHandler != nil {
					options.ErrorHandler(sysErr)
				}

				// 返回错误响应
				if !options.DisableErrorResponse {
					status := errors.GetHTTPStatus(sysErr)
					c.AbortWithStatusJSON(status, gin.H{
						"code":    errors.GetErrorCode(sysErr),
						"message": errors.GetErrorMessage(sysErr),
					})
				}
			}
		}()

		c.Next()
	}
}

// logError 记录错误日志
func logError(ctx types.Context, err error, stack string, opts *Options) {
	fields := []loggerTypes.Field{
		{
			Key:   "error",
			Value: err.Error(),
		},
		{
			Key:   "error_code",
			Value: errors.GetErrorCode(err),
		},
	}

	// 如果启用了堆栈跟踪，则添加堆栈跟踪信息
	if opts.PrintStack {
		fields = append(fields, loggerTypes.Field{
			Key:   "stack",
			Value: stack,
		})
	}

	// 添加请求相关信息
	if requestID, exists := ctx.Value(types.KeyRequestID).(string); exists {
		fields = append(fields, loggerTypes.Field{
			Key:   "request_id",
			Value: requestID,
		})
	}

	// 如果启用了客户端IP跟踪，则添加客户端IP信息
	if clientIP, exists := ctx.Value(types.KeyClientIP).(string); exists {
		fields = append(fields, loggerTypes.Field{
			Key:   "client_ip",
			Value: clientIP,
		})
	}

	opts.Logger.Error(ctx, "Panic recovered", fields...)
}
