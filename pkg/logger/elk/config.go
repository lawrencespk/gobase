package elk

import (
	"crypto/tls"
	"time"
)

// ElasticConfig Elasticsearch配置
type ElasticConfig struct {
	// 基础配置
	Addresses   []string // Elasticsearch地址
	Username    string   // 用户名
	Password    string   // 密码
	IndexPrefix string   // 索引前缀

	// 连接配置
	Sniff       bool          // 是否启用嗅探
	Healthcheck bool          // 是否启用健康检查
	RetryOnFail bool          // 是否在失败时重试
	MaxRetries  int           // 最大重试次数
	DialTimeout time.Duration // 连接超时时间
	TLS         *tls.Config   // TLS配置

	// 索引配置
	NumberOfShards   int    // 分片数
	NumberOfReplicas int    // 副本数
	RefreshInterval  string // 刷新间隔

	// 批处理配置
	BatchSize     int           // 批量大小
	FlushInterval time.Duration // 刷新间隔
	Workers       int           // 工作线程数
}

// DefaultElasticConfig 返回默认配置
func DefaultElasticConfig() *ElasticConfig {
	return &ElasticConfig{
		Addresses:        []string{"http://104.238.161.243:9200"}, // 地址
		IndexPrefix:      "logs",                                  // 索引前缀
		Sniff:            true,                                    // 是否启用嗅探
		Healthcheck:      true,                                    // 是否启用健康检查
		RetryOnFail:      true,                                    // 是否在失败时重试
		MaxRetries:       3,                                       // 最大重试次数
		DialTimeout:      5 * time.Second,                         // 连接超时时间
		NumberOfShards:   5,                                       // 分片数
		NumberOfReplicas: 1,                                       // 副本数
		RefreshInterval:  "5s",                                    // 刷新间隔
		BatchSize:        1000,                                    // 批量大小
		FlushInterval:    5 * time.Second,                         // 刷新间隔
		Workers:          2,                                       // 工作线程数
	}
}
