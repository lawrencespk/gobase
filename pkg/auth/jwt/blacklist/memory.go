package blacklist

import (
	"context"
	"sync"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
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

// MemoryStore 基于内存的存储实现
type MemoryStore struct {
	*MemoryBlacklist
}

// NewMemoryStore 创建内存存储实例
func NewMemoryStore() Store {
	bl, err := NewMemoryBlacklist(DefaultOptions())
	if err != nil {
		// 由于这是内部初始化，使用默认选项不应该出错
		// 如果出错，说明是严重的系统问题，应该panic
		panic(err)
	}
	store := &MemoryStore{
		MemoryBlacklist: bl,
	}

	// 启动定期清理任务
	go func() {
		ticker := time.NewTicker(bl.opts.CleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			store.Cleanup()
		}
	}()

	return store
}

// BlacklistItem 存储黑名单项的详细信息
type BlacklistItem struct {
	Reason     string
	ExpireTime time.Time
}

// Add 添加token到黑名单
func (s *MemoryStore) Add(ctx context.Context, tokenID, reason string, expiration time.Duration) error {
	if tokenID == "" {
		return errors.NewError(codes.InvalidParams, "token ID is required", nil)
	}
	if expiration <= 0 {
		return errors.NewError(codes.InvalidParams, "expiration must be positive", nil)
	}
	s.MemoryBlacklist.tokens.Store(tokenID, BlacklistItem{
		Reason:     reason,
		ExpireTime: time.Now().Add(expiration),
	})
	s.MemoryBlacklist.tokenCount.Inc()
	s.MemoryBlacklist.addTotal.Inc()
	return nil
}

// Get 获取黑名单原因
func (s *MemoryStore) Get(ctx context.Context, tokenID string) (string, error) {
	if tokenID == "" {
		return "", errors.NewError(codes.InvalidParams, "token ID is required", nil)
	}

	if value, ok := s.MemoryBlacklist.tokens.Load(tokenID); ok {
		item := value.(BlacklistItem)
		if time.Now().Before(item.ExpireTime) {
			return item.Reason, nil // 返回用户定义的原因
		}
		// token已过期,从黑名单中移除
		s.MemoryBlacklist.tokens.Delete(tokenID)
		s.MemoryBlacklist.tokenCount.Dec()
	}
	return "", errors.NewError(codes.StoreErrNotFound, "token not found in blacklist", nil)
}

// Remove 从黑名单中移除
func (s *MemoryStore) Remove(ctx context.Context, tokenID string) error {
	if tokenID == "" {
		return errors.NewError(codes.InvalidParams, "token ID is required", nil)
	}
	s.MemoryBlacklist.tokens.Delete(tokenID)
	s.MemoryBlacklist.tokenCount.Dec()
	return nil
}

// Close 关闭存储
func (s *MemoryStore) Close() error {
	return nil
}

// Cleanup 清理过期的token
func (s *MemoryStore) Cleanup() {
	now := time.Now()
	s.tokens.Range(func(key, value interface{}) bool {
		item := value.(BlacklistItem)
		if now.After(item.ExpireTime) {
			s.tokens.Delete(key)
			s.tokenCount.Dec()
		}
		return true
	})
}
