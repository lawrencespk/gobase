package security

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gobase/pkg/cache"
	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metrics"
	"gobase/pkg/trace/jaeger"
)

// Policy 安全策略配置
type Policy struct {
	// 密钥轮换配置
	EnableRotation   bool
	RotationInterval time.Duration

	// 绑定配置
	EnableIPBinding     bool
	EnableDeviceBinding bool

	// 会话配置
	EnableSession     bool
	MaxActiveSessions int

	// Token配置
	MaxTokenAge        time.Duration
	TokenReuseInterval time.Duration

	// 监控指标 (使用 JWTMetrics 替代 SecurityMetrics)
	metrics *metrics.JWTMetrics

	// 缓存支持
	cache  cache.Cache
	logger types.Logger
}

// NewPolicy 创建安全策略实例
func NewPolicy(opts ...PolicyOption) *Policy {
	p := &Policy{
		EnableRotation:      false,
		RotationInterval:    24 * time.Hour,
		EnableIPBinding:     false,
		EnableDeviceBinding: false,
		EnableSession:       false,
		MaxActiveSessions:   1,
		MaxTokenAge:         24 * time.Hour,
		TokenReuseInterval:  5 * time.Minute,
		metrics:             metrics.DefaultJWTMetrics,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// PolicyOption 策略配置选项
type PolicyOption func(*Policy)

// WithCache 设置缓存
func WithCache(cache cache.Cache) PolicyOption {
	return func(p *Policy) {
		p.cache = cache
	}
}

// WithLogger 设置日志记录器
func WithLogger(logger types.Logger) PolicyOption {
	return func(p *Policy) {
		p.logger = logger
	}
}

// WithMetrics 设置指标收集器
func WithMetrics(metrics *metrics.JWTMetrics) PolicyOption {
	return func(p *Policy) {
		p.metrics = metrics
	}
}

// WithRotation 配置密钥轮换
func (p *Policy) WithRotation(enabled bool, interval time.Duration) *Policy {
	p.EnableRotation = enabled
	p.RotationInterval = interval
	return p
}

// WithBinding 配置绑定
func (p *Policy) WithBinding(enableIP, enableDevice bool) *Policy {
	p.EnableIPBinding = enableIP
	p.EnableDeviceBinding = enableDevice
	return p
}

// WithSession 配置会话
func (p *Policy) WithSession(enabled bool, maxActive int) *Policy {
	p.EnableSession = enabled
	p.MaxActiveSessions = maxActive
	return p
}

// WithTokenAge 配置Token生命周期
func (p *Policy) WithTokenAge(ctx context.Context, maxAge, reuseInterval time.Duration) *Policy {
	// 添加追踪
	span, _ := jaeger.StartSpanFromContext(ctx, "Policy.WithTokenAge")
	defer span.Finish()

	span.SetTag("max_age", maxAge.String())
	span.SetTag("reuse_interval", reuseInterval.String())

	// 记录指标
	p.metrics.TokenAgeGauge.Set(float64(maxAge.Seconds()))
	p.metrics.TokenReuseIntervalGauge.Set(float64(reuseInterval.Seconds()))

	p.MaxTokenAge = maxAge
	p.TokenReuseInterval = reuseInterval
	return p
}

// ValidateTokenAge 验证令牌年龄
func (p *Policy) ValidateTokenAge(ctx context.Context, issuedAt time.Time) error {
	span, _ := jaeger.StartSpanFromContext(ctx, "security.policy.validate_age")
	defer span.Finish()

	age := time.Since(issuedAt)
	if age > p.MaxTokenAge {
		if p.metrics != nil {
			p.metrics.TokenErrors.WithLabels("validate", "age_exceeded").Inc()
		}
		return errors.NewTokenExpiredError("token age exceeds maximum allowed age", nil)
	}

	// 缓存最近验证时间
	if p.cache != nil {
		key := fmt.Sprintf("token:lastvalidated:%s", issuedAt.Format(time.RFC3339))
		if err := p.cache.Set(ctx, key, time.Now().Unix(), p.MaxTokenAge); err != nil {
			p.logger.Warn(ctx, "failed to cache token validation time",
				types.Field{Key: "issued_at", Value: issuedAt},
				types.Field{Key: "error", Value: err},
			)
		}
	}

	return nil
}

// ValidateTokenReuse 验证令牌重用
func (p *Policy) ValidateTokenReuse(ctx context.Context, tokenID string) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "security.policy.validate_reuse")
	defer span.Finish()

	if p.cache != nil {
		key := fmt.Sprintf("token:reuse:%s", tokenID)

		// 检查是否存在最近使用记录
		lastUsed, err := p.cache.Get(ctx, key)
		if err == nil {
			// 添加类型断言
			if data, ok := lastUsed.([]byte); ok {
				var lastUsedTime int64
				if err := json.Unmarshal(data, &lastUsedTime); err == nil {
					if time.Since(time.Unix(lastUsedTime, 0)) < p.TokenReuseInterval {
						if p.metrics != nil {
							p.metrics.TokenErrors.WithLabels("validate", "reuse_detected").Inc()
						}
						return errors.NewPolicyViolationError("token reuse detected", nil)
					}
				}
			}
		}

		// 更新最近使用时间
		data, _ := json.Marshal(time.Now().Unix())
		if err := p.cache.Set(ctx, key, data, p.TokenReuseInterval); err != nil {
			p.logger.Warn(ctx, "failed to cache token reuse time",
				types.Field{Key: "token_id", Value: tokenID},
				types.Field{Key: "error", Value: err},
			)
		}
	}

	return nil
}

// UpdatePolicy 更新策略配置
func (p *Policy) UpdatePolicy(ctx context.Context, maxAge, reuseInterval time.Duration) *Policy {
	span, ctx := jaeger.StartSpanFromContext(ctx, "security.policy.update")
	defer span.Finish()

	span.SetTag("max_age", maxAge.String())
	span.SetTag("reuse_interval", reuseInterval.String())

	// 记录指标
	if p.metrics != nil {
		p.metrics.TokenAgeGauge.Set(float64(maxAge.Seconds()))
		p.metrics.TokenReuseIntervalGauge.Set(float64(reuseInterval.Seconds()))
	}

	// 缓存策略配置
	if p.cache != nil {
		config := map[string]interface{}{
			"max_age":        maxAge,
			"reuse_interval": reuseInterval,
			"updated_at":     time.Now(),
		}

		data, _ := json.Marshal(config)
		if err := p.cache.Set(ctx, "security:policy:config", data, 24*time.Hour); err != nil {
			p.logger.Warn(ctx, "failed to cache security policy config",
				types.Field{Key: "error", Value: err},
			)
		}
	}

	p.MaxTokenAge = maxAge
	p.TokenReuseInterval = reuseInterval
	return p
}
