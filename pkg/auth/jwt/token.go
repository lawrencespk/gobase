package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"gobase/pkg/errors"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metrics"
	"gobase/pkg/trace/jaeger"
)

// TokenManager JWT token管理器
type TokenManager struct {
	secretKey []byte
	logger    types.Logger
	provider  *jaeger.Provider
}

// NewTokenManager 创建新的token管理器
func NewTokenManager(secretKey string) (*TokenManager, error) {
	// 创建logger
	log, err := logger.NewLogger(
		logger.WithLevel(types.InfoLevel),
		logger.WithOutputPaths([]string{"stdout"}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create logger")
	}

	// 创建 jaeger provider
	provider, err := jaeger.NewProvider()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create jaeger provider")
	}

	return &TokenManager{
		secretKey: []byte(secretKey),
		logger:    log,
		provider:  provider,
	}, nil
}

// GenerateToken 生成令牌
func (tm *TokenManager) GenerateToken(ctx context.Context, claims Claims) (string, error) {
	start := time.Now()
	defer func() {
		metrics.DefaultJWTMetrics.TokenDuration.WithLabelValues("generate").Observe(time.Since(start).Seconds())
	}()

	// 验证claims
	if err := claims.Validate(); err != nil {
		metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("generate", err.Error()).Inc()
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(tm.secretKey)
	if err != nil {
		metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("generate", err.Error()).Inc()
		return "", errors.NewTokenGenerationError("failed to generate token", err)
	}

	return tokenString, nil
}

// ValidateToken 验证令牌
func (tm *TokenManager) ValidateToken(ctx context.Context, tokenString string) (*jwt.Token, error) {
	start := time.Now()
	defer func() {
		metrics.DefaultJWTMetrics.TokenDuration.WithLabelValues("validate").Observe(time.Since(start).Seconds())
	}()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.NewSignatureInvalidError("invalid signing method", nil)
		}
		return tm.secretKey, nil
	})

	if err != nil {
		metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("validate", err.Error()).Inc()
		return nil, tm.handleValidationError(err)
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

// handleValidationError 处理token验证错误
func (tm *TokenManager) handleValidationError(err error) error {
	// 使用 jaeger.NewSpan 创建新的 span
	span, err := jaeger.NewSpan("TokenManager.handleValidationError")
	if err != nil {
		return errors.Wrap(err, "failed to create span")
	}
	defer span.Finish()

	// 记录错误日志
	tm.logger.WithError(err).Error(span.Context(), "token validation failed")

	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		span.SetTag("error.reason", "token_expired")
		metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("validate", "expired").Inc()
		return errors.NewTokenExpiredError("token has expired", err)
	case errors.Is(err, jwt.ErrSignatureInvalid):
		span.SetTag("error.reason", "invalid_signature")
		metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("validate", "invalid_signature").Inc()
		return errors.NewSignatureInvalidError("invalid token signature", err)
	default:
		span.SetTag("error.reason", "validation_failed")
		metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("validate", "validation_failed").Inc()
		return errors.NewTokenInvalidError("token validation failed", err)
	}
}
