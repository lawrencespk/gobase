package logrus

import (
	"context"
	"gobase/pkg/logger/elk"

	"github.com/sirupsen/logrus"
)

// Hook 是一个 Logrus 的 Hook，用于将日志写入 Elasticsearch
type Hook struct {
	client elk.Client
	levels []logrus.Level
}

// NewHook 创建一个新的 Hook
func NewHook() (*Hook, error) {
	client := elk.NewElkClient()
	return &Hook{client: client, levels: logrus.AllLevels}, nil
}

// NewHookWithClient 使用提供的客户端创建 Hook（用于测试）
func NewHookWithClient(client elk.Client) *Hook {
	return &Hook{client: client, levels: logrus.AllLevels}
}

// SetLevels 设置 Hook 适用的日志级别
func (h *Hook) SetLevels(levels []logrus.Level) {
	h.levels = levels
}

// Fire 将日志条目写入 Elasticsearch
func (h *Hook) Fire(entry *logrus.Entry) error {
	// 将日志条目转换为文档
	document := map[string]interface{}{
		"level":   entry.Level.String(),
		"message": entry.Message,
		"time":    entry.Time,
		"data":    entry.Data,
	}

	// 使用 IndexDocument 方法而不是 Write
	return h.client.IndexDocument(context.Background(), "logs", document)
}

// Levels 返回 Hook 适用的日志级别
func (h *Hook) Levels() []logrus.Level {
	return h.levels
}
