package jaeger

import (
	"errors"

	"gobase/pkg/config"
	"gobase/pkg/config/types"
)

// Config 别名，保持兼容性
type Config = types.JaegerConfig
type AgentConfig = types.JaegerAgentConfig
type CollectorConfig = types.JaegerCollectorConfig
type SamplerConfig = types.JaegerSamplerConfig
type BufferConfig = types.JaegerBufferConfig

// LoadConfig 从全局配置中加载 Jaeger 配置
func LoadConfig() (*Config, error) {
	globalConfig := config.GetConfig()
	if globalConfig == nil {
		return nil, errors.New("global config not initialized")
	}

	cfg := &globalConfig.Jaeger
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ValidateConfig 验证配置
func ValidateConfig(cfg *Config) error {
	if cfg.ServiceName == "" {
		return errors.New("service name is required")
	}

	if cfg.Enable {
		// 验证Agent配置
		if cfg.Agent.Host == "" {
			return errors.New("agent host is required")
		}
		if cfg.Agent.Port == "" {
			return errors.New("agent port is required")
		}

		// 验证Collector配置
		if cfg.Collector.Endpoint == "" {
			return errors.New("collector endpoint is required")
		}
		if cfg.Collector.Timeout <= 0 {
			return errors.New("collector timeout must be positive")
		}

		// 验证采样配置
		switch cfg.Sampler.Type {
		case "const", "probabilistic", "rateLimiting", "remote":
			// 有效的采样类型
		default:
			return errors.New("invalid sampler type")
		}

		if cfg.Sampler.Type == "probabilistic" && (cfg.Sampler.Param < 0 || cfg.Sampler.Param > 1) {
			return errors.New("probabilistic sampler param must be between 0 and 1")
		}

		if cfg.Sampler.Type == "remote" && cfg.Sampler.ServerURL == "" {
			return errors.New("remote sampler server URL is required")
		}

		// 验证缓冲区配置
		if cfg.Buffer.Enable {
			if cfg.Buffer.Size <= 0 {
				return errors.New("buffer size must be positive")
			}
			if cfg.Buffer.FlushInterval <= 0 {
				return errors.New("buffer flush interval must be positive")
			}
		}
	}

	return nil
}
