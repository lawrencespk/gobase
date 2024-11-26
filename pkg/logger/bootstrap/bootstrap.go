package bootstrap // 改为 bootstrap 包名

import (
	"fmt"

	"gobase/pkg/logger/config"
	"gobase/pkg/logger/manager"
	"gobase/pkg/logger/types"
)

var (
	defaultConfig  *config.LoggerConfig
	defaultManager *manager.LoggerManager
)

// Initialize 初始化日志系统
func Initialize() { // 改名为 Initialize
	defaultConfig = config.NewLoggerConfig()
	defaultManager = manager.NewLoggerManager()

	// 添加默认配置
	defaultConfig.AddLogrusConfig(
		"default",
		types.InfoLevel,
		"json",
		types.Fields{"service": "myapp"},
	)

	defaultConfig.AddELKConfig(
		"elk",
		types.InfoLevel,
		"http://localhost:9200",
		"myapp-logs",
		types.Fields{"service": "myapp"},
	)
}

// GetLogger 获取日志实例
func GetLogger(name string) (types.Logger, error) {
	cfg, ok := defaultConfig.GetConfig(name)
	if !ok {
		return nil, fmt.Errorf("logger config not found: %s", name)
	}
	return defaultManager.GetOrCreate(name, cfg)
}
