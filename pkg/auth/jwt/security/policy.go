package security

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gobase/pkg/cache"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
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
	Metrics *metrics.JWTMetrics

	// 缓存支持
	Cache  cache.Cache
	Logger types.Logger
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
		Metrics:             metrics.DefaultJWTMetrics,
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
		p.Cache = cache
	}
}

// WithLogger 设置日志记录器
func WithLogger(logger types.Logger) PolicyOption {
	return func(p *Policy) {
		p.Logger = logger
	}
}

// WithMetrics 设置指标收集器
func WithMetrics(metrics *metrics.JWTMetrics) PolicyOption {
	return func(p *Policy) {
		p.Metrics = metrics
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
	p.Metrics.TokenAgeGauge.Set(float64(maxAge.Seconds()))
	p.Metrics.TokenReuseIntervalGauge.Set(float64(reuseInterval.Seconds()))

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
		if p.Metrics != nil {
			p.Metrics.TokenErrors.WithLabels("validate", "age_exceeded").Inc()
		}
		return errors.NewTokenExpiredError("token age exceeds maximum allowed age", nil)
	}

	// 缓存最近验证时间
	if p.Cache != nil {
		key := fmt.Sprintf("token:lastvalidated:%s", issuedAt.Format(time.RFC3339))
		if err := p.Cache.Set(ctx, key, time.Now().Unix(), p.MaxTokenAge); err != nil {
			p.Logger.Warn(ctx, "failed to cache token validation time",
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

	if p.Cache != nil {
		key := fmt.Sprintf("token:reuse:%s", tokenID)

		// 1. 先尝试获取
		value, err := p.Cache.Get(ctx, key)
		if err != nil {
			// 如果是键不存在的错误，我们可以继续处理
			if errors.Is(err, errors.NewError(codes.NotFound, "", nil)) {
				// 继续执行设置操作
			} else {
				p.Logger.Warn(ctx, "failed to check token reuse",
					types.Field{Key: "token_id", Value: tokenID},
					types.Field{Key: "error", Value: err},
				)
				if p.Metrics != nil {
					p.Metrics.TokenErrors.WithLabels("validate", "reuse_check_failed").Inc()
				}
				return errors.NewPolicyViolationError("token reuse check failed", err)
			}
		}

		// 2. 如果找到了值，说明令牌在重用间隔内被使用过
		if value != nil {
			if p.Metrics != nil {
				p.Metrics.TokenErrors.WithLabels("validate", "reuse_detected").Inc()
			}
			return errors.NewPolicyViolationError("token reuse detected", nil)
		}

		// 3. 如果没有找到值，设置新的使用记录
		currentTime := time.Now().Unix()
		err = p.Cache.Set(ctx, key, currentTime, p.TokenReuseInterval)
		if err != nil {
			p.Logger.Warn(ctx, "failed to set token reuse record",
				types.Field{Key: "token_id", Value: tokenID},
				types.Field{Key: "error", Value: err},
			)
			if p.Metrics != nil {
				p.Metrics.TokenErrors.WithLabels("validate", "record_failed").Inc()
			}
			return errors.NewPolicyViolationError("failed to record token usage", err)
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
	if p.Metrics != nil {
		p.Metrics.TokenAgeGauge.Set(float64(maxAge.Seconds()))
		p.Metrics.TokenReuseIntervalGauge.Set(float64(reuseInterval.Seconds()))
	}

	// 缓存策略配置
	if p.Cache != nil {
		config := map[string]interface{}{
			"max_age":        maxAge,
			"reuse_interval": reuseInterval,
			"updated_at":     time.Now(),
		}

		data, _ := json.Marshal(config)
		if err := p.Cache.Set(ctx, "security:policy:config", data, 24*time.Hour); err != nil {
			p.Logger.Warn(ctx, "failed to cache security policy config",
				types.Field{Key: "error", Value: err},
			)
		}
	}

	p.MaxTokenAge = maxAge
	p.TokenReuseInterval = reuseInterval
	return p
}

// UpdateConfig 更新配置方法
func (p *Policy) UpdateConfig(maxAge, reuseInterval time.Duration) *Policy {
	p.MaxTokenAge = maxAge
	p.TokenReuseInterval = reuseInterval
	return p
}
