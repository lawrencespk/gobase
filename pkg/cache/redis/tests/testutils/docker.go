package testutils

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// StartRedisContainer 启动Redis容器
func StartRedisContainer() (cleanup func(), err error) {
	// 获取当前文件所在目录
	_, filename, _, _ := runtime.Caller(0)
	dockerComposeFile := filepath.Join(filepath.Dir(filename), "..", "docker-compose.yml")

	// 启动容器
	cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "up", "-d")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to start containers: %v", err)
	}

	// 等待服务就绪
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			// 超时，清理并返回错误
			cleanupFn()
			return nil, fmt.Errorf("timeout waiting for redis to be ready")
		case <-time.After(1 * time.Second):
			// 检查服务是否就绪
			cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "ps", "-q", "redis")
			if out, err := cmd.Output(); err == nil && len(out) > 0 {
				return cleanupFn, nil
			}
		}
	}
}

// cleanupFn 清理函数
func cleanupFn() {
	_, filename, _, _ := runtime.Caller(0)
	dockerComposeFile := filepath.Join(filepath.Dir(filename), "..", "docker-compose.yml")

	cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "down", "-v")
	_ = cmd.Run()
}
