package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gobase/pkg/config"
	"gobase/pkg/trace/jaeger"

	"github.com/docker/go-connections/nat"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var jaegerHost string

func TestMain(m *testing.M) {
	// 启动Jaeger容器
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "jaegertracing/all-in-one:latest",
		ExposedPorts: []string{"6831/udp", "14268/tcp", "16686/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithPort(nat.Port("16686/tcp")),
		Env: map[string]string{
			"COLLECTOR_ZIPKIN_HOST_PORT": ":9411",
		},
	}

	jaegerC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Printf("Failed to start jaeger container: %v\n", err)
		os.Exit(1)
	}

	// 获取容器IP
	host, err := jaegerC.Host(ctx)
	if err != nil {
		fmt.Printf("Failed to get jaeger host: %v\n", err)
		os.Exit(1)
	}
	jaegerHost = host

	// 运行测试
	code := m.Run()

	// 清理容器
	if err := jaegerC.Terminate(ctx); err != nil {
		fmt.Printf("Failed to terminate jaeger container: %v\n", err)
	}

	os.Exit(code)
}

func TestJaegerIntegration(t *testing.T) {
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

	// 获取并修改配置
	cfg := config.GetConfig()
	require.NotNil(t, cfg)

	// 创建容器请求配置
	containerReq := testcontainers.ContainerRequest{
		Image:        "jaegertracing/all-in-one:latest",
		ExposedPorts: []string{"6831/udp", "14268/tcp", "16686/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithPort(nat.Port("16686/tcp")),
		Env: map[string]string{
			"COLLECTOR_ZIPKIN_HOST_PORT": ":9411",
		},
	}

	// 启动容器
	ctx := context.Background()
	jaegerC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: containerReq,
		Started:          true,
	})
	require.NoError(t, err)
	defer jaegerC.Terminate(ctx)

	// 获取映射后的端口
	mappedPort, err := jaegerC.MappedPort(ctx, "14268/tcp")
	require.NoError(t, err)

	// 更新全局配置中的 Jaeger 配置
	cfg.Jaeger.Enable = true
	cfg.Jaeger.ServiceName = "jaeger-test"
	cfg.Jaeger.Agent.Host = "localhost"
	cfg.Jaeger.Agent.Port = "6831"
	cfg.Jaeger.Collector.Endpoint = fmt.Sprintf("http://localhost:%s/api/traces", mappedPort.Port())
	cfg.Jaeger.Collector.Timeout = 5
	cfg.Jaeger.Sampler.Type = "const"
	cfg.Jaeger.Sampler.Param = 1

	// 重新加载配置以应用更改
	config.SetConfig(cfg)

	// 创建和测试追踪
	provider, err := jaeger.NewProvider()
	require.NoError(t, err)
	defer provider.Close()

	// 测试完整追踪流程
	t.Run("Full trace flow", func(t *testing.T) {
		rootSpan := provider.Tracer().StartSpan("root-operation")
		rootCtx := opentracing.ContextWithSpan(context.Background(), rootSpan)

		childSpan, _ := jaeger.StartSpanFromContext(rootCtx, "child-operation")
		childSpan.SetTag("custom.tag", "test-value")
		childSpan.LogKV("event", "test-event")

		time.Sleep(100 * time.Millisecond)
		childSpan.Finish()
		rootSpan.Finish()

		// 给 Jaeger 一些时间来处理 spans
		time.Sleep(500 * time.Millisecond)
	})
}

func TestSamplingStrategies(t *testing.T) {
	// 确保配置已初始化
	if config.GetConfig() == nil {
		// 获取当前工作目录
		pwd, err := os.Getwd()
		require.NoError(t, err)

		// 构建到项目根目录的路径并设置配置文件路径
		rootDir := filepath.Join(pwd, "../../../../../")
		configPath := filepath.Join(rootDir, "config", "config.yaml")

		config.SetConfigPath(configPath)
		err = config.LoadConfig()
		require.NoError(t, err)
	}

	tests := []struct {
		name       string
		samplerCfg jaeger.SamplerConfig
	}{
		{
			name: "Constant sampler",
			samplerCfg: jaeger.SamplerConfig{
				Type:  "const",
				Param: 1,
			},
		},
		// ... 其他采样策略测试用例
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 配置和测试每种采样策略
			// ... 测试代码
		})
	}
}
