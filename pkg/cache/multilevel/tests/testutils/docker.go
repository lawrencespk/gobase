package testutils

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
)

// StartRedisContainer 启动Redis容器并返回客户端和地址
func StartRedisContainer(ctx context.Context) (redis.Client, string, error) {
	fmt.Printf("Starting Redis container...\n")

	natPort := nat.Port("6379/tcp")
	req := testcontainers.ContainerRequest{
		Image:        "redis:alpine",
		ExposedPorts: []string{string(natPort)},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
		Cmd: []string{
			"redis-server",
			"--protected-mode", "no",
			"--bind", "0.0.0.0",
		},
		HostConfigModifier: func(hc *container.HostConfig) {
			hc.PortBindings = nat.PortMap{
				natPort: []nat.PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: "0",
					},
				},
			}
			hc.Memory = 256 * 1024 * 1024
			hc.CPUShares = 512
			hc.RestartPolicy = container.RestartPolicy{
				Name:              "on-failure",
				MaximumRetryCount: 3,
			}
		},
		Env: map[string]string{
			"REDIS_LOGLEVEL": "debug",
		},
	}

	fmt.Printf("Creating container with config: %+v\n", req)

	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Printf("Failed to create container: %v\n", err)
		return nil, "", errors.NewSystemError("failed to start redis container", err)
	}

	defer func() {
		if err != nil {
			if redisContainer != nil {
				_ = redisContainer.Terminate(ctx)
			}
		}
	}()

	state, err := redisContainer.State(ctx)
	if err != nil {
		return nil, "", errors.NewCacheError("failed to get container state", err)
	}
	fmt.Printf("Container state: %+v\n", state)

	mappedPort, err := redisContainer.MappedPort(ctx, natPort)
	if err != nil {
		fmt.Printf("Failed to get mapped port: %v\n", err)
		return nil, "", errors.NewSystemError("failed to get container port", err)
	}

	hostIP, err := redisContainer.Host(ctx)
	if err != nil {
		fmt.Printf("Failed to get host IP: %v\n", err)
		return nil, "", errors.NewSystemError("failed to get container host", err)
	}
	fmt.Printf("Host IP: %s\n", hostIP)

	redisAddr := fmt.Sprintf("%s:%s", hostIP, mappedPort.Port())
	fmt.Printf("Redis address: %s\n", redisAddr)

	client, err := redis.NewClient(
		redis.WithAddresses([]string{redisAddr}),
		redis.WithDB(0),
		redis.WithPoolSize(10),
		redis.WithMinIdleConns(2),
		redis.WithDialTimeout(5*time.Second),
		redis.WithReadTimeout(2*time.Second),
		redis.WithWriteTimeout(2*time.Second),
		redis.WithMaxRetries(3),
	)
	if err != nil {
		return nil, "", errors.NewCacheError("failed to create redis client", err)
	}

	var pingErr error
	for i := 0; i < 3; i++ {
		if pingErr = client.Ping(ctx); pingErr == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if pingErr != nil {
		client.Close()
		return nil, "", errors.NewCacheError("redis server not ready", pingErr)
	}

	return client, redisAddr, nil
}

func StopRedisContainer(client redis.Client) error {
	if err := client.Close(); err != nil {
		return errors.NewCacheError("failed to close redis client", err)
	}
	return nil
}
