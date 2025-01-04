package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"gobase/pkg/cache"
	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metric"
	"gobase/pkg/trace/jaeger"
)

// Publisher 事件发布器
type Publisher struct {
	client  redis.Client
	channel string
	logger  types.Logger
	metrics *metric.Counter
	cache   cache.Cache
}

// Option 配置选项
type Option func(*Publisher)

// WithMetrics 设置指标收集器
func WithMetrics(metrics *metric.Counter) Option {
	return func(p *Publisher) {
		p.metrics = metrics
	}
}

// WithCache 设置缓存接口
func WithCache(cache cache.Cache) Option {
	return func(p *Publisher) {
		p.cache = cache
	}
}

// NewPublisher 创建事件发布器
func NewPublisher(client redis.Client, logger types.Logger, options ...Option) (*Publisher, error) {
	pub := &Publisher{
		client:  client,
		channel: "jwt:events",
		logger:  logger,
	}

	for _, opt := range options {
		opt(pub)
	}

	return pub, nil
}

// Publish 发布事件
func (p *Publisher) Publish(ctx context.Context, eventType EventType, data map[string]interface{}) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "events.publish")
	if span != nil {
		defer span.Finish()
	}

	// 创建事件
	event := &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		Payload:   data,
	}

	// 序列化事件
	payload, err := json.Marshal(event)
	if err != nil {
		return errors.NewSerializationError("failed to marshal event", err)
	}

	// 如果配置了缓存，先写入缓存
	if p.cache != nil {
		if p.cache.GetLevel() >= cache.Level(2) {
			cacheKey := fmt.Sprintf("event:%s", event.ID)
			if err := p.cache.Set(ctx, cacheKey, payload, time.Hour); err != nil {
				p.logger.WithContext(ctx).WithCaller(1).Warn(ctx, "failed to cache event",
					types.Field{Key: "event_id", Value: event.ID},
					types.Field{Key: "error", Value: err},
				)
				// 缓存失败不影响主流程
			}
		}
	}

	// 发布事件
	if err := p.client.Publish(ctx, p.channel, string(payload)); err != nil {
		p.logger.WithContext(ctx).WithCaller(1).Error(ctx, "failed to publish event",
			types.Field{Key: "event_id", Value: event.ID},
			types.Field{Key: "event_type", Value: string(event.Type)},
			types.Field{Key: "error", Value: err},
		)
		return errors.NewThirdPartyError("failed to publish event to redis", err)
	}

	// 记录成功日志
	p.logger.WithContext(ctx).WithCaller(1).Info(ctx, "event published successfully",
		types.Field{Key: "event_id", Value: event.ID},
		types.Field{Key: "event_type", Value: string(event.Type)},
	)

	// 记录指标
	if p.metrics != nil {
		p.metrics.WithLabelValues(string(event.Type), "success").Inc()
	}

	return nil
}
