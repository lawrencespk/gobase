package testutils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PrometheusContainer 封装Prometheus容器相关操作
type PrometheusContainer struct {
	Container testcontainers.Container
	URI       string
}

// StartPrometheusContainer 启动Prometheus容器
func StartPrometheusContainer(t *testing.T) (*PrometheusContainer, error) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "prom/prometheus:latest",
		ExposedPorts: []string{"9090/tcp"},
		WaitingFor:   wait.ForHTTP("/-/ready").WithPort("9090/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	// 注册cleanup函数
	t.Cleanup(func() {
		if err := container.Terminate(ctx); err != nil {
			t.Errorf("failed to terminate container: %v", err)
		}
	})

	mappedPort, err := container.MappedPort(ctx, "9090")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapped port: %v", err)
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get host: %v", err)
	}

	prometheusURI := fmt.Sprintf("http://%s:%s", hostIP, mappedPort.Port())
	return &PrometheusContainer{
		Container: container,
		URI:       prometheusURI,
	}, nil
}

// Stop 停止并清理容器
func (pc *PrometheusContainer) Stop() error {
	return pc.Container.Terminate(context.Background())
}
