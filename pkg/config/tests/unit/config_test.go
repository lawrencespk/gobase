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

	// 测试默认配置
	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := config.NewConfig()
		assert.NotNil(t, cfg)

		// 验证 ELK 默认配置
		assert.NotNil(t, cfg.ELK)
		assert.Equal(t, "5s", cfg.ELK.Bulk.Interval)          // 默认间隔
		assert.Equal(t, 1000, cfg.ELK.Bulk.BatchSize)         // 默认批次大小
		assert.Equal(t, 5*1024*1024, cfg.ELK.Bulk.FlushBytes) // 默认刷新字节数
		assert.Equal(t, 30, cfg.ELK.Timeout)                  // 默认超时时间

		// 验证 Logger 默认配置
		assert.NotNil(t, cfg.Logger)
		assert.Equal(t, "info", cfg.Logger.Level)
		assert.Equal(t, "console", cfg.Logger.Output)
	})

	// 测试配置深拷贝
	t.Run("ConfigCopy", func(t *testing.T) {
		original := &config.Config{
			ELK: config.ELKConfig{
				Addresses: []string{"http://localhost:9200"},
				Username:  "elastic",
				Password:  "password",
				Index:     "logs",
				Bulk: config.BulkConfig{
					BatchSize:  1000,
					FlushBytes: 5242880,
					Interval:   "5s",
				},
			},
			Logger: config.LoggerConfig{
				Level:  "debug",
				Output: "file",
			},
		}

		copied := original.Clone()

		// 验证深拷贝后的值相等
		assert.Equal(t, original.ELK.Addresses, copied.ELK.Addresses)
		assert.Equal(t, original.ELK.Username, copied.ELK.Username)
		assert.Equal(t, original.ELK.Password, copied.ELK.Password)
		assert.Equal(t, original.ELK.Bulk.BatchSize, copied.ELK.Bulk.BatchSize)
		assert.Equal(t, original.Logger.Level, copied.Logger.Level)

		// 验证是深拷贝而非浅拷贝
		copied.ELK.Addresses[0] = "changed"
		assert.NotEqual(t, original.ELK.Addresses[0], copied.ELK.Addresses[0])

		copied.ELK.Bulk.BatchSize = 2000
		assert.NotEqual(t, original.ELK.Bulk.BatchSize, copied.ELK.Bulk.BatchSize)

		copied.Logger.Level = "info"
		assert.NotEqual(t, original.Logger.Level, copied.Logger.Level)
	})

	// 测试配置合并
	t.Run("ConfigMerge", func(t *testing.T) {
		base := &config.Config{
			ELK: config.ELKConfig{
				Addresses: []string{"http://localhost:9200"},
				Username:  "elastic",
				Index:     "logs",
				Bulk: config.BulkConfig{
					BatchSize: 1000,
					Interval:  "5s",
				},
			},
			Logger: config.LoggerConfig{
				Level:  "info",
				Output: "console",
			},
		}

		override := &config.Config{
			ELK: config.ELKConfig{
				Password: "newpassword",
				Bulk: config.BulkConfig{
					BatchSize: 2000,
				},
			},
			Logger: config.LoggerConfig{
				Level: "debug",
			},
		}

		merged := base.Merge(override)

		// 验证保留基础配置
		assert.Equal(t, base.ELK.Addresses[0], merged.ELK.Addresses[0])
		assert.Equal(t, base.ELK.Username, merged.ELK.Username)
		assert.Equal(t, base.ELK.Index, merged.ELK.Index)
		assert.Equal(t, base.ELK.Bulk.Interval, merged.ELK.Bulk.Interval)
		assert.Equal(t, base.Logger.Output, merged.Logger.Output)

		// 验证覆盖的配置
		assert.Equal(t, override.ELK.Password, merged.ELK.Password)
		assert.Equal(t, override.ELK.Bulk.BatchSize, merged.ELK.Bulk.BatchSize)
		assert.Equal(t, override.Logger.Level, merged.Logger.Level)

		// 验证原配置未被修改
		assert.NotEqual(t, base.ELK.Password, override.ELK.Password)
		assert.NotEqual(t, base.ELK.Bulk.BatchSize, override.ELK.Bulk.BatchSize)
		assert.NotEqual(t, base.Logger.Level, override.Logger.Level)
	})

	// 测试配置序列化和反序列化
	t.Run("ConfigSerialization", func(t *testing.T) {
		original := &config.Config{
			ELK: config.ELKConfig{
				Addresses: []string{"http://localhost:9200"},
				Username:  "elastic",
				Password:  "password",
				Index:     "logs",
				Bulk: config.BulkConfig{
					BatchSize:  1000,
					FlushBytes: 5242880,
					Interval:   "5s",
				},
			},
			Logger: config.LoggerConfig{
				Level:  "debug",
				Output: "file",
			},
		}

		// 测试 YAML 序列化
		yamlData, err := config.MarshalYAML(original)
		assert.NoError(t, err)
		assert.NotEmpty(t, yamlData)

		// 测试 YAML 反序列化
		restored := &config.Config{}
		err = config.UnmarshalYAML(yamlData, restored)
		assert.NoError(t, err)
		assert.Equal(t, original.ELK.Addresses, restored.ELK.Addresses)
		assert.Equal(t, original.ELK.Username, restored.ELK.Username)
		assert.Equal(t, original.ELK.Bulk.BatchSize, restored.ELK.Bulk.BatchSize)
		assert.Equal(t, original.Logger.Level, restored.Logger.Level)
	})

	// 测试配置验证的边界情况
	t.Run("ConfigValidationEdgeCases", func(t *testing.T) {
		tests := []struct {
			name    string
			cfg     *config.Config
			wantErr string
		}{
			{
				name: "empty addresses but with other valid fields",
				cfg: &config.Config{
					ELK: config.ELKConfig{
						Username: "elastic",
						Password: "password",
						Index:    "logs",
						Timeout:  30,
					},
				},
				wantErr: "elk addresses is empty",
			},
			{
				name: "invalid timeout value",
				cfg: &config.Config{
					ELK: config.ELKConfig{
						Addresses: []string{"http://localhost:9200"},
						Username:  "elastic",
						Password:  "password",
						Index:     "logs",
						Timeout:   0,
					},
				},
				wantErr: "elk timeout must be greater than 0",
			},
			{
				name: "invalid bulk batch size",
				cfg: &config.Config{
					ELK: config.ELKConfig{
						Addresses: []string{"http://localhost:9200"},
						Username:  "elastic",
						Password:  "password",
						Index:     "logs",
						Timeout:   30,
						Bulk: config.BulkConfig{
							BatchSize:  -1,
							FlushBytes: 5242880,
							Interval:   "5s",
						},
					},
				},
				wantErr: "elk bulk batch size must be greater than 0",
			},
			{
				name: "invalid bulk flush bytes",
				cfg: &config.Config{
					ELK: config.ELKConfig{
						Addresses: []string{"http://localhost:9200"},
						Username:  "elastic",
						Password:  "password",
						Index:     "logs",
						Timeout:   30,
						Bulk: config.BulkConfig{
							BatchSize:  1000,
							FlushBytes: 0,
							Interval:   "5s",
						},
					},
				},
				wantErr: "elk bulk flush bytes must be greater than 0",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := config.ValidateConfig(tt.cfg)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
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
