package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"gobase/pkg/cache"
	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metric"
)

// Subscriber JWT事件订阅器
type Subscriber struct {
	client   redis.Client
	channel  string
	handlers map[EventType]EventHandler
	mu       sync.RWMutex
	logger   types.Logger
	metrics  *metric.Counter
	cache    cache.Cache
}

// WithSubscriberLogger 设置日志记录器
func WithSubscriberLogger(logger types.Logger) SubscriberOption {
	return func(s *Subscriber) {
		s.logger = logger
	}
}

// WithSubscriberCache 设置缓存
func WithSubscriberCache(cache cache.Cache) SubscriberOption {
	return func(s *Subscriber) {
		s.cache = cache
	}
}

// NewSubscriber 创建新的订阅器
func NewSubscriber(client redis.Client, opts ...SubscriberOption) *Subscriber {
	s := &Subscriber{
		client:   client,
		channel:  "jwt:events",
		handlers: make(map[EventType]EventHandler),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// RegisterHandler 注册事件处理器
func (s *Subscriber) RegisterHandler(eventType EventType, handler EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[eventType] = handler
}

// Subscribe 订阅事件
func (s *Subscriber) Subscribe(ctx context.Context) error {
	pubsub := s.client.Subscribe(ctx, s.channel)
	defer pubsub.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := pubsub.ReceiveMessage(ctx)
			if err != nil {
				return errors.NewRedisCommandError("failed to receive message", err)
			}

			// 解析事件
			var event Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				if s.logger != nil {
					s.logger.Error(ctx, "failed to unmarshal event",
						types.Field{Key: "error", Value: err},
						types.Field{Key: "payload", Value: msg.Payload},
					)
				}
				if s.metrics != nil {
					s.metrics.WithLabels("type", "unknown", "status", "unmarshal_error").Inc()
				}
				continue
			}

			// 如果配置了缓存，尝试从缓存获取完整事件数据
			if s.cache != nil {
				cacheKey := fmt.Sprintf("event:%s", event.ID)
				if cachedData, err := s.cache.Get(ctx, cacheKey); err == nil {
					if data, ok := cachedData.([]byte); ok {
						if err := json.Unmarshal(data, &event); err != nil {
							s.logger.Warn(ctx, "failed to unmarshal cached event",
								types.Field{Key: "event_id", Value: event.ID},
								types.Field{Key: "error", Value: err},
							)
						}
					} else {
						s.logger.Warn(ctx, "invalid cache data type",
							types.Field{Key: "event_id", Value: event.ID},
							types.Field{Key: "type", Value: fmt.Sprintf("%T", cachedData)},
						)
					}
				}
			}

			// 处理事件
			s.mu.RLock()
			handler, exists := s.handlers[event.Type]
			s.mu.RUnlock()

			if !exists {
				if s.logger != nil {
					s.logger.Warn(ctx, "no handler registered for event type",
						types.Field{Key: "event_type", Value: event.Type},
					)
				}
				if s.metrics != nil {
					s.metrics.WithLabels("type", string(event.Type), "status", "no_handler").Inc()
				}
				continue
			}

			if err := handler(&event); err != nil {
				if s.logger != nil {
					s.logger.Error(ctx, "failed to handle event",
						types.Field{Key: "event_type", Value: event.Type},
						types.Field{Key: "error", Value: err},
					)
				}
				if s.metrics != nil {
					s.metrics.WithLabels("type", string(event.Type), "status", "error").Inc()
				}
				continue
			}

			if s.metrics != nil {
				s.metrics.WithLabels("type", string(event.Type), "status", "success").Inc()
			}
		}
	}
}
