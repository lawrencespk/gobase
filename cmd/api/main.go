package main

import (
	"context"
	"log"

	"gobase/pkg/config"
	"gobase/pkg/logger"
	loggerTypes "gobase/pkg/logger/types"
	grafanaConfig "gobase/pkg/monitor/grafana/config"
)

func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}
	cfg := config.GetConfig()

	// 初始化日志
	logger, err := logger.NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// 初始化 Grafana 配置管理器，转换配置类型
	grafanaCfg := grafanaConfig.NewManager(cfg.ToTypesConfig(), logger)

	// 监听配置变更
	ctx := context.Background()
	if err := grafanaCfg.WatchConfig(ctx); err != nil {
		logger.Error(ctx, "Failed to watch grafana config", loggerTypes.Error(err))
	}

	// ... 其他初始化代码 ...
}
