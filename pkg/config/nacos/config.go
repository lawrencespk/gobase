package nacos

import (
	"time"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
)

// Config Nacos配置
type Config struct {
	Endpoint    string        // Nacos服务端点
	NamespaceID string        // 命名空间ID
	Group       string        // 配置分组
	DataID      string        // 配置ID
	Username    string        // 用户名
	Password    string        // 密码
	Timeout     time.Duration // 超时时间
	LogDir      string        // 日志目录
	CacheDir    string        // 缓存目录
	LogLevel    string        // 日志级别
	// 认证相关
	AccessKey   string // AccessKey
	SecretKey   string // SecretKey
	EnableAuth  bool   // 是否启用认证
	AuthToken   string // 认证Token
	IdentityKey string // 身份Key
	IdentityVal string // 身份Value
}

// ToClientConfig 转换为Nacos客户端配置
func (c *Config) ToClientConfig() constant.ClientConfig {
	return constant.ClientConfig{
		NamespaceId:         c.NamespaceID,
		TimeoutMs:           uint64(c.Timeout / time.Millisecond),
		LogDir:              c.LogDir,
		CacheDir:            c.CacheDir,
		LogLevel:            c.LogLevel,
		Username:            c.Username,
		Password:            c.Password,
		AccessKey:           c.AccessKey,
		SecretKey:           c.SecretKey,
		OpenKMS:             c.EnableAuth,
		NotLoadCacheAtStart: true,
	}
}

// ToServerConfig 转换为Nacos服务端配置
func (c *Config) ToServerConfig() []constant.ServerConfig {
	return []constant.ServerConfig{
		{
			Scheme:      "http",
			ContextPath: "/nacos",
			IpAddr:      c.Endpoint,
			Port:        8848,
		},
	}
}
