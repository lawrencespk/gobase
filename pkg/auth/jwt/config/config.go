package config

import (
	"time"

	"gobase/pkg/auth/jwt"
)

// Config JWT配置
type Config struct {
	// 签名方法
	SigningMethod jwt.SigningMethod `json:"signing_method" yaml:"signing_method"`

	// 密钥配置
	SecretKey  string `json:"secret_key" yaml:"secret_key"`
	PublicKey  string `json:"public_key" yaml:"public_key"`
	PrivateKey string `json:"private_key" yaml:"private_key"`

	// Token配置
	AccessTokenExpiration  time.Duration `json:"access_token_expiration" yaml:"access_token_expiration"`
	RefreshTokenExpiration time.Duration `json:"refresh_token_expiration" yaml:"refresh_token_expiration"`

	// 黑名单配置
	BlacklistEnabled bool   `json:"blacklist_enabled" yaml:"blacklist_enabled"`
	BlacklistType    string `json:"blacklist_type" yaml:"blacklist_type"` // memory or redis

	// Redis配置 (当BlacklistType为redis时使用)
	Redis *RedisConfig `json:"redis,omitempty" yaml:"redis,omitempty"`

	// 安全配置
	EnableRotation   bool          `json:"enable_rotation" yaml:"enable_rotation"`
	RotationInterval time.Duration `json:"rotation_interval" yaml:"rotation_interval"`

	// 会话配置
	EnableSession     bool `json:"enable_session" yaml:"enable_session"`
	MaxActiveSessions int  `json:"max_active_sessions" yaml:"max_active_sessions"`

	// 绑定配置
	EnableIPBinding     bool `json:"enable_ip_binding" yaml:"enable_ip_binding"`
	EnableDeviceBinding bool `json:"enable_device_binding" yaml:"enable_device_binding"`

	// 监控配置
	EnableMetrics bool `json:"enable_metrics" yaml:"enable_metrics"`

	// 链路追踪配置
	EnableTracing bool `json:"enable_tracing" yaml:"enable_tracing"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string `json:"addr" yaml:"addr"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		SigningMethod:          jwt.HS256,
		AccessTokenExpiration:  2 * time.Hour,
		RefreshTokenExpiration: 24 * time.Hour,
		BlacklistEnabled:       true,
		BlacklistType:          "memory",
		EnableRotation:         false,
		RotationInterval:       24 * time.Hour,
		EnableSession:          false,
		MaxActiveSessions:      1,
		EnableIPBinding:        false,
		EnableDeviceBinding:    false,
		EnableMetrics:          true,
		EnableTracing:          true,
	}
}
