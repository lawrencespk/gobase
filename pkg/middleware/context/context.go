package context

import (
	baseCtx "gobase/pkg/context"
	"gobase/pkg/context/types"
	"gobase/pkg/utils/requestid"

	"github.com/gin-gonic/gin"
)

// Middleware 初始化和注入上下文的中间件
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := baseCtx.NewContext(c.Request.Context())

		// 设置基本信息
		ctx.SetRequestID(requestid.Generate())
		ctx.SetClientIP(c.ClientIP())

		// 将自定义上下文设置到gin.Context
		c.Set("ctx", ctx)

		c.Next()
	}
}

// GetContextFromGin 从gin.Context中获取自定义上下文
func GetContextFromGin(c *gin.Context) types.Context {
	if ctx, exists := c.Get("ctx"); exists {
		if customCtx, ok := ctx.(types.Context); ok {
			return customCtx
		}
	}
	return baseCtx.NewContext(c.Request.Context())
}
