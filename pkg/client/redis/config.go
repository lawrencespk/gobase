package redis

import "time"

// Config Redis客户端配置
type Config struct {
	// 基础配置
	Addresses []string `yaml:"addresses"`
	Username  string   `yaml:"username"`
	Password  string   `yaml:"password"`
	Database  int      `yaml:"database"`

	// 连接池配置
	PoolSize     int `yaml:"pool_size"`
	MinIdleConns int `yaml:"min_idle_conns"`
	MaxRetries   int `yaml:"max_retries"`

	// 超时配置
	DialTimeout  time.Duration `yaml:"dial_timeout"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`

	// TLS配置
	EnableTLS   bool   `yaml:"enable_tls"`
	TLSCertFile string `yaml:"tls_cert_file"`
	TLSKeyFile  string `yaml:"tls_key_file"`

	// 集群配置
	EnableCluster bool `yaml:"enable_cluster"`
	RouteRandomly bool `yaml:"route_randomly"`

	// 监控配置
	EnableMetrics bool `yaml:"enable_metrics"`
	EnableTracing bool `yaml:"enable_tracing"`
}
