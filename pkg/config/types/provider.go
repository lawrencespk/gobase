package types

import (
	"context"
	"time"

	"gobase/pkg/logger/types" // 引入日志接口
)

// Provider 定义配置提供者接口
type Provider interface {
	// 基础操作
	Load(ctx context.Context) error  // 使用context便于传递上下文信息
	Close(ctx context.Context) error // 关闭时也传入context
	IsLoaded() bool

	// 配置读取 (出错时返回error而不是默认值,便于错误处理)
	Get(key string) (interface{}, error)
	GetString(key string) (string, error)
	GetInt(key string) (int, error)
	GetInt64(key string) (int64, error)
	GetFloat64(key string) (float64, error)
	GetBool(key string) (bool, error)
	GetTime(key string) (time.Time, error)
	GetDuration(key string) (time.Duration, error)
	GetStringSlice(key string) ([]string, error)
	GetStringMap(key string) (map[string]interface{}, error)
	GetStringMapString(key string) (map[string]string, error)

	// 配置解析
	Unmarshal(val interface{}) error
	UnmarshalKey(key string, val interface{}) error

	// 配置监听
	OnConfigChange(func(Event))
	WatchConfig(ctx context.Context) error
	StopWatchConfig(ctx context.Context) error

	// 设置日志记录器
	SetLogger(logger types.Logger)
}

// Event 定义配置变更事件
type Event struct {
	Provider  string
	Key       string
	Value     interface{}
	OldValue  interface{}
	Type      EventType
	Timestamp time.Time
	Error     error // 添加错误字段,用于传递变更过程中的错误
}

// EventType 定义事件类型
type EventType uint8

const (
	EventCreate EventType = iota
	EventUpdate
	EventDelete
)

// Options 定义配置选项
type Options struct {
	Name          string
	AutoLoad      bool
	AutoWatch     bool
	RetryInterval time.Duration
	RetryTimes    int
	WatchInterval time.Duration
	EnableCache   bool
	CacheDuration time.Duration
	Logger        types.Logger // 添加日志记录器选项
}
