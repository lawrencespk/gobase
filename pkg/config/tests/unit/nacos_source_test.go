package unit

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/config/source/nacos"
	"gobase/pkg/config/source/nacos/mock"
	configTypes "gobase/pkg/config/types"

	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

func TestNacosSource(t *testing.T) {
	// 创建mock客户端
	mockClient := mock.NewMockConfigClient()

	// 基础功能测试
	t.Run("BasicOperations", func(t *testing.T) {
		opts := &nacos.Options{
			Endpoints:   []string{"localhost:8848"},
			NamespaceID: "public",
			Group:       "DEFAULT_GROUP",
			DataID:      "test.yaml",
		}

		// 预先设置配置
		mockClient.SetConfig(opts.DataID, `
elk:
  addresses:
    - http://localhost:9200
`)

		src, err := nacos.NewSourceWithClient(mockClient, opts)
		require.NoError(t, err)

		ctx := context.Background()
		err = src.Load(ctx)
		require.NoError(t, err)

		val, err := src.Get("elk.addresses")
		require.NoError(t, err)
		assert.NotNil(t, val)
	})

	// 配置监听测试
	t.Run("ConfigWatch", func(t *testing.T) {
		changeChan := make(chan configTypes.Event, 1)
		opts := &nacos.Options{
			Endpoints:   []string{"localhost:8848"},
			NamespaceID: "public",
			Group:       "DEFAULT_GROUP",
			DataID:      "test.yaml",
			OnChange: func(event configTypes.Event) {
				changeChan <- event
			},
		}

		// 预先设置配置
		mockClient.SetConfig(opts.DataID, `
elk:
  addresses:
    - http://localhost:9200
`)

		src, err := nacos.NewSourceWithClient(mockClient, opts)
		require.NoError(t, err)

		ctx := context.Background()
		err = src.Load(ctx)
		require.NoError(t, err)

		err = src.Watch(ctx)
		require.NoError(t, err)

		// 发布新配置
		go func() {
			time.Sleep(100 * time.Millisecond)
			success, err := mockClient.PublishConfig(vo.ConfigParam{
				DataId: opts.DataID,
				Group:  opts.Group,
				Content: `
elk:
  addresses:
    - http://localhost:9200
    - http://localhost:9201
`,
			})
			require.NoError(t, err)
			require.True(t, success)
		}()

		select {
		case event := <-changeChan:
			assert.Equal(t, opts.DataID, event.Key)
			assert.Equal(t, configTypes.EventUpdate, event.Type)
		case <-time.After(2 * time.Second):
			t.Fatal("no config change detected within timeout")
		}
	})

	// 错误处理测试
	t.Run("ErrorHandling", func(t *testing.T) {
		errorMockClient := mock.NewMockConfigClient()
		errorMockClient.SetError(true)

		opts := &nacos.Options{
			Endpoints:   []string{"localhost:8848"},
			NamespaceID: "public",
			Group:       "DEFAULT_GROUP",
			DataID:      "test.yaml",
		}

		src, err := nacos.NewSourceWithClient(errorMockClient, opts)
		require.NoError(t, err)

		ctx := context.Background()
		err = src.Load(ctx)
		assert.Error(t, err)
	})

	// 重连测试
	t.Run("Reconnection", func(t *testing.T) {
		opts := &nacos.Options{
			Endpoints:   []string{"localhost:8848"},
			NamespaceID: "public",
			Group:       "DEFAULT_GROUP",
			DataID:      "test.yaml",
		}

		// 预先设置配置
		mockClient.SetConfig(opts.DataID, `
elk:
  addresses:
    - http://localhost:9200
`)

		src, err := nacos.NewSourceWithClient(mockClient, opts)
		require.NoError(t, err)

		ctx := context.Background()
		err = src.Load(ctx)
		require.NoError(t, err)

		err = src.Watch(ctx)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)

		val, err := src.Get("elk.addresses")
		assert.NoError(t, err)
		assert.NotNil(t, val)
	})

	// 并发测试
	t.Run("ConcurrentAccess", func(t *testing.T) {
		opts := &nacos.Options{
			Endpoints:   []string{"localhost:8848"},
			NamespaceID: "public",
			Group:       "DEFAULT_GROUP",
			DataID:      "test.yaml",
		}

		// 预先设置配置
		mockClient.SetConfig(opts.DataID, `
foo: bar
number: 42
`)

		src, err := nacos.NewSourceWithClient(mockClient, opts)
		require.NoError(t, err)

		ctx := context.Background()
		err = src.Load(ctx)
		require.NoError(t, err)

		// 等待配置加载完成
		time.Sleep(100 * time.Millisecond)

		var wg sync.WaitGroup
		errChan := make(chan error, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				val, err := src.Get("foo")
				if err != nil {
					errChan <- err
					return
				}
				if val != "bar" {
					errChan <- fmt.Errorf("unexpected value: %v", val)
				}
			}()
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			t.Errorf("concurrent access error: %v", err)
		}
	})

	// 选项验证测试
	t.Run("OptionsValidation", func(t *testing.T) {
		tests := []struct {
			name    string
			opts    *nacos.Options
			wantErr bool
		}{
			{
				name: "ValidOptions",
				opts: &nacos.Options{
					Endpoints:   []string{"localhost:8848"},
					NamespaceID: "public",
					Group:       "DEFAULT_GROUP",
					DataID:      "test.yaml",
				},
				wantErr: false,
			},
			{
				name:    "NilOptions",
				opts:    nil,
				wantErr: true,
			},
			{
				name: "EmptyDataID",
				opts: &nacos.Options{
					Endpoints:   []string{"localhost:8848"},
					NamespaceID: "public",
					Group:       "DEFAULT_GROUP",
					DataID:      "",
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := nacos.NewSourceWithClient(mockClient, tt.opts)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	// 关闭测试
	t.Run("Close", func(t *testing.T) {
		mockClient := mock.NewMockConfigClient()

		// 设置期望
		mockClient.On("CloseClient").Return().Once()

		opts := &nacos.Options{
			Endpoints:   []string{"localhost:8848"},
			NamespaceID: "public",
			Group:       "DEFAULT_GROUP",
			DataID:      "test.yaml",
		}

		// 预先设置配置
		mockClient.SetConfig(opts.DataID, `
foo: bar
`)

		src, err := nacos.NewSourceWithClient(mockClient, opts)
		require.NoError(t, err)

		// 执行关闭
		err = src.Close(context.Background())
		require.NoError(t, err)

		// 验证期望
		mockClient.AssertExpectations(t)
	})
}
