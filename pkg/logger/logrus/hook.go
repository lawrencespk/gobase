package logrus

import (
	"context"
	"gobase/pkg/errors"
	"gobase/pkg/logger/elk"
	"time"

	"gobase/pkg/monitor/prometheus/metrics"

	"github.com/sirupsen/logrus"
)

// Hook 是一个 Logrus 的 Hook，用于将日志写入 Elasticsearch
type Hook struct {
	client elk.Client
	levels []logrus.Level
}

// NewHook 创建一个新的 Hook
func NewHook() (*Hook, error) {
	client := elk.NewElkClient() // 创建 ELK 客户端
	if client == nil {
		return nil, errors.NewELKConnectionError("failed to create ELK client", nil) // 创建 ELK 客户端失败
	}
	return &Hook{client: client, levels: logrus.AllLevels}, nil // 返回 Hook
}

// NewHookWithClient 使用提供的客户端创建 Hook（用于测试）
func NewHookWithClient(client elk.Client) *Hook { // 使用提供的客户端创建 Hook（用于测试）
	return &Hook{client: client, levels: logrus.AllLevels} // 返回 Hook
}

// SetLevels 设置 Hook 适用的日志级别
func (h *Hook) SetLevels(levels []logrus.Level) {
	h.levels = levels
}

// Fire 将日志条目写入 Elasticsearch
func (h *Hook) Fire(entry *logrus.Entry) error {
	// 记录日志计数
	metrics.LogCounter.WithLabelValues(entry.Level.String()).Inc()

	start := time.Now()
	defer func() {
		// 记录处理延迟
		metrics.LogLatency.WithLabelValues("write").Observe(time.Since(start).Seconds())
	}()

	// 将日志条目转换为文档
	document := map[string]interface{}{
		"level":   entry.Level.String(), // 日志级别
		"message": entry.Message,        // 日志消息
		"time":    entry.Time,           // 日志时间
		"data":    entry.Data,           // 日志数据
	}

	// 使用 IndexDocument 方法而不是 Write
	if err := h.client.IndexDocument(context.Background(), "logs", document); err != nil {
		return errors.NewELKIndexError("failed to index log entry", err) // 写入日志失败
	}
	return nil
}

// Levels 返回 Hook 适用的日志级别
func (h *Hook) Levels() []logrus.Level {
	return h.levels
}
