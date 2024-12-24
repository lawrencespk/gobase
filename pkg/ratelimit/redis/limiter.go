package redis

import (
	"context"
	"time"

	"gobase/pkg/cache/redis/client"
	"gobase/pkg/errors"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metric"
	"gobase/pkg/ratelimit/core"
	"gobase/pkg/ratelimit/metrics"
)

var (
	// RequestsTotal 记录所有限流请求
	RequestsTotal = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "ratelimit",
		Name:      "requests_total",
		Help:      "Total number of requests handled by rate limiter",
	}).WithLabels("key", "result") // result: allowed/rejected

	// RejectedTotal 记录被限流的请求
	RejectedTotal = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "ratelimit",
		Name:      "rejected_total",
		Help:      "Total number of requests rejected by rate limiter",
	}).WithLabels("key")

	// LimiterLatency 记录限流器处理延迟
	LimiterLatency = metric.NewHistogram(metric.HistogramOpts{
		Namespace: "gobase",
		Subsystem: "ratelimit",
		Name:      "latency_seconds",
		Help:      "Latency of rate limiter operations",
		Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}).WithLabels("key", "operation") // operation: allow/wait
)

// 滑动窗口限流器实现
type slidingWindowLimiter struct {
	client client.Client
	opts   *core.LimiterOptions
	log    types.Logger
}

// 创建新的滑动窗口限流器
func NewSlidingWindowLimiter(client client.Client, opts ...core.LimiterOption) core.Limiter {
	options := &core.LimiterOptions{
		Name:      "sliding_window",
		Algorithm: "sliding_window",
	}

	for _, opt := range opts {
		opt(options)
	}

	// 使用 logger 包的默认 logger 并添加字段
	defaultLogger := logger.GetLogger().WithFields(
		types.Field{Key: "module", Value: "ratelimit"},
		types.Field{Key: "type", Value: "sliding_window"},
	)

	limiter := &slidingWindowLimiter{
		client: client,
		opts:   options,
		log:    defaultLogger,
	}

	// 更新活跃限流器计数
	metrics.Collector.SetActiveLimiters("sliding_window", 1)
	limiter.log.Info(context.Background(), "created new sliding window rate limiter")

	return limiter
}

// Allow 实现限流判断
func (l *slidingWindowLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	return l.AllowN(ctx, key, 1, limit, window)
}

// AllowN 判断是否允许N个请求通过
func (l *slidingWindowLimiter) AllowN(ctx context.Context, key string, n int64, limit int64, window time.Duration) (bool, error) {
	start := time.Now()
	defer func() {
		metrics.Collector.ObserveLatency(key, "allow", time.Since(start).Seconds())
	}()

	l.log.Debug(ctx, "checking rate limit",
		types.Field{Key: "key", Value: key},
		types.Field{Key: "n", Value: n},
		types.Field{Key: "limit", Value: limit},
		types.Field{Key: "window", Value: window},
	)

	// 获取当前时间戳（毫秒级别即可）
	now := time.Now().UnixMilli()
	counterKey := key + ":counter"

	// 清理过期的数据并增加计数
	script := `
        local key = KEYS[1]
        local counter_key = KEYS[2]
        local now = tonumber(ARGV[1])
        local window = tonumber(ARGV[2])
        local limit = tonumber(ARGV[3])
        local n = tonumber(ARGV[4])
        
        -- 清理过期数据
        redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window)
        
        -- 原子递增计数器
        local new_count = redis.call('INCRBY', counter_key, n)
        
        -- 如果超过限制，回滚计数并返回
        if new_count > limit then
            redis.call('DECRBY', counter_key, n)
            return 0
        end
        
        -- 添加新请求
        for i = 1, n do
            redis.call('ZADD', key, now, now .. ':' .. i)
        end
        
        -- 设置过期时间
        local expire_time = math.ceil(window/1000) + 1
        redis.call('EXPIRE', key, expire_time)
        redis.call('EXPIRE', counter_key, expire_time)
        
        return 1
    `

	// 执行Redis Lua脚本
	result, err := l.client.Eval(ctx, script, []string{key, counterKey},
		now,                   // 当前时间戳（毫秒）
		window.Milliseconds(), // 窗口大小（毫秒）
		limit,                 // 限制数量
		n,                     // 请求数量
	)
	if err != nil {
		l.log.Error(ctx, "failed to evaluate rate limit script",
			types.Field{Key: "key", Value: key},
			types.Field{Key: "error", Value: err},
		)
		return false, errors.Wrap(err, "failed to evaluate rate limit script")
	}

	allowed := result.(int64) == 1
	metrics.Collector.ObserveRequest(key, allowed)

	if !allowed {
		l.log.Debug(ctx, "rate limit exceeded",
			types.Field{Key: "key", Value: key},
			types.Field{Key: "limit", Value: limit},
		)
	}

	return allowed, nil
}

// Wait 等待直到允许通过或超时
func (l *slidingWindowLimiter) Wait(ctx context.Context, key string, limit int64, window time.Duration) error {
	start := time.Now()
	defer func() {
		metrics.Collector.ObserveLatency(key, "wait", time.Since(start).Seconds())
	}()

	l.log.Debug(ctx, "starting wait for rate limit",
		types.Field{Key: "key", Value: key},
		types.Field{Key: "limit", Value: limit},
		types.Field{Key: "window", Value: window},
	)

	waitingCount := float64(0)
	defer func() {
		metrics.Collector.SetWaitingQueueSize(key, waitingCount)
	}()

	for {
		select {
		case <-ctx.Done():
			l.log.Debug(ctx, "wait cancelled by context",
				types.Field{Key: "key", Value: key},
				types.Field{Key: "error", Value: ctx.Err()},
			)
			return ctx.Err()
		default:
			waitingCount++
			metrics.Collector.SetWaitingQueueSize(key, waitingCount)

			allowed, err := l.Allow(ctx, key, limit, window)
			if err != nil {
				l.log.Error(ctx, "error while waiting for rate limit",
					types.Field{Key: "error", Value: err},
					types.Field{Key: "key", Value: key},
				)
				return err
			}
			if allowed {
				l.log.Debug(ctx, "rate limit wait completed",
					types.Field{Key: "key", Value: key},
					types.Field{Key: "waited_cycles", Value: waitingCount},
				)
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// Reset 重置限流器
func (l *slidingWindowLimiter) Reset(ctx context.Context, key string) error {
	l.log.Info(ctx, "resetting rate limiter",
		types.Field{Key: "key", Value: key},
	)

	err := l.client.Del(ctx, key)
	if err != nil {
		l.log.Error(ctx, "failed to reset rate limiter",
			types.Field{Key: "error", Value: err},
			types.Field{Key: "key", Value: key},
		)
		return errors.Wrap(err, "failed to reset rate limiter")
	}

	metrics.Collector.SetWaitingQueueSize(key, 0)
	l.log.Info(ctx, "rate limiter reset successful",
		types.Field{Key: "key", Value: key},
	)

	return nil
}

// ... 实现其他接口方法 ...
