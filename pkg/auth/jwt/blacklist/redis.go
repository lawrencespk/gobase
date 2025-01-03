package blacklist

import (
	"context"
	"time"

	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metric"
	"gobase/pkg/trace/jaeger"
)

// RedisBlacklist Redis实现的JWT黑名单
type RedisBlacklist struct {
	client  redis.Client
	prefix  string
	logger  types.Logger
	metrics struct {
		tokenCount *metric.Gauge
		addTotal   *metric.Counter
		hitTotal   *metric.Counter
		missTotal  *metric.Counter
	}
}

// NewRedisBlacklist 创建Redis黑名单实例
func NewRedisBlacklist(client redis.Client, opts *Options) (*RedisBlacklist, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	bl := &RedisBlacklist{
		client: client,
		prefix: opts.KeyPrefix + "blacklist:",
		logger: opts.Log,
	}

	// 初始化监控指标
	if opts.EnableMetrics {
		bl.metrics.tokenCount = metric.NewGauge(metric.GaugeOpts{
			Namespace: "gobase",
			Subsystem: "jwt_blacklist",
			Name:      "token_count",
			Help:      "Number of tokens in blacklist",
		})
		bl.metrics.addTotal = metric.NewCounter(metric.CounterOpts{
			Namespace: "gobase",
			Subsystem: "jwt_blacklist",
			Name:      "add_total",
			Help:      "Total number of tokens added to blacklist",
		})
		// ... 其他指标初始化
	}

	return bl, nil
}

// Add 将Token添加到黑名单
func (bl *RedisBlacklist) Add(ctx context.Context, token string, expiration time.Duration) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "blacklist.redis.add")
	defer span.Finish()

	key := bl.prefix + token
	if err := bl.client.Set(ctx, key, "1", expiration); err != nil {
		bl.logger.Error(ctx, "failed to add token to blacklist",
			types.Field{Key: "token", Value: token},
			types.Field{Key: "error", Value: err},
		)
		return errors.NewCacheError("failed to add token to blacklist", err)
	}

	if bl.metrics.tokenCount != nil {
		bl.metrics.tokenCount.Inc()
	}
	if bl.metrics.addTotal != nil {
		bl.metrics.addTotal.Inc()
	}

	return nil
}

// IsBlacklisted 检查Token是否在黑名单中
func (bl *RedisBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	span, ctx := jaeger.StartSpanFromContext(ctx, "blacklist.redis.check")
	defer span.Finish()

	key := bl.prefix + token
	exists, err := bl.client.Exists(ctx, key)
	if err != nil {
		return false, errors.NewCacheError("failed to check token blacklist", err)
	}

	if exists {
		if bl.metrics.hitTotal != nil {
			bl.metrics.hitTotal.Inc()
		}
		return true, nil
	}

	if bl.metrics.missTotal != nil {
		bl.metrics.missTotal.Inc()
	}
	return false, nil
}
