package logrus

import (
	"encoding/json"
	"time"

	"gobase/pkg/errors/types"

	"github.com/sirupsen/logrus"
)

// customFormatter 自定义格式化器
type customFormatter struct {
	TimestampFormat string // 时间格式
	PrettyPrint     bool   // 是否美化打印
}

// newFormatter 创建格式化器
func newFormatter(opts *Options) logrus.Formatter {
	return &customFormatter{
		TimestampFormat: time.RFC3339,     // 时间格式
		PrettyPrint:     opts.Development, // 是否美化打印
	}
}

// Format 格式化日志
func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(map[string]interface{})

	// 时间格式
	if f.TimestampFormat != "" {
		data["timestamp"] = entry.Time.Format(f.TimestampFormat) // 使用自定义时间格式
	} else {
		data["timestamp"] = entry.Time.Format(time.RFC3339) // 使用默认时间格式
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
		if customErr, ok := err.(types.Error); ok {
			data["error_code"] = customErr.Code()   // 错误代码
			data["error_stack"] = customErr.Stack() // 错误堆栈
		}
	}

	// 添加字段
	for k, v := range entry.Data {
		if k == "error" {
			if err, ok := v.(error); ok {
				data[k] = err.Error() // 错误信息
			} else {
				data[k] = v // 其他数据
			}
			continue
		}
		data[k] = v // 其他数据
	}

	if f.PrettyPrint {
		return json.MarshalIndent(data, "", "    ") // 美化打印
	}
	return json.Marshal(data) // 普通打印
}
