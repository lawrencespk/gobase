package types

import "time"

// Loader 配置加载器接口
type Loader interface {
	Load() error
	Get(key string) interface{}
	GetString(key string) string
	GetInt(key string) int
	GetBool(key string) bool
	GetFloat64(key string) float64
	GetDuration(key string) time.Duration
	GetStringSlice(key string) []string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	IsSet(key string) bool
	Set(key string, value interface{})
	Watch(key string, callback func(key string, value interface{})) error
}

// Parser 配置解析器接口
type Parser interface {
	Parse(key string, out interface{}) error
}
