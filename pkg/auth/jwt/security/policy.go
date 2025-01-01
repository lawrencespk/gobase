package security

import (
	"context"
	"time"

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

	// 监控指标
	metrics *metrics.JWTMetrics
}

// NewPolicy 创建安全策略
func NewPolicy() *Policy {
	return &Policy{
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
	span, ctx := jaeger.StartSpanFromContext(ctx, "Policy.WithTokenAge")
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
