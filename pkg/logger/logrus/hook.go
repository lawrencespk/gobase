package logrus

import (
	"fmt"
	"io"
	"time"

	"gobase/pkg/logger/elk"

	"github.com/sirupsen/logrus"
)

// ElasticHook Elasticsearch钩子
type ElasticHook struct {
	client *elk.ElkClient // 客户端
	levels []logrus.Level // 日志级别
}

// NewElasticHook 创建Elasticsearch钩子
func NewElasticHook(config *elk.ElasticConfig, levels ...logrus.Level) (logrus.Hook, error) {
	client, err := elk.NewElkClient(config) // 创建Elk客户端
	if err != nil {
		return nil, fmt.Errorf("failed to create elk client: %w", err) // 创建Elk客户端失败
	}

	return &ElasticHook{
		client: client, // 客户端
		levels: levels, // 日志级别
	}, nil
}

// Levels 返回支持的日志级别
func (h *ElasticHook) Levels() []logrus.Level {
	return h.levels
}

// Fire 处理日志事件
func (h *ElasticHook) Fire(entry *logrus.Entry) error {
	data := make(map[string]interface{})

	// 添加基础字段
	data["timestamp"] = entry.Time.UTC().Format(time.RFC3339) // 设置时间
	data["level"] = entry.Level.String()                      // 设置日志级别
	data["message"] = entry.Message                           // 设置消息

	// 添加调用信息
	if entry.HasCaller() { // 如果有调用信息
		data["caller"] = map[string]interface{}{ // 设置调用信息
			"file":     entry.Caller.File,     // 文件
			"line":     entry.Caller.Line,     // 行号
			"function": entry.Caller.Function, // 函数
		}
	}

	// 添加额外字段
	for k, v := range entry.Data { // 遍历日志数据
		if k == "error" { // 如果字段为error
			if err, ok := v.(error); ok { // 如果值为error
				data[k] = err.Error() // 设置为错误信息
			} else {
				data[k] = v // 设置为值
			}
			continue
		}
		data[k] = v // 设置为值
	}

	return h.client.Write(data)
}

// Close 关闭钩子
func (h *ElasticHook) Close() error {
	return h.client.Close() // 关闭客户端
}

// LogHook 自定义日志钩子
type LogHook struct {
	writer    io.Writer        // 写入器
	formatter logrus.Formatter // 格式化器
	levels    []logrus.Level   // 日志级别
}

// NewLogHook 创建一个新的日志钩子
func NewLogHook(writer io.Writer, formatter logrus.Formatter, levels ...logrus.Level) *LogHook {
	return &LogHook{
		writer:    writer,    // 写入器
		formatter: formatter, // 格式化器
		levels:    levels,    // 日志级别
	}
}

// Levels 返回支持的日志级别
func (h *LogHook) Levels() []logrus.Level {
	return h.levels // 返回支持的日志级别
}

// Fire 处理日志事件
func (h *LogHook) Fire(entry *logrus.Entry) error {
	bytes, err := h.formatter.Format(entry) // 格式化日志
	if err != nil {
		return err
	}
	_, err = h.writer.Write(bytes) // 写入日志
	return err
}
