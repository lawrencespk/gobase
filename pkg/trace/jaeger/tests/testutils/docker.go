package testutils

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// StartJaegerContainer 启动 Jaeger 容器并返回清理函数
func StartJaegerContainer(ctx context.Context) (func(), error) {
	// 获取当前文件所在目录
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filepath.Dir(filename))
	dockerComposeFile := filepath.Join(testDir, "integration", "docker-compose.yml")

	// 启动 docker-compose
	cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "up", "-d")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to start jaeger container: %v", err)
	}

	// 返回清理函数
	cleanup := func() {
		cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "down")
		if err := cmd.Run(); err != nil {
			fmt.Printf("failed to stop jaeger container: %v\n", err)
		}
	}

	return cleanup, nil
}

// WaitForJaeger 等待 Jaeger 服务就绪
func WaitForJaeger(ctx context.Context) error {
	// 最多等待30秒
	timeout := time.After(30 * time.Second)
	tick := time.Tick(1 * time.Second)

	// 检查 Jaeger UI 端口
	url := "http://localhost:16686"

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for jaeger to be ready")
		case <-tick:
			resp, err := http.Get(url)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return nil
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
