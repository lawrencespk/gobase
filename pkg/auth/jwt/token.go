package jwt

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
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

	// 验证必要字段
	if claims.GetUserID() == "" {
		return "", errors.NewError(codes.TokenGenerationError, "user_id is required", nil)
	}

	if claims.GetTokenType() == "" {
		return "", errors.NewError(codes.TokenGenerationError, "token_type is required", nil)
	}

	// 验证过期时间
	if exp := claims.GetExpiresAt(); !exp.IsZero() && exp.Before(time.Now()) {
		return "", errors.NewError(codes.TokenExpired, "token is expired", nil)
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	tokenString, err := token.SignedString(tm.secretKey)
	if err != nil {
		if tm.metrics {
			metrics.DefaultJWTMetrics.TokenErrors.WithLabelValues("generate", err.Error()).Inc()
		}
		return "", errors.NewError(codes.TokenSignFailed, "failed to sign token", err)
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

// Token JWT令牌结构
type Token struct {
	token      *jwt.Token
	claims     Claims
	secret     string
	expiration time.Duration
	validate   bool
}

// NewToken 创建新的JWT token
func NewToken(opts ...TokenOption) (*Token, error) {
	t := &Token{}
	for _, opt := range opts {
		opt(t)
	}

	if t.claims == nil {
		return nil, errors.NewError(codes.TokenGenerationError, "claims is required", nil)
	}

	// 创建JWT token
	t.token = jwt.NewWithClaims(jwt.SigningMethodHS256, t.claims)

	return t, nil
}

// ParseToken 解析JWT token
func ParseToken(tokenString string, opts ...TokenOption) (*Token, error) {
	t := &Token{}
	for _, opt := range opts {
		opt(t)
	}

	if t.secret == "" {
		return nil, errors.NewError(codes.TokenInvalid, "secret is required", nil)
	}

	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.secret), nil
	})

	if err != nil {
		return nil, errors.NewError(codes.TokenInvalid, "failed to parse token", err)
	}

	t.token = token
	t.claims = token.Claims.(*StandardClaims)

	return t, nil
}

// SignedString 获取签名后的token字符串
func (t *Token) SignedString() (string, error) {
	if t.secret == "" {
		return "", errors.NewError(codes.TokenSignFailed, "secret is required", nil)
	}
	return t.token.SignedString([]byte(t.secret))
}

// Claims 获取token的claims
func (t *Token) Claims() Claims {
	return t.claims
}

// TokenOption token配置选项
type TokenOption func(*Token)

// WithClaims 设置claims
func WithClaims(claims Claims) TokenOption {
	return func(t *Token) {
		t.claims = claims
	}
}

// WithSecret 设置密钥
func WithSecret(secret string) TokenOption {
	return func(t *Token) {
		t.secret = secret
	}
}

// WithExpiration 设置过期时间
func WithExpiration(exp time.Duration) TokenOption {
	return func(t *Token) {
		t.expiration = exp
	}
}

// 添加错误检查函数
func IsTokenExpiredError(err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return true
	default:
		return false
	}
}

func IsSignatureInvalidError(err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, jwt.ErrSignatureInvalid):
		return true
	default:
		return false
	}
}

// 添加 WithValidation 选项
func WithValidation(validate bool) TokenOption {
	return func(t *Token) {
		t.validate = validate
	}
}

// 添加带上下文的解析函数
func ParseTokenWithContext(ctx context.Context, tokenString string, opts ...TokenOption) (*Token, error) {
	t := &Token{}
	for _, opt := range opts {
		opt(t)
	}

	if t.secret == "" {
		return nil, errors.NewError(codes.TokenInvalid, "secret is required", nil)
	}

	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.secret), nil
	})

	if err != nil {
		return nil, errors.NewError(codes.TokenInvalid, "failed to parse token", err)
	}

	t.token = token
	t.claims = token.Claims.(*StandardClaims)

	return t, nil
}
