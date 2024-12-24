package types

import "time"

// Config Redis配置结构
type Config struct {
	// 基础配置
	Addresses []string `json:"addresses" yaml:"addresses"` // Redis地址列表
	Username  string   `json:"username" yaml:"username"`   // 用户名
	Password  string   `json:"password" yaml:"password"`   // 密码
	Database  int      `json:"database" yaml:"database"`   // 数据库索引

	// 连接池配置
	PoolSize     int `json:"poolSize" yaml:"poolSize"`         // 连接池大小
	MinIdleConns int `json:"minIdleConns" yaml:"minIdleConns"` // 最小空闲连接数
	MaxRetries   int `json:"maxRetries" yaml:"maxRetries"`     // 最大重试次数

	// 超时配置
	DialTimeout  time.Duration `json:"dialTimeout" yaml:"dialTimeout"`   // 连接超时
	ReadTimeout  time.Duration `json:"readTimeout" yaml:"readTimeout"`   // 读取超时
	WriteTimeout time.Duration `json:"writeTimeout" yaml:"writeTimeout"` // 写入超时

	// TLS配置
	EnableTLS   bool   `json:"enableTLS" yaml:"enableTLS"`     // 是否启用TLS
	TLSCertFile string `json:"tlsCertFile" yaml:"tlsCertFile"` // TLS证书文件
	TLSKeyFile  string `json:"tlsKeyFile" yaml:"tlsKeyFile"`   // TLS密钥文件

	// 集群配置
	EnableCluster bool `json:"enableCluster" yaml:"enableCluster"` // 是否启用集群
	RouteRandomly bool `json:"routeRandomly" yaml:"routeRandomly"` // 是否随机路由

	// 监控配置
	EnableMetrics bool `json:"enableMetrics" yaml:"enableMetrics"` // 是否启用监控
	EnableTracing bool `json:"enableTracing" yaml:"enableTracing"` // 是否启用链路追踪
}
