package integration_test

import (
	"testing"

	"gobase/pkg/auth/jwt/config"

	"github.com/stretchr/testify/assert"
)

func TestConfigWithRedisBlacklist(t *testing.T) {
	cfg := config.DefaultConfig()

	// 配置 Redis 黑名单
	config.WithBlacklist(true, "redis")(cfg)
	config.WithRedis("localhost:6379", "", 0)(cfg)

	// 验证配置是否正确集成
	assert.True(t, cfg.BlacklistEnabled)
	assert.Equal(t, "redis", cfg.BlacklistType)
	assert.NotNil(t, cfg.Redis)
	assert.Equal(t, "localhost:6379", cfg.Redis.Addr)
}
