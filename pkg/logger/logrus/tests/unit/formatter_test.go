package unit

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	syslogrus "github.com/sirupsen/logrus"
)

// customError 是用于测试的自定义错误类型
type customError struct {
	code    string
	message string
	stack   []string
}

func (e *customError) Error() string   { return e.message }
func (e *customError) Code() string    { return e.code }
func (e *customError) Stack() []string { return e.stack }

// customFormatter 是测试用的格式化器
type customFormatter struct {
	TimestampFormat string // 时间格式
	PrettyPrint     bool   // 是否美化打印
}

func (f *customFormatter) Format(entry *syslogrus.Entry) ([]byte, error) {
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
			"function": entry.Caller.Function,
			"file":     entry.Caller.File,
			"line":     entry.Caller.Line,
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

func TestCustomFormatter(t *testing.T) {
	// 基础格式化测试
	t.Run("Basic Formatting", func(t *testing.T) {
		formatter := &customFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		}

		entry := &syslogrus.Entry{
			Time:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			Level:   syslogrus.InfoLevel,
			Message: "test message",
			Data: syslogrus.Fields{
				"string_key": "string_value",
				"int_key":    123,
				"bool_key":   true,
			},
		}

		output, err := formatter.Format(entry)
		if err != nil {
			t.Fatalf("Failed to format entry: %v", err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal(output, &data); err != nil {
			t.Fatalf("Failed to parse formatted output: %v", err)
		}

		// 验证基本字段
		expectedTime := "2024-01-01T12:00:00Z"
		if data["timestamp"] != expectedTime {
			t.Errorf("Expected timestamp %v, got %v", expectedTime, data["timestamp"])
		}
		if data["level"] != "info" {
			t.Errorf("Expected level info, got %v", data["level"])
		}
		if data["message"] != "test message" {
			t.Errorf("Expected message 'test message', got %v", data["message"])
		}
	})

	// 错误格式化测试
	t.Run("Error Formatting", func(t *testing.T) {
		formatter := &customFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     true,
		}

		customErr := &customError{
			code:    "ERR001",
			message: "custom error message",
			stack:   []string{"func1", "func2"},
		}

		entry := &syslogrus.Entry{
			Time:    time.Now(),
			Level:   syslogrus.ErrorLevel,
			Message: "error occurred",
			Data: syslogrus.Fields{
				"error": customErr,
			},
		}

		output, err := formatter.Format(entry)
		if err != nil {
			t.Fatalf("Failed to format error entry: %v", err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal(output, &data); err != nil {
			t.Fatalf("Failed to parse formatted output: %v", err)
		}

		// 验证错误相关字段
		if data["error"] != "custom error message" {
			t.Errorf("Expected error message 'custom error message', got %v", data["error"])
		}
		if data["error_code"] != "ERR001" {
			t.Errorf("Expected error code 'ERR001', got %v", data["error_code"])
		}
	})

	// 开发模式测试
	t.Run("Development Mode", func(t *testing.T) {
		formatter := &customFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     true,
		}

		entry := &syslogrus.Entry{
			Time:    time.Now(),
			Level:   syslogrus.DebugLevel,
			Message: "debug message",
			Data: syslogrus.Fields{
				"key": "value",
			},
		}

		output, err := formatter.Format(entry)
		if err != nil {
			t.Fatalf("Failed to format entry in development mode: %v", err)
		}

		// 验证输出是否为格式化的 JSON
		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, output, "", "    ")
		if err != nil {
			t.Fatalf("Output is not valid JSON: %v", err)
		}

		// 验证输出是否包含预期的缩进
		if !strings.Contains(string(output), "\n    ") {
			t.Error("Expected formatted output to contain indentation in development mode")
		}
	})

	// 标准错误测试
	t.Run("Standard Error", func(t *testing.T) {
		formatter := &customFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		}

		stdErr := errors.New("standard error")
		entry := &syslogrus.Entry{
			Time:    time.Now(),
			Level:   syslogrus.ErrorLevel,
			Message: "error message",
			Data: syslogrus.Fields{
				"error": stdErr,
			},
		}

		output, err := formatter.Format(entry)
		if err != nil {
			t.Fatalf("Failed to format standard error entry: %v", err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal(output, &data); err != nil {
			t.Fatalf("Failed to parse formatted output: %v", err)
		}

		if data["error"] != "standard error" {
			t.Errorf("Expected error 'standard error', got %v", data["error"])
		}
	})

	// 空字段测试
	t.Run("Empty Fields", func(t *testing.T) {
		formatter := &customFormatter{
			TimestampFormat: time.RFC3339,
			PrettyPrint:     false,
		}

		entry := &syslogrus.Entry{
			Time:    time.Now(),
			Level:   syslogrus.InfoLevel,
			Message: "",
			Data:    syslogrus.Fields{},
		}

		output, err := formatter.Format(entry)
		if err != nil {
			t.Fatalf("Failed to format empty entry: %v", err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal(output, &data); err != nil {
			t.Fatalf("Failed to parse formatted output: %v", err)
		}

		if data["message"] != "" {
			t.Errorf("Expected empty message, got %v", data["message"])
		}
	})
}

// 性能测试
func BenchmarkFormatter(b *testing.B) {
	formatter := &customFormatter{
		TimestampFormat: time.RFC3339,
		PrettyPrint:     false,
	}

	entry := &syslogrus.Entry{
		Time:    time.Now(),
		Level:   syslogrus.InfoLevel,
		Message: "benchmark message",
		Data: syslogrus.Fields{
			"string_key": "string_value",
			"int_key":    123,
			"bool_key":   true,
			"error":      errors.New("test error"),
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			formatter.Format(entry)
		}
	})
}
