package jwt

import (
	"github.com/gin-gonic/gin"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/errors"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
	"gobase/pkg/middleware/jwt/context"
	"gobase/pkg/middleware/jwt/extractor"
)

// Middleware JWT中间件结构
type Middleware struct {
	// JWT token管理器
	tokenManager *jwt.TokenManager
	// token提取器
	extractor extractor.TokenExtractor
	// 日志记录器
	logger types.Logger
	// 配置选项
	opts *Options
}

// New 创建新的JWT中间件
func New(tokenManager *jwt.TokenManager, opts ...Option) (*Middleware, error) {
	// 创建默认日志记录器
	log, err := logger.NewLogger(
		logger.WithLevel(types.InfoLevel),
		logger.WithOutputPaths([]string{"stdout"}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create logger")
	}

	// 创建默认中间件实例
	m := &Middleware{
		tokenManager: tokenManager,
		logger:       log,
		opts:         NewOptions(opts...),
	}

	// 设置token提取器
	if m.opts.Extractor != nil {
		m.extractor = m.opts.Extractor
	} else {
		// 默认使用链式提取器
		m.extractor = extractor.ChainExtractor{
			extractor.NewHeaderExtractor("Authorization", "Bearer "),
			extractor.NewCookieExtractor("jwt"),
			extractor.NewQueryExtractor("token"),
		}
	}

	return m, nil
}

// Handle 实现gin中间件处理函数
func (m *Middleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始处理计时
		start := StartTimer()
		defer ObserveTokenValidationDuration(start)

		// 提取token
		tokenStr, err := m.extractor.Extract(c)
		if err != nil {
			IncTokenValidationError("extract_failed")
			m.handleError(c, errors.NewTokenNotFoundError("failed to extract token", err))
			return
		}

		// 验证token
		token, err := m.tokenManager.ValidateToken(c.Request.Context(), tokenStr)
		if err != nil {
			IncTokenValidationError("validation_failed")
			m.handleError(c, errors.NewTokenInvalidError("failed to validate token", err))
			return
		}

		// 转换claims类型
		claims, ok := token.Claims.(*jwt.StandardClaims)
		if !ok {
			IncTokenValidationError("invalid_claims_type")
			m.handleError(c, errors.NewClaimsInvalidError("invalid claims type", nil))
			return
		}

		// 设置token到上下文，供后续验证器使用
		c.Set("jwt_token", tokenStr)

		// 执行验证器链
		if m.opts.Validator != nil {
			if err := m.opts.Validator.Validate(c, claims); err != nil {
				IncTokenValidationError("validation_chain_failed")
				m.handleError(c, err)
				return
			}
		}

		// 将claims存储到gin.Context中
		jwt.ToContext(c, claims)

		// 设置请求上下文
		c.Request = c.Request.WithContext(context.WithJWTContext(
			c.Request.Context(),
			claims,
			tokenStr,
		))

		// 继续处理请求
		c.Next()
	}
}

// handleError 处理中间件错误
func (m *Middleware) handleError(c *gin.Context, err error) {
	// 记录错误日志
	m.logger.WithError(err).Error(c.Request.Context(), "jwt middleware error")

	// 根据错误类型设置不同的响应
	switch {
	case jwt.IsTokenExpiredError(err) || errors.IsTokenExpiredError(err):
		c.AbortWithStatusJSON(401, gin.H{
			"code":    errors.GetErrorCode(err),
			"message": "token has expired",
		})
	case errors.IsTokenInvalidError(err):
		c.AbortWithStatusJSON(401, gin.H{
			"code":    errors.GetErrorCode(err),
			"message": "invalid token",
		})
	case errors.IsSignatureInvalidError(err):
		c.AbortWithStatusJSON(401, gin.H{
			"code":    errors.GetErrorCode(err),
			"message": "invalid token signature",
		})
	default:
		c.AbortWithStatusJSON(401, gin.H{
			"code":    errors.GetErrorCode(err),
			"message": "unauthorized",
		})
	}
}

// MustNew 创建新的JWT中间件，如果出错则panic
func MustNew(tokenManager *jwt.TokenManager, opts ...Option) *Middleware {
	m, err := New(tokenManager, opts...)
	if err != nil {
		panic(err)
	}
	return m
}
