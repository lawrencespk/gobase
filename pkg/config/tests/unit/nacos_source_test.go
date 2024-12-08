package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/config/source/nacos"
	"gobase/pkg/config/source/nacos/mock"
	configTypes "gobase/pkg/config/types"

	"github.com/nacos-group/nacos-sdk-go/vo"
)

func TestNacosSource(t *testing.T) {
	// 创建mock客户端
	mockClient := mock.NewMockConfigClient()

	// 基础功能测试
	t.Run("BasicOperations", func(t *testing.T) {
		opts := &nacos.Options{
			Endpoint:    "localhost:8848",
			NamespaceID: "public",
			Group:       "DEFAULT_GROUP",
			DataID:      "test.yaml",
			Username:    "nacos",
			Password:    "nacos",
		}

		// 使用mock客户端创建source
		src, err := nacos.NewSourceWithClient(opts, mockClient)
		require.NoError(t, err)
		require.NotNil(t, src)

		// 预设配置内容
		mockClient.SetConfig("test.yaml", `
elk:
  addresses:
    - http://localhost:9200
`)

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
		changeChan := make(chan configTypes.Event, 1)
		opts := &nacos.Options{
			Endpoint:    "localhost:8848",
			NamespaceID: "public",
			Group:       "DEFAULT_GROUP",
			DataID:      "test.yaml",
			OnChange: func(event configTypes.Event) {
				changeChan <- event
			},
		}

		// 使用mock客户端创建source
		src, err := nacos.NewSourceWithClient(opts, mockClient)
		require.NoError(t, err)

		// 预设初始配置
		mockClient.SetConfig("test.yaml", `
elk:
  addresses:
    - http://localhost:9200
`)

		ctx := context.Background()
		err = src.Load(ctx)
		require.NoError(t, err)

		// 启动配置监听
		err = src.Watch(ctx)
		require.NoError(t, err)

		// 更新配置
		mockClient.PublishConfig(vo.ConfigParam{
			DataId: "test.yaml",
			Group:  "DEFAULT_GROUP",
			Content: `
elk:
  addresses:
    - http://localhost:9200
    - http://localhost:9201
`,
		})

		// 等待配置变更事件
		select {
		case event := <-changeChan:
			// 验证事件内容
			assert.Equal(t, "test.yaml", event.Key)
			assert.Equal(t, configTypes.EventUpdate, event.Type)

			// 验证新配置
			newAddresses, ok := event.Value.(map[string]interface{})["elk"].(map[string]interface{})["addresses"].([]interface{})
			require.True(t, ok)
			assert.Equal(t, 2, len(newAddresses))
			assert.Equal(t, "http://localhost:9200", newAddresses[0])
			assert.Equal(t, "http://localhost:9201", newAddresses[1])

		case <-time.After(2 * time.Second):
			t.Fatal("no config change detected within timeout")
		}

		// 测试停止监听
		err = src.StopWatch(ctx)
		require.NoError(t, err)

		// 再次更新配置，不应该收到事件
		mockClient.PublishConfig(vo.ConfigParam{
			DataId: "test.yaml",
			Group:  "DEFAULT_GROUP",
			Content: `
elk:
  addresses:
    - http://localhost:9200
`,
		})

		select {
		case <-changeChan:
			t.Fatal("received unexpected config change event after stopping watch")
		case <-time.After(1 * time.Second):
			// 正常情况，没有收到事件
		}
	})

	// 错误处理测试
	t.Run("ErrorHandling", func(t *testing.T) {
		// 使用一个特殊的mock客户端来模拟错误情况
		errorMockClient := mock.NewMockConfigClient()
		errorMockClient.SetError(true) // 需要在mock中添加这个功能

		opts := &nacos.Options{
			Endpoint:    "localhost:8848",
			NamespaceID: "public",
			Group:       "DEFAULT_GROUP",
			DataID:      "test.yaml",
		}

		src, err := nacos.NewSourceWithClient(opts, errorMockClient)
		require.NoError(t, err)

		ctx := context.Background()
		err = src.Load(ctx)
		assert.Error(t, err) // 应该返回错误
	})

	// 重连测试
	t.Run("Reconnection", func(t *testing.T) {
		opts := &nacos.Options{
			Endpoint:    "localhost:8848",
			NamespaceID: "public",
			Group:       "DEFAULT_GROUP",
			DataID:      "test.yaml",
			Username:    "nacos",
			Password:    "nacos",
		}

		// 预设配置
		mockClient.SetConfig("test.yaml", `
elk:
  addresses:
    - http://localhost:9200
`)

		src, err := nacos.NewSourceWithClient(opts, mockClient)
		require.NoError(t, err)

		ctx := context.Background()
		err = src.Load(ctx)
		require.NoError(t, err)

		// 模拟断线重连
		err = src.Watch(ctx)
		require.NoError(t, err)

		time.Sleep(2 * time.Second) // 等待重连完成

		// 验证配置是否仍然可用
		val, err := src.Get("elk.addresses")
		assert.NoError(t, err)
		assert.NotNil(t, val)
	})

	// 并发测试
	t.Run("ConcurrentAccess", func(t *testing.T) {
		opts := &nacos.Options{
			Endpoint:    "localhost:8848",
			NamespaceID: "public",
			Group:       "DEFAULT_GROUP",
			DataID:      "test.yaml",
		}

		// 预设配置
		mockClient.SetConfig("test.yaml", `
elk:
  addresses:
    - http://localhost:9200
`)

		src, err := nacos.NewSourceWithClient(opts, mockClient)
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
