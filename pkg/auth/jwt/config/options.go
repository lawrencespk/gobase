package config

import (
	"gobase/pkg/auth/jwt"
	"time"
)

// Option 配置选项函数
type Option func(*Config)

// WithSigningMethod 设置签名方法
func WithSigningMethod(method jwt.SigningMethod) Option {
	return func(c *Config) {
		c.SigningMethod = method
	}
}

// WithSecretKey 设置密钥
func WithSecretKey(key string) Option {
	return func(c *Config) {
		c.SecretKey = key
	}
}

// WithKeyPair 设置密钥对
func WithKeyPair(publicKey, privateKey string) Option {
	return func(c *Config) {
		c.PublicKey = publicKey
		c.PrivateKey = privateKey
	}
}

// WithAccessTokenExpiration 设置访问令牌过期时间
func WithAccessTokenExpiration(d time.Duration) Option {
	return func(c *Config) {
		c.AccessTokenExpiration = d
	}
}

// WithRefreshTokenExpiration 设置刷新令牌过期时间
func WithRefreshTokenExpiration(d time.Duration) Option {
	return func(c *Config) {
		c.RefreshTokenExpiration = d
	}
}

// WithBlacklist 配置黑名单
func WithBlacklist(enabled bool, typ string) Option {
	return func(c *Config) {
		c.BlacklistEnabled = enabled
		c.BlacklistType = typ
	}
}

// WithRedis 配置Redis
func WithRedis(addr, password string, db int) Option {
	return func(c *Config) {
		c.Redis = &RedisConfig{
			Addr:     addr,
			Password: password,
			DB:       db,
		}
	}
}

// WithRotation 配置Token轮换
func WithRotation(enabled bool, interval time.Duration) Option {
	return func(c *Config) {
		c.EnableRotation = enabled
		c.RotationInterval = interval
	}
}

// WithSession 配置会话管理
func WithSession(enabled bool, maxActive int) Option {
	return func(c *Config) {
		c.EnableSession = enabled
		c.MaxActiveSessions = maxActive
	}
}

// WithBinding 配置绑定
func WithBinding(enableIP, enableDevice bool) Option {
	return func(c *Config) {
		c.EnableIPBinding = enableIP
		c.EnableDeviceBinding = enableDevice
	}
}

// WithObservability 配置可观测性
func WithObservability(enableMetrics, enableTracing bool) Option {
	return func(c *Config) {
		c.EnableMetrics = enableMetrics
		c.EnableTracing = enableTracing
	}
}
