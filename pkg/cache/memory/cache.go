package memory

import (
	"context"
	"sync"
	"time"

	"gobase/pkg/cache"
	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metric"
)

// Cache 内存缓存实现
type Cache struct {
	// 缓存数据
	data sync.Map

	// 配置信息
	config *Config

	// 日志记录器
	logger types.Logger

	// 监控指标
	metrics *metric.Counter

	// 停止信号
	stopCh chan struct{}
}

// NewCache 创建内存缓存
func NewCache(config *Config, logger types.Logger) (*Cache, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	c := &Cache{
		config: config,
		logger: logger,
		metrics: metric.NewCounter(metric.CounterOpts{
			Namespace: "gobase",
			Subsystem: "cache",
			Name:      "memory_operations_total",
			Help:      "Total number of memory cache operations",
		}).WithLabels("operation", "status"),
		stopCh: make(chan struct{}),
	}

	// 启动清理协程
	go c.cleanupLoop()

	return c, nil
}

// Get 获取缓存
func (c *Cache) Get(ctx context.Context, key string) (interface{}, error) {
	value, ok := c.data.Load(key)
	if !ok {
		c.metrics.WithLabels("get", "miss").Inc()
		return nil, errors.NewRedisKeyNotFoundError("cache miss", nil)
	}

	item := value.(*cacheItem)
	if item.isExpired() {
		c.data.Delete(key)
		c.metrics.WithLabels("get", "expired").Inc()
		return nil, errors.NewRedisKeyNotFoundError("cache expired", nil)
	}

	c.metrics.WithLabels("get", "hit").Inc()
	return item.value, nil
}

// Set 设置缓存
func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	item := &cacheItem{
		value:      value,
		expiration: time.Now().Add(expiration),
	}

	c.data.Store(key, item)
	c.metrics.WithLabels("set", "success").Inc()
	return nil
}

// GetLevel 获取缓存级别
func (c *Cache) GetLevel() cache.Level {
	return cache.L1Cache
}

// Delete 删除缓存
func (c *Cache) Delete(ctx context.Context, key string) error {
	c.data.Delete(key)
	c.metrics.WithLabels("delete", "success").Inc()
	return nil
}

// Clear 清空缓存
func (c *Cache) Clear(ctx context.Context) error {
	c.data = sync.Map{}
	c.metrics.WithLabels("clear", "success").Inc()
	return nil
}

type cacheItem struct {
	value      interface{}
	expiration time.Time
}

func (i *cacheItem) isExpired() bool {
	return time.Now().After(i.expiration)
}

func (c *Cache) cleanupLoop() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopCh:
			return
		}
	}
}

func (c *Cache) cleanup() {
	var count int
	c.data.Range(func(key, value interface{}) bool {
		item := value.(*cacheItem)
		if item.isExpired() {
			c.data.Delete(key)
			count++
		}
		return true
	})

	if count > 0 {
		c.logger.Debug(context.Background(), "cleaned up expired cache items",
			types.Field{Key: "count", Value: count})
	}
}
