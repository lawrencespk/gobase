package unit

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/config/source/local"
	configTypes "gobase/pkg/config/types"
)

func TestLocalSource(t *testing.T) {
	// 基础功能测试
	t.Run("BasicOperations", func(t *testing.T) {
		// 创建临时配置文件
		configPath, cleanup := createTempConfig(t)
		defer cleanup()

		// 创建本地配置源
		src, err := local.NewSource(&local.Options{
			Path: configPath,
		})
		require.NoError(t, err)
		require.NotNil(t, src)

		// 测试加载配置
		ctx := context.Background()
		err = src.Load(ctx)
		require.NoError(t, err)

		// 测试读取配置值
		val, err := src.Get("elk.addresses")
		require.NoError(t, err)
		assert.NotNil(t, val)

		// 测试关闭配置源
		err = src.Close(ctx)
		require.NoError(t, err)
	})

	// 配置监听测试
	t.Run("ConfigWatch", func(t *testing.T) {
		// 创建临时配置文件
		configPath, cleanup := createTempConfig(t)
		defer cleanup()

		// 创建变更通知通道
		changes := make(chan configTypes.Event)

		// 创建本地配置源
		src, err := local.NewSource(&local.Options{
			Path: configPath,
			OnChange: func(e configTypes.Event) {
				changes <- e
			},
		})
		require.NoError(t, err)

		// 启动配置监听
		ctx := context.Background()
		err = src.Watch(ctx)
		require.NoError(t, err)

		// 修改配置文件
		newConfig := []byte(`elk:
  addresses:
    - http://localhost:9201
  username: elastic
  password: newpassword
`)
		err = os.WriteFile(configPath, newConfig, 0644)
		require.NoError(t, err)

		// 等待配置变更事件
		select {
		case event := <-changes:
			assert.Equal(t, configTypes.EventUpdate, event.Type)
			assert.Equal(t, configPath, event.Key)
			assert.NotNil(t, event.Value)
		case <-time.After(3 * time.Second):
			t.Fatal("timeout waiting for config change event")
		}

		// 停止监听
		err = src.StopWatch(ctx)
		require.NoError(t, err)
	})

	// 错误处理测试
	t.Run("ErrorHandling", func(t *testing.T) {
		// 测试无效的配置文件路径
		src, err := local.NewSource(&local.Options{
			Path: "nonexistent.yaml",
		})
		require.NoError(t, err) // 创建源不应该出错

		// 加载不存在的配置文件应该返回错误
		ctx := context.Background()
		err = src.Load(ctx)
		assert.Error(t, err)

		// 测试损坏的配置文件
		invalidPath, cleanup := createInvalidConfig(t)
		defer cleanup()

		src, err = local.NewSource(&local.Options{
			Path: invalidPath,
		})
		require.NoError(t, err)

		err = src.Load(ctx)
		assert.Error(t, err)
	})

	// 并发测试
	t.Run("ConcurrentAccess", func(t *testing.T) {
		configPath, cleanup := createTempConfig(t)
		defer cleanup()

		src, err := local.NewSource(&local.Options{
			Path: configPath,
		})
		require.NoError(t, err)

		ctx := context.Background()
		err = src.Load(ctx)
		require.NoError(t, err)

		// 并发读取测试
		concurrency := 10
		done := make(chan bool)

		for i := 0; i < concurrency; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					_, err := src.Get("elk.addresses")
					assert.NoError(t, err)
				}
				done <- true
			}()
		}

		// 等待所有goroutine完成
		for i := 0; i < concurrency; i++ {
			<-done
		}
	})
}

// 创建无效的配置文件
func createInvalidConfig(t *testing.T) (string, func()) {
	t.Helper()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "invalid.yaml")

	// 写入无效的YAML内容
	invalidContent := []byte(`
elk:
  addresses:
    - http://localhost:9200
  username: elastic
  password: password
  invalid yaml content
`)

	err := os.WriteFile(configPath, invalidContent, 0644)
	require.NoError(t, err)

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return configPath, cleanup
}
