package memory

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"gobase/pkg/cache"
	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metric"
	"hash/fnv"
)

// Cache 内存缓存实现
type Cache struct {
	// 使用分片来减少锁竞争
	shards    []*cacheShard
	numShards int

	// 配置信息
	config *Config

	// 日志记录器
	logger types.Logger

	// 监控指标
	metrics *metric.Counter

	// 停止信号
	stopCh chan struct{}

	// 有效缓存项数量
	count int64
}

// cacheShard 缓存分片
type cacheShard struct {
	data sync.Map
}

var itemPool = sync.Pool{
	New: func() interface{} {
		return &cacheItem{}
	},
}

// NewCache 创建内存缓存
func NewCache(config *Config, logger types.Logger) (*Cache, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// 创建256个分片
	numShards := 256
	shards := make([]*cacheShard, numShards)
	for i := 0; i < numShards; i++ {
		shards[i] = &cacheShard{}
	}

	c := &Cache{
		shards:    shards,
		numShards: numShards,
		config:    config,
		logger:    logger,
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

// getShard 获取key对应的分片
func (c *Cache) getShard(key string) *cacheShard {
	// 使用 fnv hash 算法来确定分片
	h := fnv.New64()
	h.Write([]byte(key))
	return c.shards[h.Sum64()%uint64(c.numShards)]
}

// Get 获取缓存数据
func (c *Cache) Get(ctx context.Context, key string) (interface{}, error) {
	shard := c.getShard(key)
	if value, ok := shard.data.Load(key); ok {
		item := value.(*cacheItem)
		if !item.isExpired() {
			c.metrics.WithLabels("get", "hit").Inc()
			return item.value, nil
		}
		// 使用 LoadAndDelete 替代 Delete,确保原子性
		if _, loaded := shard.data.LoadAndDelete(key); loaded {
			atomic.AddInt64(&c.count, -1)
		}
		c.metrics.WithLabels("get", "expired").Inc()
		return nil, errors.NewCacheExpiredError("cache expired", nil)
	}
	c.metrics.WithLabels("get", "miss").Inc()
	return nil, errors.NewCacheNotFoundError("cache miss", nil)
}

// Set 设置缓存数据
func (c *Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	shard := c.getShard(key)
	item := itemPool.Get().(*cacheItem)
	item.value = value
	item.expiration = time.Now().Add(ttl)

	// 使用 LoadOrStore 保证原子性
	for {
		// 检查容量
		if atomic.LoadInt64(&c.count) >= int64(c.config.MaxEntries) {
			c.metrics.WithLabels("set", "full").Inc()
			c.cleanup()
			// 如果清理后仍然满了，返回错误
			if atomic.LoadInt64(&c.count) >= int64(c.config.MaxEntries) {
				return errors.NewCacheCapacityError("memory cache max entries exceeded", nil)
			}
		}

		oldValue, loaded := shard.data.LoadOrStore(key, item)
		if !loaded {
			// 新增项
			atomic.AddInt64(&c.count, 1)
			c.metrics.WithLabels("set", "success").Inc()
			return nil
		}

		// 检查已存在项
		old := oldValue.(*cacheItem)
		if !old.isExpired() {
			// 未过期，直接更新
			shard.data.Store(key, item)
			c.metrics.WithLabels("set", "update").Inc()
			return nil
		}

		// 已过期，删除后重试
		if _, ok := shard.data.LoadAndDelete(key); ok {
			atomic.AddInt64(&c.count, -1)
			continue
		}
	}
}

// GetLevel 获取缓存级别
func (c *Cache) GetLevel() cache.Level {
	return cache.L1Cache
}

// Delete 删除缓存数据
func (c *Cache) Delete(ctx context.Context, key string) error {
	shard := c.getShard(key)
	if _, ok := shard.data.LoadAndDelete(key); ok {
		atomic.AddInt64(&c.count, -1)
	}
	c.metrics.WithLabels("delete", "success").Inc()
	return nil
}

// Clear 清空缓存
func (c *Cache) Clear(ctx context.Context) error {
	for _, shard := range c.shards {
		shard.data.Range(func(key, _ interface{}) bool {
			shard.data.Delete(key)
			atomic.AddInt64(&c.count, -1)
			return true
		})
	}
	c.metrics.WithLabels("clear", "success").Inc()
	return nil
}

// Stop 停止缓存清理
func (c *Cache) Stop() {
	close(c.stopCh)
}

// SetCleanupInterval 设置清理间隔
func (c *Cache) SetCleanupInterval(d time.Duration) {
	c.config.CleanupInterval = d
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

	memoryTicker := time.NewTicker(time.Second) // 每秒检查内存使用
	defer memoryTicker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-memoryTicker.C:
			// 如果内存使用超过80%，触发清理
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			if m.Alloc > m.TotalAlloc*8/10 {
				c.cleanup()
			}
		case <-c.stopCh:
			return
		}
	}
}

// cleanup 清理过期数据
func (c *Cache) cleanup() {
	var totalRemoved int64

	// 并发清理各个分片
	var wg sync.WaitGroup
	for i := 0; i < c.numShards; i++ {
		wg.Add(1)
		go func(shard *cacheShard) {
			defer wg.Done()
			removed := c.cleanupShard(shard)
			if removed > 0 {
				atomic.AddInt64(&totalRemoved, removed)
			}
		}(c.shards[i])
	}
	wg.Wait()

	if totalRemoved > 0 {
		c.logger.Debug(context.Background(), "cleaned up expired cache items",
			types.Field{Key: "count", Value: totalRemoved})
	}
}

// cleanupShard 清理单个分片的过期数据
func (c *Cache) cleanupShard(shard *cacheShard) int64 {
	var removed int64
	keysToDelete := make([]interface{}, 0)

	shard.data.Range(func(key, value interface{}) bool {
		item := value.(*cacheItem)
		if item.isExpired() {
			keysToDelete = append(keysToDelete, key)
			removed++
		}
		return true
	})

	if removed > 0 {
		for _, key := range keysToDelete {
			// 使用 LoadAndDelete 确保不重复删除
			if _, ok := shard.data.LoadAndDelete(key); ok {
				atomic.AddInt64(&c.count, -1)
			}
		}
	}

	return removed
}
