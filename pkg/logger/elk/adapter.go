// Package elk 提供 ELK 日志适配器实现
package elk

import (
	"context"
	"fmt"
	"os"
	"time"

	"gobase/pkg/logger/types"

	"github.com/olivere/elastic/v7"
)

// Adapter ELK适配器，实现 types.Logger 接口
type Adapter struct {
	// elastic client
	client *elastic.Client
	// 日志配置
	config types.Config
	// 当前字段
	fields types.Fields
	// 上下文
	ctx context.Context
}

// NewAdapter 创建新的 ELK 适配器
func NewAdapter(cfg types.Config) (types.Logger, error) {
	// 创建 elasticsearch 客户端
	client, err := elastic.NewClient(
		elastic.SetURL(cfg.ElkEndpoint),
		elastic.SetSniff(false),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	return &Adapter{
		client: client,
		config: cfg,
		fields: cfg.DefaultFields,
		ctx:    context.Background(),
	}, nil
}

// createLogEntry 创建日志条目
func (a *Adapter) createLogEntry(level string, message interface{}) map[string]interface{} {
	// 基础日志字段
	entry := map[string]interface{}{
		"timestamp": time.Now().Format(a.config.TimeFormat),
		"level":     level,
		"message":   fmt.Sprint(message),
	}

	// 添加默认字段
	for k, v := range a.fields {
		entry[k] = v
	}

	return entry
}

// sendToElk 发送日志到 ELK
func (a *Adapter) sendToElk(entry map[string]interface{}) error {
	_, err := a.client.Index().
		Index(a.config.ElkIndex).
		Type(a.config.ElkType).
		BodyJson(entry).
		Do(a.ctx)
	return err
}

// shouldLog 判断是否应该记录该级别的日志
func (a *Adapter) shouldLog(level types.Level) bool {
	levels := map[types.Level]int{
		types.DebugLevel: 0,
		types.InfoLevel:  1,
		types.WarnLevel:  2,
		types.ErrorLevel: 3,
		types.FatalLevel: 4,
		types.PanicLevel: 5,
	}

	return levels[level] >= levels[a.config.Level]
}

// Debug 实现 Debug 级别日志
func (a *Adapter) Debug(args ...interface{}) {
	if a.shouldLog(types.DebugLevel) {
		entry := a.createLogEntry("debug", fmt.Sprint(args...))
		if err := a.sendToElk(entry); err != nil {
			fmt.Printf("Failed to send log to ELK: %v\n", err)
		}
	}
}

// Debugf 实现格式化的 Debug 级别日志
func (a *Adapter) Debugf(format string, args ...interface{}) {
	if a.shouldLog(types.DebugLevel) {
		entry := a.createLogEntry("debug", fmt.Sprintf(format, args...))
		if err := a.sendToElk(entry); err != nil {
			fmt.Printf("Failed to send log to ELK: %v\n", err)
		}
	}
}

// Info 实现 Info 级别日志
func (a *Adapter) Info(args ...interface{}) {
	if a.shouldLog(types.InfoLevel) {
		entry := a.createLogEntry("info", fmt.Sprint(args...))
		if err := a.sendToElk(entry); err != nil {
			fmt.Printf("Failed to send log to ELK: %v\n", err)
		}
	}
}

// Infof 实现格式化的 Info 级别日志
func (a *Adapter) Infof(format string, args ...interface{}) {
	if a.shouldLog(types.InfoLevel) {
		entry := a.createLogEntry("info", fmt.Sprintf(format, args...))
		if err := a.sendToElk(entry); err != nil {
			fmt.Printf("Failed to send log to ELK: %v\n", err)
		}
	}
}

// Warn 实现 Warn 级别日志
func (a *Adapter) Warn(args ...interface{}) {
	if a.shouldLog(types.WarnLevel) {
		entry := a.createLogEntry("warn", fmt.Sprint(args...))
		if err := a.sendToElk(entry); err != nil {
			fmt.Printf("Failed to send log to ELK: %v\n", err)
		}
	}
}

// Warnf 实现格式化的 Warn 级别日志
func (a *Adapter) Warnf(format string, args ...interface{}) {
	if a.shouldLog(types.WarnLevel) {
		entry := a.createLogEntry("warn", fmt.Sprintf(format, args...))
		if err := a.sendToElk(entry); err != nil {
			fmt.Printf("Failed to send log to ELK: %v\n", err)
		}
	}
}

// Error 实现 Error 级别日志
func (a *Adapter) Error(args ...interface{}) {
	if a.shouldLog(types.ErrorLevel) {
		entry := a.createLogEntry("error", fmt.Sprint(args...))
		if err := a.sendToElk(entry); err != nil {
			fmt.Printf("Failed to send log to ELK: %v\n", err)
		}
	}
}

// Errorf 实现格式化的 Error 级别日志
func (a *Adapter) Errorf(format string, args ...interface{}) {
	if a.shouldLog(types.ErrorLevel) {
		entry := a.createLogEntry("error", fmt.Sprintf(format, args...))
		if err := a.sendToElk(entry); err != nil {
			fmt.Printf("Failed to send log to ELK: %v\n", err)
		}
	}
}

// Fatal 实现 Fatal 级别日志
func (a *Adapter) Fatal(args ...interface{}) {
	if a.shouldLog(types.FatalLevel) {
		entry := a.createLogEntry("fatal", fmt.Sprint(args...))
		if err := a.sendToElk(entry); err != nil {
			fmt.Printf("Failed to send log to ELK: %v\n", err)
		}
		os.Exit(1)
	}
}

// Fatalf 实现格式化的 Fatal 级别日志
func (a *Adapter) Fatalf(format string, args ...interface{}) {
	if a.shouldLog(types.FatalLevel) {
		entry := a.createLogEntry("fatal", fmt.Sprintf(format, args...))
		if err := a.sendToElk(entry); err != nil {
			fmt.Printf("Failed to send log to ELK: %v\n", err)
		}
		os.Exit(1)
	}
}

// Panic 实现 Panic 级别日志
func (a *Adapter) Panic(args ...interface{}) {
	if a.shouldLog(types.PanicLevel) {
		entry := a.createLogEntry("panic", fmt.Sprint(args...))
		if err := a.sendToElk(entry); err != nil {
			fmt.Printf("Failed to send log to ELK: %v\n", err)
		}
		panic(fmt.Sprint(args...))
	}
}

// Panicf 实现格式化的 Panic 级别日志
func (a *Adapter) Panicf(format string, args ...interface{}) {
	if a.shouldLog(types.PanicLevel) {
		entry := a.createLogEntry("panic", fmt.Sprintf(format, args...))
		if err := a.sendToElk(entry); err != nil {
			fmt.Printf("Failed to send log to ELK: %v\n", err)
		}
		panic(fmt.Sprintf(format, args...))
	}
}

// WithFields 实现字段追加
func (a *Adapter) WithFields(fields types.Fields) types.Logger {
	newAdapter := *a
	newFields := make(types.Fields)

	// 复制现有字段
	for k, v := range a.fields {
		newFields[k] = v
	}
	// 添加新字段
	for k, v := range fields {
		newFields[k] = v
	}

	newAdapter.fields = newFields
	return &newAdapter
}

// WithContext 实现上下文追加
func (a *Adapter) WithContext(ctx context.Context) types.Logger {
	newAdapter := *a
	newAdapter.ctx = ctx
	return &newAdapter
}

// WithError 实现错误信息追加
func (a *Adapter) WithError(err error) types.Logger {
	return a.WithFields(types.Fields{
		"error": err.Error(),
	})
}

// WithTime 实现时间信息追加
func (a *Adapter) WithTime(t time.Time) types.Logger {
	return a.WithFields(types.Fields{
		"custom_time": t.Format(time.RFC3339),
	})
}

// AddHook ELK 适配器不支持 hook
func (a *Adapter) AddHook(hook interface{}) error {
	return fmt.Errorf("elk adapter doesn't support hooks")
}
