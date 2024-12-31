package redis

import (
	"context"
	"time"

	"gobase/pkg/cache/redis/ratelimit"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
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
	store *ratelimit.Store
	opts  *core.LimiterOptions
	log   types.Logger
}

// 创建新的滑动窗口限流器
func NewSlidingWindowLimiter(store *ratelimit.Store, opts ...core.LimiterOption) core.Limiter {
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
		store: store,
		opts:  options,
		log:   defaultLogger,
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

	// 修改 Lua 脚本
	script := `
        local key = KEYS[1]
        local counter_key = KEYS[2]
        local now = tonumber(ARGV[1])
        local window = tonumber(ARGV[2])
        local limit = tonumber(ARGV[3])
        local n = tonumber(ARGV[4])
        
        -- 清理过期数据
        redis.call('ZREMRANGEBYSCORE', key, '-inf', now - window)
        
        -- 获取当前窗口内的请求总数
        local total = 0
        local members = redis.call('ZRANGE', key, 0, -1, 'WITHSCORES')
        for i = 1, #members, 2 do
            local count = tonumber(redis.call('HGET', counter_key, members[i]))
            if count then
                total = total + count
            end
        end
        
        -- 检查是否超过限制
        if (total + n) > limit then
            return 0
        end
        
        -- 添加新请求记录
        local member = tostring(now)
        redis.call('ZADD', key, now, member)
        redis.call('HINCRBY', counter_key, member, n)
        
        -- 设置过期时间
        redis.call('EXPIRE', key, math.ceil(window/1000) + 1)
        redis.call('EXPIRE', counter_key, math.ceil(window/1000) + 1)
        
        return 1
    `

	// 执行Redis Lua脚本
	result, err := l.store.Eval(ctx, script, []string{key, counterKey},
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

	// 获取 context 的截止时间
	deadline, hasDeadline := ctx.Deadline()
	maxWaitTime := window // 默认最大等待时间为一个完整窗口
	if hasDeadline {
		if remaining := time.Until(deadline); remaining < window {
			maxWaitTime = remaining
		}
	}

	// 调整基础睡眠时间计算
	baseSleep := window / time.Duration(limit*2) // 增加基础等待时间
	if baseSleep < time.Millisecond {
		baseSleep = time.Millisecond
	}
	maxSleep := window / 4 // 增加最大等待时间

	retryCount := 0
	totalWaitTime := time.Duration(0)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			allowed, err := l.Allow(ctx, key, limit, window)
			if err != nil {
				return err
			}
			if allowed {
				return nil
			}

			// 检查总等待时间是否超过最大等待时间
			if totalWaitTime >= maxWaitTime {
				return errors.NewError(codes.TooManyRequests, "wait timeout exceeded", nil)
			}

			// 使用指数退避策略，但有最大值限制
			sleepTime := baseSleep * time.Duration(1<<uint(retryCount))
			if sleepTime > maxSleep {
				sleepTime = maxSleep
			}

			// 确保不会超过最大等待时间
			if totalWaitTime+sleepTime > maxWaitTime {
				sleepTime = maxWaitTime - totalWaitTime
			}

			retryCount++
			totalWaitTime += sleepTime
			metrics.Collector.SetWaitingQueueSize(key, float64(retryCount))

			timer := time.NewTimer(sleepTime)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
				continue
			}
		}
	}
}

// Reset 重置限流器
func (l *slidingWindowLimiter) Reset(ctx context.Context, key string) error {
	l.log.Info(ctx, "resetting rate limiter",
		types.Field{Key: "key", Value: key},
	)

	counterKey := key + ":counter"

	// 删除主key和计数器key
	err := l.store.Del(ctx, key, counterKey)
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
