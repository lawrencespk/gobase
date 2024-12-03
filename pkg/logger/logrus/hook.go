package logrus

import (
	"gobase/pkg/logger/elk"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

// elasticHook Elasticsearch钩子
type elasticHook struct {
	client *elk.ElkClient // Elasticsearch客户端
	levels []logrus.Level // 日志级别
}

// newElasticHook 创建Elasticsearch钩子
func newElasticHook(opts *Options) (logrus.Hook, error) {
	client, err := elk.NewElkClient(
		opts.ElasticURLs,  // Elasticsearch URL
		opts.ElasticIndex, // Elasticsearch 索引
		&elk.Options{
			Workers:       2,                // 工作线程数
			BatchSize:     100,              // 批量大小
			FlushInterval: 5 * time.Second,  // 刷新间隔
			RetryTimeout:  30 * time.Second, // 重试超时
		},
	)
	if err != nil {
		return nil, err
	}

	return &elasticHook{
		client: client,
		levels: []logrus.Level{
			logrus.InfoLevel,  // 信息级别
			logrus.WarnLevel,  // 警告级别
			logrus.ErrorLevel, // 错误级别
			logrus.FatalLevel, // 严重级别
		},
	}, nil
}

// Levels 返回日志级别
func (h *elasticHook) Levels() []logrus.Level {
	return h.levels // 返回日志级别
}

// Fire 处理日志
func (h *elasticHook) Fire(entry *logrus.Entry) error {
	data := make(map[string]interface{}) // 创建数据

	// 复制字段
	for k, v := range entry.Data {
		data[k] = v // 复制字段
	}

	// 添加基本信息
	data["message"] = entry.Message      // 日志消息
	data["level"] = entry.Level.String() // 日志级别
	data["timestamp"] = entry.Time       // 日志时间

	return h.client.Write(data)
}

// LogHook 通用日志钩子
type LogHook struct {
	writer    io.Writer        // 日志输出writer
	formatter logrus.Formatter // 日志格式化器
	levels    []logrus.Level   // 支持的日志级别
}

// NewLogHook 创建新的日志钩子
func NewLogHook(writer io.Writer, formatter logrus.Formatter, levels ...logrus.Level) logrus.Hook {
	return &LogHook{
		writer:    writer,
		formatter: formatter,
		levels:    levels,
	}
}

// Levels 返回支持的日志级别
func (h *LogHook) Levels() []logrus.Level {
	return h.levels
}

// Fire 处理日志事件
func (h *LogHook) Fire(entry *logrus.Entry) error {
	bytes, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}
	_, err = h.writer.Write(bytes)
	return err
}
