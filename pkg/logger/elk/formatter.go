package elk

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

// ElkFormatter 实现适合 Elasticsearch 的日志格式化
type ElkFormatter struct {
	TimestampFormat string
}

// Format 将 logrus 条目转换为适合 Elasticsearch 的格式
func (f *ElkFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(map[string]interface{})

	// 基础字段
	data["@timestamp"] = entry.Time.UTC().Format(time.RFC3339)
	data["level"] = entry.Level.String()
	data["message"] = entry.Message

	// 调用者信息
	if entry.HasCaller() {
		data["caller"] = map[string]interface{}{
			"file":     entry.Caller.File,
			"line":     entry.Caller.Line,
			"function": entry.Caller.Function,
		}
	}

	// 错误信息
	if err, ok := entry.Data["error"]; ok {
		if err, ok := err.(error); ok {
			data["error"] = map[string]interface{}{
				"message": err.Error(),
			}
			// 如果错误实现了堆栈接口
			if stackTracer, ok := err.(interface{ StackTrace() []string }); ok {
				data["error"].(map[string]interface{})["stack_trace"] = stackTracer.StackTrace()
			}
		}
	}

	// 添加其他字段
	for k, v := range entry.Data {
		if k != "error" {
			data[k] = v
		}
	}

	return json.Marshal(data)
}
