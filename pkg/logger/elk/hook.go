package elk

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ElkHook 实现 logrus.Hook 接口
type ElkHook struct {
	client      Client
	levels      []logrus.Level
	formatter   logrus.Formatter
	index       string
	buffer      BulkProcessor
	ctx         context.Context
	cancel      context.CancelFunc
	errorLogger logrus.FieldLogger
}

// ElkHookOptions ELK Hook的配置选项
type ElkHookOptions struct {
	// Elasticsearch配置
	Config *ElkConfig
	// 日志级别，为空则使用所有级别
	Levels []logrus.Level
	// 索引名称，为空则使用默认值 "logs"
	Index string
	// 批量处理配置
	BatchConfig *BulkProcessorConfig
	// 错误日志记录器
	ErrorLogger logrus.FieldLogger
}

// NewElkHook 创建新的ELK Hook
func NewElkHook(opts ElkHookOptions) (*ElkHook, error) {
	if opts.Config == nil {
		opts.Config = DefaultElkConfig()
	}
	if opts.Index == "" {
		opts.Index = "logs"
	}
	if opts.BatchConfig == nil {
		opts.BatchConfig = &BulkProcessorConfig{
			BatchSize:  100,
			FlushBytes: 5 * 1024 * 1024, // 5MB
			Interval:   time.Second * 5,
			RetryCount: 3,
			RetryWait:  time.Second,
		}
	}
	if opts.ErrorLogger == nil {
		opts.ErrorLogger = logrus.StandardLogger()
	}

	client := NewElkClient()
	if err := client.Connect(opts.Config); err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	hook := &ElkHook{
		client:      client,
		levels:      opts.Levels,
		index:       opts.Index,
		ctx:         ctx,
		cancel:      cancel,
		errorLogger: opts.ErrorLogger,
		formatter:   &logrus.JSONFormatter{},
	}

	// 初始化批量处理器
	hook.buffer = NewBulkProcessor(client, opts.BatchConfig)

	return hook, nil
}

// Levels 实现 logrus.Hook 接口
func (h *ElkHook) Levels() []logrus.Level {
	if len(h.levels) == 0 {
		return logrus.AllLevels
	}
	return h.levels
}

// Fire 实现 logrus.Hook 接口
func (h *ElkHook) Fire(entry *logrus.Entry) error {
	data, err := h.formatter.Format(entry)
	if err != nil {
		return fmt.Errorf("failed to format log entry: %w", err)
	}

	// 使用批量处理器添加文档
	err = h.buffer.Add(h.ctx, h.index, data)
	if err != nil {
		h.errorLogger.WithError(err).Error("failed to add log entry to bulk processor")
		return err
	}

	return nil
}

// Close 关闭Hook，确保所有日志都被发送
func (h *ElkHook) Close() error {
	h.cancel()

	if err := h.buffer.Close(); err != nil {
		return fmt.Errorf("failed to close bulk processor: %w", err)
	}

	if err := h.client.Close(); err != nil {
		return fmt.Errorf("failed to close elasticsearch client: %w", err)
	}

	return nil
}

// GetBulkProcessor 返回内部的 BulkProcessor
func (h *ElkHook) GetBulkProcessor() BulkProcessor {
	return h.buffer
}
