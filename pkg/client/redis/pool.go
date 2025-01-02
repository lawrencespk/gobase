package redis

import (
	"context"
	"strings"
	"sync"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/collector"

	"github.com/go-redis/redis/v8"
)

// Pool Redis连接池接口
type Pool interface {
	// Stats 获取连接池统计信息
	Stats() *PoolStats
	// Close 关闭连接池
	Close() error
}

// PoolStats 连接池统计信息
type PoolStats struct {
	// Hits 命中次数
	Hits uint32
	// Misses 未命中次数
	Misses uint32
	// Timeouts 超时次数
	Timeouts uint32
	// TotalConns 总连接数
	TotalConns uint32
	// IdleConns 空闲连接数
	IdleConns uint32
	// ActiveCount 活跃连接数
	ActiveCount int64
	// IdleCount 空闲连接数
	IdleCount int64
	// TotalCount 总连接数
	TotalCount int64
	// WaitCount 等待连接数
	WaitCount int64
	// TimeoutCount 超时连接数
	TimeoutCount int64
	// HitCount 命中次数
	HitCount int64
	// MissCount 未命中次数
	MissCount int64
}

// pool Redis连接池实现
type pool struct {
	client  redis.UniversalClient // 改为 UniversalClient 类型
	logger  types.Logger
	metrics *collector.RedisCollector
	options *Options
}

// NewPool 创建连接池
func NewPool(client redis.UniversalClient, logger types.Logger, options *Options, metrics *collector.RedisCollector) Pool {
	if client == nil {
		logger.Error(context.Background(), "redis client is nil")
		return nil
	}

	p := &pool{
		client:  client,
		logger:  logger,
		options: options,
		metrics: metrics,
	}

	// 添加连接池预热
	p.warmup()

	// 如果启用了指标监控，启动指标记录
	if options != nil && options.EnableMetrics {
		p.recordMetrics(context.Background())
	}

	return p
}

// warmup 预热连接池
func (p *pool) warmup() {
	if p.options == nil {
		return
	}

	// 使用完整的 PoolSize 进行预热
	targetConns := p.options.PoolSize

	var wg sync.WaitGroup
	for i := 0; i < targetConns; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			_ = p.client.Ping(ctx)
		}()
	}
	wg.Wait()

	// 等待连接池稳定
	time.Sleep(100 * time.Millisecond)
}

// Stats 获取连接池统计信息
func (p *pool) Stats() *PoolStats {
	stats := p.client.PoolStats()
	return &PoolStats{
		Hits:         uint32(stats.Hits),
		Misses:       uint32(stats.Misses),
		Timeouts:     uint32(stats.Timeouts),
		TotalConns:   uint32(stats.TotalConns),
		IdleConns:    uint32(stats.IdleConns),
		ActiveCount:  int64(stats.TotalConns - stats.IdleConns),
		IdleCount:    int64(stats.IdleConns),
		TotalCount:   int64(stats.TotalConns),
		WaitCount:    0, // go-redis v8 不提供此信息
		TimeoutCount: int64(stats.Timeouts),
		HitCount:     int64(stats.Hits),
		MissCount:    int64(stats.Misses),
	}
}

// recordMetrics 定期记录连接池指标
func (p *pool) recordMetrics(ctx context.Context) {
	if p.metrics == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				stats := p.client.PoolStats()

				// 记录连接池指标
				p.metrics.SetPoolStats(
					float64(stats.Hits),
					float64(stats.Misses),
					float64(stats.Timeouts),
					float64(stats.TotalConns),
					float64(stats.IdleConns),
					float64(stats.StaleConns),
				)

				// 记录连接池配置指标
				if p.options != nil {
					p.metrics.SetPoolConfig(
						float64(p.options.PoolSize),
						float64(p.options.MinIdleConns),
					)
				}

				// 记录当前连接池大小
				p.metrics.SetCurrentPoolSize(float64(stats.TotalConns))
			}
		}
	}()
}

// Pool 返回连接池实例
func (c *client) Pool() Pool {
	if c.client == nil {
		c.logger.Error(context.Background(), "redis client is nil")
		return nil
	}
	return NewPool(c.client, c.logger, c.options, c.metrics)
}

// Close 关闭连接池
func (p *pool) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	switch c := p.client.(type) {
	case *redis.Client:
		err = c.Close()
	case *redis.ClusterClient:
		err = c.Close()
	default:
		p.logger.Error(ctx, "unknown redis client type")
		return errors.NewRedisInvalidConfigError("unknown redis client type", nil)
	}

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return errors.NewRedisTimeoutError("connection pool close timeout", err)
		}

		if isPoolExhaustedError(err) {
			return errors.NewRedisPoolExhaustedError("connection pool exhausted", err)
		}

		p.logger.WithError(err).WithFields(types.Field{
			Key:   "timeout",
			Value: "5s",
		}).Error(ctx, "failed to close redis connection pool")
		return errors.NewRedisConnError("failed to close redis connection pool", err)
	}

	return nil
}

// isPoolExhaustedError 检查是否是连接池耗尽错误
func isPoolExhaustedError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "pool exhausted") ||
		strings.Contains(err.Error(), "connection pool")
}
