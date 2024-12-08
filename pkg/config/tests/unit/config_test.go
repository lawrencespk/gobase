package unit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/config"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

func TestConfig(t *testing.T) {
	// 测试配置结构体
	t.Run("ConfigStruct", func(t *testing.T) {
		cfg := &config.Config{
			ELK: config.ELKConfig{
				Addresses: []string{"http://localhost:9200"},
				Username:  "elastic",
				Password:  "password",
				Index:     "logs",
				Timeout:   30,
				Bulk: config.BulkConfig{
					BatchSize:  1000,
					FlushBytes: 5242880, // 5MB
					Interval:   "5s",
				},
			},
		}

		// 验证ELK配置
		assert.NotEmpty(t, cfg.ELK.Addresses)
		assert.NotEmpty(t, cfg.ELK.Username)
		assert.NotEmpty(t, cfg.ELK.Password)
		assert.NotEmpty(t, cfg.ELK.Index)
		assert.Greater(t, cfg.ELK.Timeout, 0)

		// 验证Bulk配置
		assert.Greater(t, cfg.ELK.Bulk.BatchSize, 0)
		assert.Greater(t, cfg.ELK.Bulk.FlushBytes, 0)
		assert.NotEmpty(t, cfg.ELK.Bulk.Interval)
	})

	// 测试全局配置操作
	t.Run("GlobalConfig", func(t *testing.T) {
		// 初始状态应该为nil
		assert.Nil(t, config.GetConfig())

		// 设置新配置
		newCfg := &config.Config{
			ELK: config.ELKConfig{
				Addresses: []string{"http://localhost:9200"},
			},
		}
		config.SetConfig(newCfg)

		// 获取并验证配置
		cfg := config.GetConfig()
		assert.NotNil(t, cfg)
		assert.Equal(t, newCfg.ELK.Addresses, cfg.ELK.Addresses)
	})

	// 测试配置加载
	t.Run("LoadConfig", func(t *testing.T) {
		// 准备测试配置文件
		configPath, cleanup := createTempConfig(t)
		defer cleanup()

		// 设置配置文件路径
		config.SetConfigPath(configPath)

		err := config.LoadConfig()
		if err != nil {
			assert.True(t, errors.Is(err, errors.NewConfigError("", nil)), "应该返回配置错误类型")
		}
		require.NoError(t, err)

		cfg := config.GetConfig()
		require.NotNil(t, cfg)

		// 验证加载的配置
		assert.NotEmpty(t, cfg.ELK.Addresses)
		assert.Equal(t, "elastic", cfg.ELK.Username)
		assert.Equal(t, "password", cfg.ELK.Password)
		assert.Equal(t, "logs", cfg.ELK.Index)
		assert.Equal(t, 30, cfg.ELK.Timeout)
		assert.Equal(t, 1000, cfg.ELK.Bulk.BatchSize)
		assert.Equal(t, 5242880, cfg.ELK.Bulk.FlushBytes)
		assert.Equal(t, "5s", cfg.ELK.Bulk.Interval)
	})

	// 测试配置验证
	t.Run("ValidateConfig", func(t *testing.T) {
		invalidCfg := &config.Config{
			ELK: config.ELKConfig{
				// 缺少必要字段
			},
		}

		// 先保存当前配置
		oldCfg := config.GetConfig()

		// 设置无效配置
		config.SetConfig(invalidCfg)

		// 验证配置应该失败
		err := config.ValidateConfig(invalidCfg)
		assert.Error(t, err)
		assert.Equal(t, codes.ConfigError, errors.GetErrorCode(err), "应该返回配置错误类型")
		assert.Contains(t, err.Error(), "elk addresses is empty")

		// 恢复原配置
		config.SetConfig(oldCfg)
	})
}

// 测试辅助函数
func createTempConfig(t *testing.T) (string, func()) {
	t.Helper() // 标记这是一个辅助函数

	// 创建临时目录
	dir := t.TempDir()

	// 创建配置文件
	configPath := filepath.Join(dir, "config.yaml")
	configData := []byte(`elk:
  addresses:
    - http://localhost:9200
  username: elastic
  password: password
  index: logs
  timeout: 30
  bulk:
    batchSize: 1000
    flushBytes: 5242880
    interval: 5s
`)

	err := os.WriteFile(configPath, configData, 0644)
	require.NoError(t, err)

	// 返回清理函数
	cleanup := func() {
		os.RemoveAll(dir)
	}

	return configPath, cleanup
}
