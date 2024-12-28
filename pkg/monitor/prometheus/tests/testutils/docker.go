package testutils

import (
	"context"
	"os"
	"testing"
	"time"

	"gobase/pkg/errors"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PrometheusContainer 表示Prometheus容器
type PrometheusContainer struct {
	Container testcontainers.Container
	URI       string
}

// Terminate 终止容器
func (p *PrometheusContainer) Terminate(ctx context.Context) error {
	if p.Container != nil {
		// 先尝试停止容器
		timeout := 10 * time.Second
		if err := p.Container.Stop(ctx, &timeout); err != nil {
			// 如果停止失败，记录错误但继续尝试终止
			return errors.NewSystemError("failed to stop container", err)
		}

		// 等待容器完全停止
		time.Sleep(2 * time.Second)

		// 终止容器
		if err := p.Container.Terminate(ctx); err != nil {
			return errors.NewSystemError("failed to terminate container", err)
		}
	}
	return nil
}

// StartPrometheusContainer 启动Prometheus容器
func StartPrometheusContainer(t *testing.T) (*PrometheusContainer, error) {
	ctx := context.Background()

	// 创建临时的Prometheus配置文件
	configContent := `
global:
  scrape_interval: 1s    # 设置较短的抓取间隔以加快测试
  evaluation_interval: 1s

scrape_configs:
  - job_name: 'test-metrics'
    static_configs:
      - targets: ['host.docker.internal:9091']  # 确保端口与 exporter 配置一致
    metrics_path: '/metrics'
    scrape_timeout: 1s    # 添加较短的超时时间
`
	// 创建临时配置文件
	tmpConfigFile, err := os.CreateTemp("", "prometheus-*.yml")
	if err != nil {
		return nil, errors.NewSystemError("failed to create temp config file", err)
	}
	defer os.Remove(tmpConfigFile.Name())

	if _, err := tmpConfigFile.WriteString(configContent); err != nil {
		return nil, errors.NewSystemError("failed to write config content", err)
	}
	if err := tmpConfigFile.Close(); err != nil {
		return nil, errors.NewSystemError("failed to close config file", err)
	}

	req := testcontainers.ContainerRequest{
		Image:        "prom/prometheus:latest",
		ExposedPorts: []string{"9090/tcp"},
		WaitingFor:   wait.ForHTTP("/-/ready").WithPort("9090/tcp").WithStartupTimeout(30 * time.Second),
		Files: []testcontainers.ContainerFile{
			{
				HostFilePath:      tmpConfigFile.Name(),
				ContainerFilePath: "/etc/prometheus/prometheus.yml",
				FileMode:          0644,
			},
		},
		// 添加额外的网络设置以支持访问宿主机
		ExtraHosts: []string{"host.docker.internal:host-gateway"},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, errors.NewSystemError("failed to start container", err)
	}

	// 获取容器映射端口
	mappedPort, err := container.MappedPort(ctx, "9090")
	if err != nil {
		// 如果获取端口失败，确保清理容器
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("Warning: failed to terminate container after port mapping error: %v", termErr)
		}
		return nil, errors.NewSystemError("failed to get mapped port", err)
	}

	// 获取容器主机
	host, err := container.Host(ctx)
	if err != nil {
		// 如果获取主机失败，确保清理容器
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("Warning: failed to terminate container after host error: %v", termErr)
		}
		return nil, errors.NewSystemError("failed to get container host", err)
	}

	return &PrometheusContainer{
		Container: container,
		URI:       "http://" + host + ":" + mappedPort.Port(),
	}, nil
}

// Stop 停止并清理容器
func (pc *PrometheusContainer) Stop() error {
	if err := pc.Container.Terminate(context.Background()); err != nil {
		return errors.NewSystemError("failed to stop container", err)
	}
	return nil
}
