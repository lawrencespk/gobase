package elk

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

// ElkHook 实现 logrus.Hook 接口
type ElkHook struct {
	client      Client
	levels      []logrus.Level
	formatter   logrus.Formatter
	index       string
	processor   BulkProcessor
	ctx         context.Context
	cancel      context.CancelFunc
	errorLogger logrus.FieldLogger
}

// ElkHookOptions ELK Hook的配置选项
type ElkHookOptions struct {
	Config      *ElkConfig
	Levels      []logrus.Level
	Index       string
	BatchConfig *BulkProcessorConfig
	ErrorLogger logrus.FieldLogger
	MaxDocSize  int64  // 单个文档最大大小
	IndexPrefix string // 索引前缀
	IndexSuffix string // 索引后缀
}

// validateOptions 验证并规范化配置选项
func validateOptions(opts *ElkHookOptions) error {
	if opts.Config == nil {
		opts.Config = DefaultElkConfig()
	}

	if opts.Index == "" {
		if opts.IndexPrefix == "" {
			opts.IndexPrefix = "logs"
		}
		if opts.IndexSuffix == "" {
			opts.IndexSuffix = time.Now().Format("2006.01.02")
		}
		opts.Index = fmt.Sprintf("%s-%s", opts.IndexPrefix, opts.IndexSuffix)
	}

	if opts.BatchConfig == nil {
		opts.BatchConfig = &BulkProcessorConfig{
			BatchSize:    50,
			FlushBytes:   512 * 1024,
			Interval:     100 * time.Millisecond,
			RetryCount:   5,
			RetryWait:    time.Second,
			CloseTimeout: 10 * time.Second,
		}
	} else {
		if err := validateBatchConfig(opts.BatchConfig); err != nil {
			return errors.Wrap(err, "invalid batch config")
		}
	}

	if opts.MaxDocSize <= 0 {
		opts.MaxDocSize = 5 * 1024 * 1024 // 默认5MB
	}

	if opts.ErrorLogger == nil {
		opts.ErrorLogger = logrus.StandardLogger()
	}

	return nil
}

// validateBatchConfig 验证批处理配置
func validateBatchConfig(cfg *BulkProcessorConfig) error {
	if cfg.BatchSize <= 0 || cfg.BatchSize > 1000 {
		return errors.WrapWithCode(nil, codes.ELKConfigError, "invalid batch size")
	}
	if cfg.FlushBytes <= 0 {
		return errors.WrapWithCode(nil, codes.ELKConfigError, "invalid flush bytes")
	}
	if cfg.Interval <= 0 {
		return errors.WrapWithCode(nil, codes.ELKConfigError, "invalid interval")
	}
	if cfg.RetryCount < 3 {
		return errors.WrapWithCode(nil, codes.ELKConfigError, "retry count too low")
	}
	if cfg.RetryWait <= 0 {
		return errors.WrapWithCode(nil, codes.ELKConfigError, "invalid retry wait")
	}
	if cfg.CloseTimeout <= 0 {
		return errors.WrapWithCode(nil, codes.ELKConfigError, "invalid close timeout")
	}
	return nil
}

// NewElkHook 创建新的ELK Hook
func NewElkHook(opts ElkHookOptions) (*ElkHook, error) {
	if err := validateOptions(&opts); err != nil {
		return nil, errors.WrapWithCode(err, codes.ELKConfigError, "invalid options")
	}

	client := NewElkClient()
	if err := client.Connect(opts.Config); err != nil {
		return nil, errors.WrapWithCode(err, codes.ELKConnectionError, "failed to connect to Elasticsearch")
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

	processor := NewBulkProcessor(client, opts.BatchConfig)
	hook.processor = processor

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
	// 创建日志文档
	doc := map[string]interface{}{
		"level": entry.Level.String(),
		"msg":   entry.Message,
		"time":  entry.Time,
	}

	// 添加额外的字段
	for k, v := range entry.Data {
		doc[k] = v
	}

	// 检查文档大小（使用 JSON 编码后的大小）
	docBytes, err := json.Marshal(doc)
	if err != nil {
		return errors.WrapWithCode(err, codes.SerializationError, "failed to marshal log entry")
	}

	maxSize := h.processor.MaxDocSize()
	if len(docBytes) > int(maxSize) {
		return errors.WrapWithCode(nil, codes.ELKBulkError, "document size exceeds limit")
	}

	// 使用批量处理器添加文档对象
	if err := h.processor.Add(h.ctx, h.index, doc); err != nil {
		h.errorLogger.WithError(err).Error("failed to add log entry to bulk processor")
		return errors.WrapWithCode(err, codes.ELKBulkError, "failed to add log entry")
	}

	return nil
}

// Close 关闭Hook，确保所有日志都被发送
func (h *ElkHook) Close() error {
	h.cancel()

	if err := h.processor.Close(); err != nil {
		return errors.WrapWithCode(err, codes.ELKBulkError, "failed to close bulk processor")
	}

	if err := h.client.Close(); err != nil {
		return errors.WrapWithCode(err, codes.ELKConnectionError, "failed to close elasticsearch client")
	}

	return nil
}

// GetBulkProcessor 返回内部的 BulkProcessor
func (h *ElkHook) GetBulkProcessor() BulkProcessor {
	return h.processor
}

// GetStats 获取Hook的统计信息
func (h *ElkHook) GetStats() *HookStats {
	procStats := h.processor.Stats()
	return &HookStats{
		ProcessorStats: procStats,
		Index:          h.index,
		ActiveLevels:   len(h.levels),
	}
}

// HookStats Hook的统计信息
type HookStats struct {
	ProcessorStats *BulkStats
	Index          string
	ActiveLevels   int
}
