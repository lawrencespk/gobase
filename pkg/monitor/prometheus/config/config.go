package config

import (
	"fmt"
	"gobase/pkg/errors"
	"gobase/pkg/monitor/prometheus/config/types"
)

// Validate 验证配置
func Validate(cfg *types.Config) error {
	if cfg == nil {
		return errors.NewConfigError("prometheus config is nil", nil)
	}

	if cfg.Enabled {
		if cfg.Port <= 0 || cfg.Port > 65535 {
			return errors.NewConfigError(
				fmt.Sprintf("invalid prometheus port: %d", cfg.Port),
				nil,
			)
		}

		if cfg.Path == "" {
			return errors.NewConfigError("prometheus metrics path cannot be empty", nil)
		}

		if cfg.Sampling.Enabled && (cfg.Sampling.Rate <= 0 || cfg.Sampling.Rate > 1) {
			return errors.NewConfigError(
				fmt.Sprintf("invalid sampling rate: %f", cfg.Sampling.Rate),
				nil,
			)
		}
	}

	return nil
}
