package blacklist

import (
	"context"
	"sync"
	"time"

	"gobase/pkg/monitor/prometheus/metric"
)

// MemoryBlacklist 基于内存的JWT黑名单实现
type MemoryBlacklist struct {
	tokens sync.Map
	opts   *Options

	// 监控指标
	tokenCount *metric.Gauge   // 当前黑名单中的token数量
	addTotal   *metric.Counter // 添加token的总次数
	hitTotal   *metric.Counter // 命中黑名单的总次数
	missTotal  *metric.Counter // 未命中黑名单的总次数
}

// NewMemoryBlacklist 创建内存黑名单实例
func NewMemoryBlacklist(opts *Options) (*MemoryBlacklist, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	bl := &MemoryBlacklist{
		opts: opts,
	}

	// 初始化监控指标
	bl.tokenCount = metric.NewGauge(metric.GaugeOpts{
		Namespace: "gobase",
		Subsystem: "jwt_blacklist",
		Name:      "tokens_total",
		Help:      "Total number of tokens in blacklist",
	})

	bl.addTotal = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "jwt_blacklist",
		Name:      "add_total",
		Help:      "Total number of tokens added to blacklist",
	})

	bl.hitTotal = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "jwt_blacklist",
		Name:      "hit_total",
		Help:      "Total number of blacklist hits",
	})

	bl.missTotal = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "jwt_blacklist",
		Name:      "miss_total",
		Help:      "Total number of blacklist misses",
	})

	// 注册指标收集器
	if err := metric.Register(bl); err != nil {
		return nil, err
	}

	return bl, nil
}

// Add 添加token到黑名单
func (bl *MemoryBlacklist) Add(ctx context.Context, token string, expiration time.Time) error {
	bl.tokens.Store(token, expiration)
	bl.tokenCount.Inc()
	bl.addTotal.Inc()
	return nil
}

// IsBlacklisted 检查token是否在黑名单中
func (bl *MemoryBlacklist) IsBlacklisted(ctx context.Context, token string) bool {
	if exp, ok := bl.tokens.Load(token); ok {
		expTime := exp.(time.Time)
		if time.Now().Before(expTime) {
			bl.hitTotal.Inc()
			return true
		}
		// token已过期,从黑名单中移除
		bl.tokens.Delete(token)
		bl.tokenCount.Dec()
	}
	bl.missTotal.Inc()
	return false
}

// Describe 实现 prometheus.Collector 接口
func (bl *MemoryBlacklist) Describe(ch chan<- *metric.Desc) {
	bl.tokenCount.Describe(ch)
	bl.addTotal.Describe(ch)
	bl.hitTotal.Describe(ch)
	bl.missTotal.Describe(ch)
}

// Collect 实现 prometheus.Collector 接口
func (bl *MemoryBlacklist) Collect(ch chan<- metric.Metric) {
	bl.tokenCount.Collect(ch)
	bl.addTotal.Collect(ch)
	bl.hitTotal.Collect(ch)
	bl.missTotal.Collect(ch)
}
