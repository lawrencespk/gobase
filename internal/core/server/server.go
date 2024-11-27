package server

import (
	"gobase/pkg/errors"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()

	// 注册错误处理中间件
	r.Use(errors.ErrorHandler())

	// ... 其他中间件和路由配置

	return r
}
