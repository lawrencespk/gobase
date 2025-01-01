package events

import (
	"context"
	"encoding/json"
	"sync"

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
}

// WithSubscriberLogger 设置日志记录器
func WithSubscriberLogger(logger types.Logger) SubscriberOption {
	return func(s *Subscriber) {
		s.logger = logger
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
	// 使用我们自己的redis客户端进行订阅
	pubsub := s.client.Subscribe(ctx, s.channel)
	defer pubsub.Close()

	// 处理消息
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
