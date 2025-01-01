package events

import "gobase/pkg/monitor/prometheus/metric"

// PublisherOption 发布器配置选项
type PublisherOption func(*Publisher)

// SubscriberOption 订阅器配置选项
type SubscriberOption func(*Subscriber)

// WithPublisherMetrics 设置发布器指标收集器
func WithPublisherMetrics(metrics *metric.Counter) PublisherOption {
	return func(p *Publisher) {
		p.metrics = metrics
	}
}

// WithSubscriberMetrics 设置订阅器指标收集器
func WithSubscriberMetrics(metrics *metric.Counter) SubscriberOption {
	return func(s *Subscriber) {
		s.metrics = metrics
	}
}

// WithChannel 设置频道名
func WithChannel(channel string) PublisherOption {
	return func(p *Publisher) {
		p.channel = channel
	}
}
