package context

import (
	"gobase/pkg/context/types"
	"gobase/pkg/utils/requestid"

	"github.com/gin-gonic/gin"
)

// Options 中间件配置选项
type Options struct {
	// 请求ID生成器配置
	RequestIDOptions *requestid.Options
	// 是否在响应头中设置请求ID
	SetRequestIDHeader bool
	// 请求ID的响应头名称
	RequestIDHeaderName string
}

// DefaultOptions 返回默认配置
func DefaultOptions() *Options {
	return &Options{
		RequestIDOptions:    requestid.DefaultOptions(),
		SetRequestIDHeader:  true,
		RequestIDHeaderName: "X-Request-ID",
	}
}

// Middleware 创建上下文中间件
func Middleware(opts *Options) gin.HandlerFunc {
	if opts == nil {
		opts = DefaultOptions()
	}

	// 创建请求ID生成器
	generator := requestid.NewGenerator(opts.RequestIDOptions)

	return func(c *gin.Context) {
		// 从请求头获取请求ID，如果没有则生成新的
		reqID := c.GetHeader(opts.RequestIDHeaderName)
		if reqID == "" {
			reqID = generator.Generate()
		}

		// 创建请求上下文
		ctx := types.NewContext(c.Request.Context())

		// 设置基本信息
		ctx.SetRequestID(reqID)
		ctx.SetClientIP(c.ClientIP())

		// 设置响应头
		if opts.SetRequestIDHeader {
			c.Header(opts.RequestIDHeaderName, reqID)
		}

		// 将上下文保存到gin.Context中
		c.Set(types.ContextKey, ctx)

		// 将上下文设置到请求中
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// GetContextFromGin 从gin.Context中获取自定义上下文
func GetContextFromGin(c *gin.Context) types.Context {
	if ctx, exists := c.Get(types.ContextKey); exists {
		if customCtx, ok := ctx.(types.Context); ok {
			return customCtx
		}
	}

	// 如果获取失败，创建新的上下文
	ctx := types.NewContext(c.Request.Context())
	c.Set(types.ContextKey, ctx)
	return ctx
}
