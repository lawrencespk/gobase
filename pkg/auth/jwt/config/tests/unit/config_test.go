package config_test

import (
	"testing"
	"time"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/config"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()

	assert.Equal(t, jwt.HS256, cfg.SigningMethod)
	assert.Equal(t, 2*time.Hour, cfg.AccessTokenExpiration)
	assert.Equal(t, 24*time.Hour, cfg.RefreshTokenExpiration)
	assert.True(t, cfg.BlacklistEnabled)
	assert.Equal(t, "memory", cfg.BlacklistType)
	// ... 验证其他默认值
}

func TestConfigOptions(t *testing.T) {
	tests := []struct {
		name     string
		options  []config.Option
		validate func(*testing.T, *config.Config)
	}{
		{
			name: "WithSigningMethod",
			options: []config.Option{
				config.WithSigningMethod(jwt.RS256),
			},
			validate: func(t *testing.T, c *config.Config) {
				assert.Equal(t, jwt.RS256, c.SigningMethod)
			},
		},
		{
			name: "WithKeyPair",
			options: []config.Option{
				config.WithKeyPair("public-key", "private-key"),
			},
			validate: func(t *testing.T, c *config.Config) {
				assert.Equal(t, "public-key", c.PublicKey)
				assert.Equal(t, "private-key", c.PrivateKey)
			},
		},
		// ... 测试其他配置选项
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.DefaultConfig()
			for _, opt := range tt.options {
				opt(cfg)
			}
			tt.validate(t, cfg)
		})
	}
}
