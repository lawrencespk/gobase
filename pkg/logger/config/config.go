package config

import (
	"time"

	"gobase/pkg/logger/types"
)

// LoggerConfig 日志配置管理器
type LoggerConfig struct {
	// 内部配置存储
	configs map[string]types.Config
}

// NewLoggerConfig 创建新的日志配置管理器
func NewLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		configs: make(map[string]types.Config),
	}
}

// AddLogrusConfig 添加 Logrus 配置
func (c *LoggerConfig) AddLogrusConfig(name string, level types.Level, format string, defaultFields types.Fields) {
	c.configs[name] = types.Config{
		Type:          "logrus",
		Level:         level,
		Format:        format,
		TimeFormat:    time.RFC3339,
		Caller:        true,
		DefaultFields: defaultFields,
	}
}

// AddELKConfig 添加 ELK 配置
func (c *LoggerConfig) AddELKConfig(name string, level types.Level, endpoint, index string, defaultFields types.Fields) {
	c.configs[name] = types.Config{
		Type:          "elk",
		Level:         level,
		TimeFormat:    time.RFC3339,
		ElkEndpoint:   endpoint,
		ElkIndex:      index,
		ElkType:       "_doc",
		DefaultFields: defaultFields,
	}
}

// GetConfig 获取指定名称的配置
func (c *LoggerConfig) GetConfig(name string) (types.Config, bool) {
	cfg, ok := c.configs[name]
	return cfg, ok
}
