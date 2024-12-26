package core

// WithName 设置限流器名称
func WithName(name string) LimiterOption {
	return func(opts *LimiterOptions) {
		opts.Name = name
	}
}

// WithAlgorithm 设置限流算法
func WithAlgorithm(algorithm string) LimiterOption {
	return func(opts *LimiterOptions) {
		opts.Algorithm = algorithm
	}
}

// WithRedisConfig 设置Redis配置
func WithRedisConfig(config *RedisConfig) LimiterOption {
	return func(opts *LimiterOptions) {
		opts.RedisConfig = config
	}
}

// WithMetrics 设置是否启用监控
func WithMetrics(enabled bool) LimiterOption {
	return func(opts *LimiterOptions) {
		opts.EnableMetrics = enabled
	}
}

// WithMetricsPrefix 设置监控指标前缀
func WithMetricsPrefix(prefix string) LimiterOption {
	return func(opts *LimiterOptions) {
		opts.MetricsPrefix = prefix
	}
}

// WithTracing 设置是否启用链路追踪
func WithTracing(enabled bool) LimiterOption {
	return func(opts *LimiterOptions) {
		opts.EnableTracing = enabled
	}
}
