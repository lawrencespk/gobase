package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"gobase/pkg/errors"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
)

var (
	// TokenDuration 令牌操作耗时指标
	TokenDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "jwt_token_duration_seconds",
			Help: "Duration of JWT token operations in seconds",
		},
		[]string{"operation"},
	)

	// TokenErrors 令牌错误计数
	TokenErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jwt_token_errors_total",
			Help: "Total number of JWT token errors",
		},
		[]string{"operation", "error"},
	)

	// TokenGenerateCounter 令牌生成计数
	TokenGenerateCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jwt_token_generate_total",
			Help: "Total number of JWT token generations",
		},
		[]string{"status"},
	)

	// TokenValidateCounter 令牌验证计数
	TokenValidateCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jwt_token_validate_total",
			Help: "Total number of JWT token validations",
		},
		[]string{"status"},
	)
)

func init() {
	// 注册指标
	prometheus.MustRegister(TokenDuration)
	prometheus.MustRegister(TokenErrors)
	prometheus.MustRegister(TokenGenerateCounter)
	prometheus.MustRegister(TokenValidateCounter)
}

// TokenManager JWT token管理器
type TokenManager struct {
	secretKey []byte
	logger    types.Logger
	tracer    opentracing.Tracer
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

	return &TokenManager{
		secretKey: []byte(secretKey),
		logger:    log,
		tracer:    opentracing.GlobalTracer(),
	}, nil
}

// GenerateToken 生成令牌
func (tm *TokenManager) GenerateToken(ctx context.Context, claims Claims) (string, error) {
	start := time.Now()
	defer func() {
		TokenDuration.WithLabelValues("generate").Observe(time.Since(start).Seconds())
	}()

	// 验证claims
	if err := claims.Validate(); err != nil {
		TokenErrors.WithLabelValues("generate", err.Error()).Inc()
		TokenGenerateCounter.WithLabelValues("error").Inc()
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(tm.secretKey)
	if err != nil {
		TokenErrors.WithLabelValues("generate", err.Error()).Inc()
		TokenGenerateCounter.WithLabelValues("error").Inc()
		return "", errors.NewTokenGenerationError("failed to generate token", err)
	}

	TokenGenerateCounter.WithLabelValues("success").Inc()
	return tokenString, nil
}

// ValidateToken 验证令牌
func (tm *TokenManager) ValidateToken(ctx context.Context, tokenString string) (*jwt.Token, error) {
	start := time.Now()
	defer func() {
		TokenDuration.WithLabelValues("validate").Observe(time.Since(start).Seconds())
	}()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.NewSignatureInvalidError("invalid signing method", nil)
		}
		return tm.secretKey, nil
	})

	if err != nil {
		TokenErrors.WithLabelValues("validate", err.Error()).Inc()
		TokenValidateCounter.WithLabelValues("error").Inc()
		return nil, tm.handleValidationError(err)
	}

	TokenValidateCounter.WithLabelValues("success").Inc()
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
	// 添加追踪
	span, ctx := opentracing.StartSpanFromContext(context.Background(), "TokenManager.handleValidationError")
	defer span.Finish()

	// 添加错误标签
	span.SetTag("error", true)
	span.SetTag("error.type", "validation_error")

	// 记录日志
	tm.logger.WithContext(ctx).WithFields(
		types.Field{Key: "error", Value: err},
	).Error(ctx, "token validation failed")

	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		span.SetTag("error.reason", "token_expired")
		return errors.NewTokenExpiredError("token has expired", err)
	case errors.Is(err, jwt.ErrSignatureInvalid):
		span.SetTag("error.reason", "invalid_signature")
		return errors.NewSignatureInvalidError("invalid token signature", err)
	default:
		span.SetTag("error.reason", "validation_failed")
		return errors.NewTokenInvalidError("token validation failed", err)
	}
}
