package multilevel

import (
	"context"
	"sync"
	"time"

	"gobase/pkg/cache"
	"gobase/pkg/cache/memory"
	"gobase/pkg/cache/redis"
	redisClient "gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metric"
	"gobase/pkg/trace/jaeger"
)

// Manager 多级缓存管理器
type Manager struct {
	// 缓存层级映射
	caches map[cache.Level]cache.Cache

	// 缓存配置
	config *Config

	// 互斥锁
	mu sync.RWMutex

	// 日志记录器
	logger types.Logger

	// 监控指标
	metrics *metric.Counter

	// 添加 redisClient 字段
	redisClient redisClient.Client
}

// NewManager 创建多级缓存管理器
func NewManager(config *Config, redisClient redisClient.Client, logger types.Logger) (*Manager, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	m := &Manager{
		config:      config,
		logger:      logger,
		redisClient: redisClient,
		caches:      make(map[cache.Level]cache.Cache),
		metrics: metric.NewCounter(metric.CounterOpts{
			Namespace: "gobase",
			Subsystem: "cache",
			Name:      "multilevel_operations_total",
			Help:      "Total number of multilevel cache operations",
		}).WithLabels("operation", "level", "status"),
	}

	if err := m.initCaches(); err != nil {
		return nil, err
	}

	return m, nil
}

// Get 获取缓存,按照L1->L2的顺序查找
func (m *Manager) Get(ctx context.Context, key string) (interface{}, error) {
	span, ctx := jaeger.StartSpanFromContext(ctx, "cache.multilevel.get")
	if span != nil {
		defer span.Finish()
	}

	// 先从L1缓存获取
	value, err := m.GetFromLevel(ctx, key, cache.L1Cache)
	if err == nil {
		m.metrics.WithLabels("get", "l1", "hit").Inc()
		return value, nil
	}

	// L1未命中,从L2缓存获取
	value, err = m.GetFromLevel(ctx, key, cache.L2Cache)
	if err != nil {
		m.metrics.WithLabels("get", "l2", "miss").Inc()
		// 确保返回 RedisKeyNotFoundError
		if errors.HasErrorCode(err, codes.CacheMissError) ||
			errors.HasErrorCode(err, codes.CacheExpiredError) {
			return nil, errors.NewRedisKeyNotFoundError("cache not found", err)
		}
		return nil, err
	}

	// L2命中,回写L1缓存
	if err := m.SetToLevel(ctx, key, value, m.config.L1TTL, cache.L1Cache); err != nil {
		m.logger.Warn(ctx, "failed to write back to L1 cache",
			types.Field{Key: "error", Value: err})
	}

	m.metrics.WithLabels("get", "l2", "hit").Inc()
	return value, nil
}

// Set 设置缓存,同时写入L1和L2
func (m *Manager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "cache.multilevel.set")
	if span != nil {
		defer span.Finish()
	}

	// 使用对象池
	k := keyPool.Get().(*string)
	*k = key
	defer keyPool.Put(k)

	v := valuePool.Get().(*interface{})
	*v = value
	defer valuePool.Put(v)

	// 并发写入 L1 和 L2
	var wg sync.WaitGroup
	var l1Err, l2Err error

	wg.Add(2)
	go func() {
		defer wg.Done()
		l1Err = m.SetToLevel(ctx, *k, *v, m.config.L1TTL, cache.L1Cache)
	}()

	go func() {
		defer wg.Done()
		l2Err = m.SetToLevel(ctx, *k, *v, expiration, cache.L2Cache)
	}()

	wg.Wait()

	// 处理错误
	if l1Err != nil {
		m.logger.Warn(ctx, "failed to set L1 cache",
			types.Field{Key: "error", Value: l1Err})
	}
	if l2Err != nil {
		return l2Err
	}
	return nil
}

// Delete 删除缓存,同时删除L1和L2
func (m *Manager) Delete(ctx context.Context, key string) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "cache.multilevel.delete")
	if span != nil {
		defer span.Finish()
	}

	// 删除L1缓存
	if err := m.DeleteFromLevel(ctx, key, cache.L1Cache); err != nil {
		m.logger.Warn(ctx, "failed to delete L1 cache",
			types.Field{Key: "error", Value: err})
	}

	// 删除L2缓存
	if err := m.DeleteFromLevel(ctx, key, cache.L2Cache); err != nil {
		m.metrics.WithLabels("delete", "l2", "error").Inc()
		return err
	}

	m.metrics.WithLabels("delete", "all", "success").Inc()
	return nil
}

