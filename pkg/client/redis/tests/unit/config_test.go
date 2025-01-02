package unit

import (
	"gobase/pkg/client/redis"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	// 启用测试模式，禁用连接检查和追踪
	redis.DisableConnectionCheck = true
	redis.DisableTracing = true
	defer func() {
		redis.DisableConnectionCheck = false
		redis.DisableTracing = false
	}()

	t.Run("valid config", func(t *testing.T) {
		cfg := &redis.Config{
			Addresses:     []string{"localhost:6379"},
			Username:      "user",
			Password:      "pass",
			Database:      0,
			PoolSize:      10,
			MinIdleConns:  2,
			MaxRetries:    3,
			DialTimeout:   time.Second,
			ReadTimeout:   time.Second,
			WriteTimeout:  time.Second,
			EnableTLS:     false,
			EnableCluster: false,
			EnableMetrics: true,
			EnableTracing: true,
		}

		client, err := redis.NewClientFromConfig(cfg)
		assert.NoError(t, err)
		if client != nil {
			defer client.Close()
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		cfg := &redis.Config{
			// 缺少必要的地址配置
			Database: -1, // 无效的数据库编号
		}

		client, err := redis.NewClientFromConfig(cfg)
		assert.Error(t, err)
		assert.Nil(t, client)
	})

	t.Run("cluster config", func(t *testing.T) {
		cfg := &redis.Config{
			Addresses:     []string{"localhost:6379", "localhost:6380"},
			EnableCluster: true,
			RouteRandomly: true,
		}

		client, err := redis.NewClientFromConfig(cfg)
		assert.NoError(t, err)
		if client != nil {
			defer client.Close()
		}
	})
}
