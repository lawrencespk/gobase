package logrus

import (
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

type customFormatter struct {
	TimestampFormat string // 时间格式
	PrettyPrint     bool   // 是否美化打印
}

func newFormatter(opts *Options) logrus.Formatter {
	return &customFormatter{
		TimestampFormat: time.RFC3339,     // 时间格式
		PrettyPrint:     opts.Development, // 是否美化打印
	}
}

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(map[string]interface{})

	// 时间格式
	if f.TimestampFormat != "" {
		data["timestamp"] = entry.Time.Format(f.TimestampFormat)
	} else {
		data["timestamp"] = entry.Time.Format(time.RFC3339)
	}

	data["level"] = entry.Level.String() // 日志级别
	data["message"] = entry.Message      // 日志消息

	// 添加调用者信息
	if entry.HasCaller() {
		data["caller"] = map[string]interface{}{
			"function": entry.Caller.Function, // 函数名
			"file":     entry.Caller.File,     // 文件名
			"line":     entry.Caller.Line,     // 行号
		}
	}

	// 错误字段特殊处理
	if err, ok := entry.Data["error"]; ok {
		if customErr, ok := err.(interface{ Code() string }); ok {
			data["error_code"] = customErr.Code()
		}
		if customErr, ok := err.(interface{ Stack() []string }); ok {
			data["error_stack"] = customErr.Stack()
		}
	}

	// 添加字段
	for k, v := range entry.Data {
		if k == "error" {
			if err, ok := v.(error); ok {
				data[k] = err.Error()
			} else {
				data[k] = v
			}
			continue
		}
		data[k] = v
	}

	if f.PrettyPrint {
		return json.MarshalIndent(data, "", "    ")
	}
	return json.Marshal(data)
}