// Warmup 缓存预热
func (m *Manager) Warmup(ctx context.Context, keys []string) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "cache.multilevel.warmup")
	if span != nil {
		defer span.Finish()
	}

	// 并发预热
	var wg sync.WaitGroup
	errChan := make(chan error, len(keys))

	for _, key := range keys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()

			// 从L2获取数据
			value, err := m.GetFromLevel(ctx, key, cache.L2Cache)
			if err != nil {
				errChan <- err
				return
			}

			// 写入L1缓存
			if err := m.SetToLevel(ctx, key, value, m.config.L1TTL, cache.L1Cache); err != nil {
				m.logger.Warn(ctx, "failed to warmup L1 cache",
					types.Field{Key: "key", Value: key},
					types.Field{Key: "error", Value: err})
			}
		}(key)
	}

	// 等待所有预热任务完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		if err != nil {
			m.metrics.WithLabels("warmup", "all", "error").Inc()
			return err
		}
	}

	m.metrics.WithLabels("warmup", "all", "success").Inc()
	return nil
}

// initCaches 初始化缓存层级
func (m *Manager) initCaches() error {
	// 转换L1配置
	memoryConfig := &memory.Config{
		MaxEntries:      m.config.L1Config.MaxEntries,
		CleanupInterval: m.config.L1Config.CleanupInterval,
		DefaultTTL:      m.config.L1TTL,
	}

	// 初始化L1缓存(内存缓存)
	l1Cache, err := memory.NewCache(memoryConfig, m.logger)
	if err != nil {
		return errors.NewInitializationError("failed to init L1 cache", err)
	}
	m.caches[cache.L1Cache] = l1Cache

	// 初始化L2缓存(Redis缓存)
	redisConfig := redis.Options{
		Client: m.redisClient,
		Logger: m.logger,
	}
	l2Cache, err := redis.NewCache(redisConfig)
	if err != nil {
		return errors.NewInitializationError("failed to init L2 cache", err)
	}
	m.caches[cache.L2Cache] = l2Cache

	return nil
}

// GetFromLevel 从指定级别获取缓存
func (m *Manager) GetFromLevel(ctx context.Context, key string, level cache.Level) (interface{}, error) {
	span, ctx := jaeger.StartSpanFromContext(ctx, "cache.multilevel.GetFromLevel")
	if span != nil {
		defer span.Finish()
	}

	m.mu.RLock()
	cache, ok := m.caches[level]
	m.mu.RUnlock()
	if !ok {
		return nil, errors.NewCacheNotFoundError("cache level not found", nil)
	}

	value, err := cache.Get(ctx, key)
	if err != nil {
		m.metrics.WithLabels("get", m.getLevelString(level), "error").Inc()
		return nil, err
	}

	m.metrics.WithLabels("get", m.getLevelString(level), "success").Inc()
	return value, nil
}

// SetToLevel 设置缓存到指定级别
func (m *Manager) SetToLevel(ctx context.Context, key string, value interface{}, expiration time.Duration, level cache.Level) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "cache.multilevel.SetToLevel")
	if span != nil {
		defer span.Finish()
	}

	m.mu.RLock()
	cache, ok := m.caches[level]
	m.mu.RUnlock()
	if !ok {
		return errors.NewCacheNotFoundError("cache level not found", nil)
	}

	if err := cache.Set(ctx, key, value, expiration); err != nil {
		m.metrics.WithLabels("set", m.getLevelString(level), "error").Inc()
		return err
	}

	m.metrics.WithLabels("set", m.getLevelString(level), "success").Inc()
	return nil
}

// DeleteFromLevel 从指定级别删除缓存
func (m *Manager) DeleteFromLevel(ctx context.Context, key string, level cache.Level) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "cache.multilevel.DeleteFromLevel")
	if span != nil {
		defer span.Finish()
	}

	m.mu.RLock()
	cache, ok := m.caches[level]
	m.mu.RUnlock()
	if !ok {
		return errors.NewCacheNotFoundError("cache level not found", nil)
	}

	if err := cache.Delete(ctx, key); err != nil {
		m.metrics.WithLabels("delete", m.getLevelString(level), "error").Inc()
		return err
	}

	m.metrics.WithLabels("delete", m.getLevelString(level), "success").Inc()
	return nil
}

// ClearLevel 清空指定级别的缓存
func (m *Manager) ClearLevel(ctx context.Context, level cache.Level) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "cache.multilevel.ClearLevel")
	if span != nil {
		defer span.Finish()
	}

	m.mu.RLock()
	cache, ok := m.caches[level]
	m.mu.RUnlock()
	if !ok {
		return errors.NewCacheNotFoundError("cache level not found", nil)
	}

	if err := cache.Clear(ctx); err != nil {
		m.metrics.WithLabels("clear", m.getLevelString(level), "error").Inc()
		return err
	}

	m.metrics.WithLabels("clear", m.getLevelString(level), "success").Inc()
	return nil
}

// getLevelString 获取缓存级别的字符串表示
func (m *Manager) getLevelString(level cache.Level) string {
	switch level {
	case cache.L1Cache:
		return "l1"
	case cache.L2Cache:
		return "l2"
	default:
		return "unknown"
	}
}

// 添加对象池以减少内存分配
var (
	keyPool = sync.Pool{
		New: func() interface{} {
			return new(string)
		},
	}
	valuePool = sync.Pool{
		New: func() interface{} {
			return new(interface{})
		},
	}
)
