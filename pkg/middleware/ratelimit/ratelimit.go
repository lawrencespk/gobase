package ratelimit

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"gobase/pkg/errors"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
	"gobase/pkg/ratelimit/core"
)

// Config 限流中间件配置
type Config struct {
	// 限流器实例
	Limiter core.Limiter
	// 限流键生成函数
	KeyFunc func(*gin.Context) string
	// 限流阈值
	Limit int64
	// 时间窗口
	Window time.Duration
	// 是否启用等待模式(true:等待直到允许通过, false:直接返回429)
	WaitMode bool
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP() // 默认使用客户端IP作为限流键
		},
		Limit:    100,         // 默认100个请求
		Window:   time.Minute, // 默认1分钟窗口
		WaitMode: false,       // 默认不等待
		// Limiter 必须由用户提供，因为它需要依赖外部存储
	}
}

// RateLimit 创建限流中间件
func RateLimit(cfg *Config) gin.HandlerFunc {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// 验证必要的配置
	if cfg.Limiter == nil {
		panic("ratelimit middleware requires a limiter instance")
	}

	// 确保有默认的KeyFunc
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = func(c *gin.Context) string {
			return c.ClientIP() // 默认使用客户端IP作为限流键
		}
	}

	// 获取logger
	log := logger.GetLogger().WithFields(
		types.Field{Key: "module", Value: "middleware"},
		types.Field{Key: "component", Value: "ratelimit"},
	)

	return func(c *gin.Context) {
		// 生成限流键
		key := cfg.KeyFunc(c)

		var allowed bool
		var err error

		if cfg.WaitMode {
			// 等待模式
			err = cfg.Limiter.Wait(c.Request.Context(), key, cfg.Limit, cfg.Window)
			allowed = err == nil
		} else {
			// 非等待模式
			allowed, err = cfg.Limiter.Allow(c.Request.Context(), key, cfg.Limit, cfg.Window)
		}

		if err != nil {
			log.Error(c.Request.Context(), "rate limit check failed",
				types.Field{Key: "error", Value: err},
				types.Field{Key: "key", Value: key},
			)
			c.AbortWithError(http.StatusInternalServerError, errors.Wrap(err, "rate limit check failed"))
			return
		}

		if !allowed {
			log.Warn(c.Request.Context(), "rate limit exceeded",
				types.Field{Key: "key", Value: key},
				types.Field{Key: "limit", Value: cfg.Limit},
				types.Field{Key: "window", Value: cfg.Window},
			)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":  "rate limit exceeded",
				"limit":  cfg.Limit,
				"window": cfg.Window.String(),
			})
			return
		}

		c.Next()
	}
}
