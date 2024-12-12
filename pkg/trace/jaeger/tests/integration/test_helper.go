package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/config"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
	"gobase/pkg/trace/jaeger/tests/testutils"
)

var (
	once           sync.Once
	instance       *testHelper
	containerReady bool
	log            types.Logger
)

func init() {
	// 初始化日志
	var err error
	log, err = logger.NewLogger()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
}

type testHelper struct {
	cleanup func()
}

func setupTest(t *testing.T) *testHelper {
	t.Helper()

	once.Do(func() {
		// 启动Jaeger容器（只执行一次）
		cleanup, err := testutils.StartJaegerContainer(context.Background())
		require.NoError(t, err)

		// 等待Jaeger服务就绪，增加重试机制
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		for {
			if err := testutils.WaitForJaeger(ctx); err == nil {
				containerReady = true
				break
			}
			select {
			case <-ctx.Done():
				t.Fatal("timeout waiting for Jaeger container")
				return
			case <-time.After(time.Second):
				continue
			}
		}

		// 获取当前工作目录
		pwd, err := os.Getwd()
		require.NoError(t, err)

		// 构建到项目根目录的路径并设置配置文件路径
		rootDir := filepath.Join(pwd, "../../../../../")
		configPath := filepath.Join(rootDir, "config", "config.yaml")

		// 设置并加载配置
		config.SetConfigPath(configPath)
		err = config.LoadConfig()
		require.NoError(t, err)

		// 获取配置并验证
		cfg := config.GetConfig()
		require.NotNil(t, cfg)
		require.True(t, cfg.Jaeger.Enable)
		require.Equal(t, "localhost", cfg.Jaeger.Agent.Host)
		require.Equal(t, "6831", cfg.Jaeger.Agent.Port)

		instance = &testHelper{
			cleanup: cleanup,
		}

		// 注册清理函数
		t.Cleanup(func() {
			if instance != nil && instance.cleanup != nil {
				instance.cleanup()
			}
		})
	})

	// 确保容器已就绪
	require.True(t, containerReady, "Jaeger container is not ready")

	return instance
}

// 添加辅助方

// WaitForSpans 等待span上报完成
func WaitForSpans(duration time.Duration) {
	time.Sleep(duration)
}
