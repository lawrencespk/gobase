package metrics

import (
	"context"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
)

// Collector 全局限流器指标收集器实例
var Collector = NewRateLimitCollector()

func init() {
	// 获取默认logger并添加模块标识
	log := logger.GetLogger().WithFields(
		types.Field{Key: "module", Value: "ratelimit"},
		types.Field{Key: "component", Value: "metrics"},
	)

	// 注册限流器指标收集器
	if err := Collector.Register(); err != nil {
		log.Error(context.Background(), "failed to register rate limit metrics collector",
			types.Field{Key: "error", Value: err},
		)
		panic(err)
	}

	log.Info(context.Background(), "rate limit metrics collector registered successfully")
}
