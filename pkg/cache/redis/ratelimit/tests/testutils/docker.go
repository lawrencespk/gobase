package testutils

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RedisContainer 封装 Redis 测试容器
type RedisContainer struct {
	testcontainers.Container
	Address string
}

// StartRedisContainer 启动 Redis 测试容器
func StartRedisContainer(tb testing.TB) (*RedisContainer, error) {
	ctx := context.Background()

	// 修改这里：使用正确的端口映射配置
	req := testcontainers.ContainerRequest{
		Image: "redis:6-alpine",
		ExposedPorts: []string{
			"6379/tcp",
		},
		// 使用 nat.PortMap 进行端口映射
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.PortBindings = nat.PortMap{
				"6379/tcp": []nat.PortBinding{
					{HostIP: "0.0.0.0", HostPort: "0"}, // 使用动态端口
				},
			}
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("Ready to accept connections"),
			wait.ForListeningPort("6379/tcp"),
		),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %v", err)
	}

	// 获取容器地址
	mappedPort, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return nil, fmt.Errorf("failed to get container external port: %v", err)
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get container host: %v", err)
	}

	address := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())

	return &RedisContainer{
		Container: container,
		Address:   address,
	}, nil
}

// GetAddress 获取 Redis 容器地址
func (c *RedisContainer) GetAddress() string {
	return c.Address
}

// Terminate 停止并删除容器
func (c *RedisContainer) Terminate(tb testing.TB) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := c.Container.Terminate(ctx); err != nil {
		tb.Fatalf("failed to terminate container: %v", err)
	}
}
