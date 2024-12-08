package integration

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/config/source/local"
	configTypes "gobase/pkg/config/types"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

func TestLocalConfigIntegration(t *testing.T) {
	// 创建测试目录
	testDir := t.TempDir()
	configPath := filepath.Join(testDir, "config.yaml")

	// 初始配置内容
	initialConfig := []byte(`
elk:
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
logger:
    level: info
    output: stdout
`)

	require.NoError(t, os.WriteFile(configPath, initialConfig, 0644))

	t.Run("FullConfigLifecycle", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)
		changes := make(chan configTypes.Event, 1)

		// 创建本地配置源
		src, err := local.NewSource(&local.Options{
			Path: configPath,
			OnChange: func(event configTypes.Event) {
				changes <- event
				wg.Done()
			},
		})
		require.NoError(t, err)

		// 加载配置
		ctx := context.Background()
		err = src.Load(ctx)
		require.NoError(t, err)

		// 测试配置读取
		addresses, err := src.Get("elk.addresses")
		require.NoError(t, err)
		assert.NotNil(t, addresses)

		// 开始监听
		err = src.Watch(ctx)
		require.NoError(t, err)

		// 修改配置文件
		newConfig := []byte(`
elk:
    addresses:
        - http://localhost:9200
        - http://localhost:9201
    username: elastic
    password: password
    index: logs
    timeout: 30
    bulk:
        batchSize: 1000
        flushBytes: 5242880
        interval: 5s
logger:
    level: debug
    output: file
`)
		require.NoError(t, os.WriteFile(configPath, newConfig, 0644))

		// 等待配置变更事件
		select {
		case event := <-changes:
			assert.Equal(t, configTypes.EventUpdate, event.Type)
			// 验证新配置
			newAddresses, err := src.Get("elk.addresses")
			require.NoError(t, err)
			assert.NotNil(t, newAddresses)
		case <-time.After(5 * time.Second):
			t.Fatal("配置变更超时")
		}

		wg.Wait()

		// 测试关闭
		require.NoError(t, src.Close(ctx))
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		src, err := local.NewSource(&local.Options{
			Path: configPath,
		})
		require.NoError(t, err)

		ctx := context.Background()
		err = src.Load(ctx)
		require.NoError(t, err)

		var wg sync.WaitGroup
		concurrency := 10

		// 并发读取测试
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					_, err := src.Get("elk.addresses")
					assert.NoError(t, err)

					_, err = src.Get("logger.level")
					assert.NoError(t, err)
				}
			}()
		}

		wg.Wait()
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// 测试无效配置文件
		invalidPath := filepath.Join(testDir, "invalid.yaml")
		require.NoError(t, os.WriteFile(invalidPath, []byte("invalid: ][yaml"), 0644))

		src, err := local.NewSource(&local.Options{
			Path: invalidPath,
		})
		require.NoError(t, err)

		// 应该返回解析错误
		err = src.Load(context.Background())
		assert.Error(t, err)

		// 测试文件不存在
		src, err = local.NewSource(&local.Options{
			Path: "nonexistent.yaml",
		})
		require.NoError(t, err)

		err = src.Load(context.Background())
		assert.Error(t, err)
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		// 测试配置验证
		invalidConfig := []byte(`
elk:
    addresses: []  # 空地址列表
    username: ""   # 空用户名
    bulk:
        batchSize: -1  # 无效的批次大小
logger:
    level: invalid_level  # 无效的日志级别
    output: invalid_output  # 无效的输出类型
`)
		require.NoError(t, os.WriteFile(configPath, invalidConfig, 0644))

		src, err := local.NewSource(&local.Options{
			Path: configPath,
			ValidateConfig: func(cfg map[string]interface{}) error {
				// 验证 ELK 配置
				if elk, ok := cfg["elk"].(map[string]interface{}); ok {
					// 验证地址列表
					if addresses, ok := elk["addresses"].([]interface{}); ok && len(addresses) == 0 {
						return errors.NewConfigError("elk addresses cannot be empty", nil)
					}
					// 验证用户名
					if username, ok := elk["username"].(string); ok && username == "" {
						return errors.NewConfigError("elk username cannot be empty", nil)
					}
					// 验证批次大小
					if bulk, ok := elk["bulk"].(map[string]interface{}); ok {
						if batchSize, ok := bulk["batchSize"].(int); ok && batchSize <= 0 {
							return errors.NewConfigError("elk bulk batch size must be greater than 0", nil)
						}
					}
				}
				return nil
			},
		})
		require.NoError(t, err)

		// 应该返回验证错误
		err = src.Load(context.Background())
		assert.Error(t, err)

		// 验证错误类型
		assert.Equal(t, codes.ConfigError, errors.GetErrorCode(err))
		assert.Contains(t, err.Error(), "elk addresses cannot be empty")
	})
}
