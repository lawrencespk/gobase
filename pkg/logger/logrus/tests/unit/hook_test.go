package unit

import (
	"bytes"
	"errors"
	"testing"

	locallogrus "gobase/pkg/logger/logrus"

	slogrus "github.com/sirupsen/logrus"
)

// TestNewHook 测试Hook的创建
func TestNewHook(t *testing.T) {
	writer := &bytes.Buffer{}
	formatter := &slogrus.TextFormatter{}
	levels := []slogrus.Level{slogrus.InfoLevel, slogrus.ErrorLevel}

	h := locallogrus.NewLogHook(writer, formatter, levels...)

	if h == nil {
		t.Error("期望创建的 Hook 不为 nil")
	}

	// 验证支持的日志级别
	hookLevels := h.Levels()
	if len(hookLevels) != len(levels) {
		t.Errorf("期望支持 %d 个日志级别, 实际支持 %d 个", len(levels), len(hookLevels))
	}

	for i, level := range levels {
		if hookLevels[i] != level {
			t.Errorf("期望第 %d 个级别为 %v, 实际为 %v", i, level, hookLevels[i])
		}
	}
}

// TestHookLevels 测试Hook的日志级别过滤
func TestHookLevels(t *testing.T) {
	writer := &bytes.Buffer{}
	formatter := &slogrus.TextFormatter{}

	tests := []struct {
		name   string
		levels []slogrus.Level
	}{
		{
			name:   "单个级别",
			levels: []slogrus.Level{slogrus.InfoLevel},
		},
		{
			name:   "多个级别",
			levels: []slogrus.Level{slogrus.InfoLevel, slogrus.WarnLevel, slogrus.ErrorLevel},
		},
		{
			name:   "所有级别",
			levels: slogrus.AllLevels,
		},
		{
			name:   "无级别",
			levels: []slogrus.Level{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := locallogrus.NewLogHook(writer, formatter, tt.levels...)
			levels := h.Levels()

			if len(levels) != len(tt.levels) {
				t.Errorf("期望 %d 个级别, 实际得到 %d 个", len(tt.levels), len(levels))
			}

			for i, level := range tt.levels {
				if levels[i] != level {
					t.Errorf("期望级别 %v, 实际得到 %v", level, levels[i])
				}
			}
		})
	}
}

// TestHookFire 测试Hook的Fire方法
func TestHookFire(t *testing.T) {
	writer := &bytes.Buffer{}
	formatter := &slogrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
	}

	h := locallogrus.NewLogHook(writer, formatter, slogrus.InfoLevel)

	entry := &slogrus.Entry{
		Logger:  slogrus.New(),
		Level:   slogrus.InfoLevel,
		Message: "test message",
	}

	err := h.Fire(entry)
	if err != nil {
		t.Errorf("期望 Fire 执行成功, 实际得到错误: %v", err)
	}

	if !bytes.Contains(writer.Bytes(), []byte("test message")) {
		t.Error("期望输出包含日志消息")
	}
}

// TestHookFireWithFields 测试带字段的Hook Fire
func TestHookFireWithFields(t *testing.T) {
	writer := &bytes.Buffer{}
	formatter := &slogrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
	}

	h := locallogrus.NewLogHook(writer, formatter, slogrus.InfoLevel)

	entry := &slogrus.Entry{
		Logger: slogrus.New(),
		Level:  slogrus.InfoLevel,
		Data: slogrus.Fields{
			"key": "value",
		},
		Message: "test message with fields",
	}

	err := h.Fire(entry)
	if err != nil {
		t.Errorf("期望 Fire 执行成功, 实际得到错误: %v", err)
	}

	if !bytes.Contains(writer.Bytes(), []byte("key=value")) {
		t.Error("期望输出包含字段信息")
	}
}

// TestHookFireError 测试Hook Fire的错误处理
func TestHookFireError(t *testing.T) {
	// 创建一个会产生错误的writer
	errorWriter := &errorWriter{}
	formatter := &slogrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
	}

	h := locallogrus.NewLogHook(errorWriter, formatter, slogrus.InfoLevel)

	entry := &slogrus.Entry{
		Logger:  slogrus.New(),
		Level:   slogrus.InfoLevel,
		Message: "test message",
	}

	err := h.Fire(entry)
	if err == nil {
		t.Error("期望得到一个错误，但是没有")
	}
}

// errorWriter 用于测试写入错误的情况
type errorWriter struct{}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("模拟的写入错误")
}
