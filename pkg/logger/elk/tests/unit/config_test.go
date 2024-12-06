package unit

import (
	"testing"
	"time"

	"gobase/pkg/config"
	"gobase/pkg/logger/elk"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestElkConfig(t *testing.T) {
	// 确保在测试开始时设置一个空的配置
	config.SetConfig(nil)

	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := elk.DefaultElkConfig()
		assert.NotNil(t, cfg)
		assert.Equal(t, []string{"http://localhost:9200"}, cfg.Addresses)
		assert.Equal(t, "", cfg.Username)
		assert.Equal(t, "", cfg.Password)
		assert.Equal(t, "default-index", cfg.Index)
		assert.Equal(t, 30*time.Second, cfg.Timeout)
	})

	t.Run("LoadFromConfig", func(t *testing.T) {
		// 模拟配置
		testConfig := &config.Config{
			ELK: config.ELKConfig{
				Addresses: []string{"http://test-elk:9200"},
				Username:  "testuser",
				Password:  "testpass",
				Index:     "test-index",
				Timeout:   60,
			},
		}

		// 设置配置
		config.SetConfig(testConfig)
		defer config.SetConfig(nil) // 清理

		cfg := elk.DefaultElkConfig()
		assert.NotNil(t, cfg)
		assert.Equal(t, []string{"http://test-elk:9200"}, cfg.Addresses)
		assert.Equal(t, "testuser", cfg.Username)
		assert.Equal(t, "testpass", cfg.Password)
		assert.Equal(t, "test-index", cfg.Index)
		assert.Equal(t, 60*time.Second, cfg.Timeout)
	})

	t.Run("ValidateConfig", func(t *testing.T) {
		client := elk.NewElkClient()

		t.Run("ValidConfig", func(t *testing.T) {
			cfg := &elk.ElkConfig{
				Addresses: []string{"http://localhost:9200"},
				Timeout:   time.Duration(30),
			}
			err := client.Connect(cfg)
			assert.NoError(t, err)
		})

		t.Run("EmptyAddresses", func(t *testing.T) {
			cfg := &elk.ElkConfig{
				Addresses: []string{},
				Timeout:   time.Duration(30),
			}
			err := client.Connect(cfg)
			require.Error(t, err, "应该返回错误，因为地址列表为空")
			if err != nil {
				assert.Contains(t, err.Error(), "no elasticsearch addresses provided")
			}
		})

		t.Run("InvalidTimeout", func(t *testing.T) {
			cfg := &elk.ElkConfig{
				Addresses: []string{"http://localhost:9200"},
				Timeout:   time.Duration(0),
			}
			err := client.Connect(cfg)
			require.Error(t, err, "应该返回错误，因为超时值无效")
			if err != nil {
				assert.Contains(t, err.Error(), "invalid timeout value")
			}
		})

		t.Run("NilConfig", func(t *testing.T) {
			err := client.Connect(nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "configuration is nil")
		})
	})

	t.Run("ConfigTimeouts", func(t *testing.T) {
		t.Run("ConnectionTimeout", func(t *testing.T) {
			client := elk.NewElkClient()
			cfg := &elk.ElkConfig{
				Addresses: []string{"http://nonexistent:9200"},
				Timeout:   time.Duration(1), // 1秒超时
			}

			start := time.Now()
			err := client.Connect(cfg)
			duration := time.Since(start)

			assert.Error(t, err)
			assert.True(t, duration < 2*time.Second, "连接超时应该在1秒左右")
		})
	})

	t.Run("ConfigAuthentication", func(t *testing.T) {
		client := elk.NewElkClient()

		t.Run("WithAuth", func(t *testing.T) {
			cfg := &elk.ElkConfig{
				Addresses: []string{"http://localhost:9200"},
				Username:  "",
				Password:  "",
				Timeout:   time.Duration(30),
			}
			err := client.Connect(cfg)
			require.NoError(t, err)
			assert.True(t, client.IsConnected())
		})

		t.Run("WithoutAuth", func(t *testing.T) {
			cfg := &elk.ElkConfig{
				Addresses: []string{"http://localhost:9200"},
				Timeout:   time.Duration(30),
			}
			err := client.Connect(cfg)
			require.NoError(t, err)
			assert.True(t, client.IsConnected())
		})
	})
}
