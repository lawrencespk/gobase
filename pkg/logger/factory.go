package logger

import (
	"fmt"

	"gobase/pkg/logger/provider"
	"gobase/pkg/logger/types"
)

// Factory 日志工厂
type Factory struct{}

// NewLogger 创建新的日志实例
func (f *Factory) NewLogger(cfg types.Config) (types.Logger, error) {
	switch cfg.Type {
	case "logrus":
		return provider.NewLogrusLogger(cfg)
	case "elk":
		return provider.NewElkLogger(cfg)
	default:
		return nil, fmt.Errorf("unsupported logger type: %s", cfg.Type)
	}
}
