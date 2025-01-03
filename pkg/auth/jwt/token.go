package jwt

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"gobase/pkg/errors"
	errorTypes "gobase/pkg/errors/types"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metrics"
	"gobase/pkg/trace/jaeger"
)

// TokenManager JWT token管理器
type TokenManager struct {
	secretKey  []byte
	logger     types.Logger
	provider   *jaeger.Provider
	metrics    bool
	claimsPool sync.Pool
}

// TokenManagerOption 定义 TokenManager 的选项
type TokenManagerOption func(*TokenManager)

// WithoutTracing 禁用链路追踪
func WithoutTracing() TokenManagerOption {
	return func(tm *TokenManager) {
		tm.provider = nil
	}
}

// WithLogger 设置logger
func WithLogger(logger types.Logger) TokenManagerOption {
	return func(tm *TokenManager) {
		tm.logger = logger
	}
}

// WithoutMetrics 禁用指标收集
func WithoutMetrics() TokenManagerOption {
	return func(tm *TokenManager) {
		tm.metrics = false
	}
}

// NewTokenManager 创建新的token管理器
func NewTokenManager(secretKey string, opts ...TokenManagerOption) (*TokenManager, error) {
	// 创建logger
	log, err := logger.NewLogger(
		logger.WithLevel(types.InfoLevel),
		logger.WithOutputPaths([]string{"stdout"}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create logger")
	}

	tm := &TokenManager{
		secretKey: []byte(secretKey),
		logger:    log,
		metrics:   true,
		claimsPool: sync.Pool{
			New: func() interface{} {
				return &StandardClaims{}
			},
		},
	}

	// 默认创建 jaeger provider
	if provider, err := jaeger.NewProvider(); err == nil {
		tm.provider = provider
	}

	// 应用选项
	for _, opt := range opts {
		opt(tm)
	}

	return tm, nil
}

// GenerateToken 生成令牌
func (tm *TokenManager) GenerateToken(ctx context.Context, claims Claims) (string, error) {
	start := time.Now()
	defer func() {
		if tm.metrics {
			metrics.DefaultJWTMetrics.TokenDuration.WithLabelValues("generate").Observe(time.Since(start).Seconds())
		}
	}()

	// 验证claims
	if err := claims.Validate(); err != nil {
		if tm.metrics {
			metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("generate", err.Error()).Inc()
		}
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(tm.secretKey)
	if err != nil {
		if tm.metrics {
			metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("generate", err.Error()).Inc()
		}
		return "", errors.NewTokenGenerationError("failed to generate token", err)
	}

	return tokenString, nil
}

// ValidateToken 验证令牌
func (tm *TokenManager) ValidateToken(ctx context.Context, tokenString string) (*jwt.Token, error) {
	start := time.Now()
	defer func() {
		if tm.metrics {
			metrics.DefaultJWTMetrics.TokenDuration.WithLabelValues("validate").Observe(time.Since(start).Seconds())
		}
	}()

	// 从对象池获取claims对象
	claims := tm.claimsPool.Get().(*StandardClaims)
	defer tm.claimsPool.Put(claims)

	// 使用预分配的claims对象
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.NewSignatureInvalidError("invalid signing method", nil)
		}
		return tm.secretKey, nil
	})

	if err != nil {
		if tm.metrics {
			metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("validate", err.Error()).Inc()
		}
		return nil, tm.HandleValidationError(err)
	}

	return token, nil
}

// ParseToken 解析JWT token
func (tm *TokenManager) ParseToken(ctx context.Context, tokenString string) (jwt.Claims, error) {
	// 验证token
	token, err := tm.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	// 获取claims
	if token.Claims == nil {
		return nil, errors.NewClaimsInvalidError("invalid token claims", nil)
	}

	return token.Claims, nil
}

// HandleValidationError 处理token验证错误
func (tm *TokenManager) HandleValidationError(tokenErr error) error {
	span, spanErr := jaeger.NewSpan("TokenManager.handleValidationError")
	if spanErr != nil {
		return errors.Wrap(spanErr, "failed to create span")
	}
	defer span.Finish()

	// 添加调试日志
	tm.logger.Debug(span.Context(), "handling validation error",
		types.Field{Key: "error", Value: tokenErr},
		types.Field{Key: "error_type", Value: fmt.Sprintf("%T", tokenErr)},
	)

	var err error
	switch {
	case errors.Is(tokenErr, jwt.ErrTokenMalformed):
		span.SetTag("error.reason", "token_malformed")
		if tm.metrics {
			metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("validate", "malformed").Inc()
		}
		err = errors.NewTokenInvalidError("token is malformed", tokenErr)
		tm.logger.Debug(span.Context(), "mapped to token invalid error",
			types.Field{Key: "error_code", Value: "TokenInvalid"})
	case errors.Is(tokenErr, jwt.ErrTokenExpired) || strings.Contains(tokenErr.Error(), "token is expired by"):
		span.SetTag("error.reason", "token_expired")
		if tm.metrics {
			metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("validate", "expired").Inc()
		}
		err = errors.NewTokenExpiredError("token has expired", tokenErr)
		tm.logger.Debug(span.Context(), "mapped to token expired error",
			types.Field{Key: "error_code", Value: "TokenExpired"})
	case errors.Is(tokenErr, jwt.ErrSignatureInvalid) ||
		strings.Contains(strings.ToLower(tokenErr.Error()), "signature is invalid") ||
		strings.Contains(strings.ToLower(tokenErr.Error()), "invalid signature"):
		span.SetTag("error.reason", "invalid_signature")
		if tm.metrics {
			metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("validate", "invalid_signature").Inc()
		}
		err = errors.NewSignatureInvalidError("invalid token signature", tokenErr)
		tm.logger.Debug(span.Context(), "mapped to signature invalid error",
			types.Field{Key: "error_code", Value: "SignatureInvalid"})
	default:
		span.SetTag("error.reason", "validation_failed")
		if tm.metrics {
			metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("validate", "validation_failed").Inc()
		}
		if strings.Contains(tokenErr.Error(), "token contains an invalid number of segments") {
			err = errors.NewTokenInvalidError("token format is invalid", tokenErr)
			tm.logger.Debug(span.Context(), "mapped to token invalid error (segments)",
				types.Field{Key: "error_code", Value: "TokenInvalid"})
		} else {
			err = errors.NewTokenInvalidError("token validation failed", tokenErr)
			tm.logger.Debug(span.Context(), "mapped to token invalid error (other)",
				types.Field{Key: "error_code", Value: "TokenInvalid"})
		}
	}

	// 添加最终错误的调试日志
	if customErr, ok := err.(errorTypes.Error); ok {
		tm.logger.Debug(span.Context(), "final error mapping",
			types.Field{Key: "error_code", Value: customErr.Code()},
			types.Field{Key: "error_msg", Value: customErr.Message()},
		)
	}

	return err
}
