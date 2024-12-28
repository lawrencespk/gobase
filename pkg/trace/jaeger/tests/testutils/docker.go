package testutils

import (
	"context"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"gobase/pkg/errors"
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
		return nil, errors.NewSystemError("failed to start jaeger container", err)
	}

	// 返回清理函数
	cleanup := func() {
		cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "down")
		if err := cmd.Run(); err != nil {
			// 这里只记录错误，不返回，因为是清理函数
			errors.NewSystemError("failed to stop jaeger container", err)
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
			return errors.NewSystemError("timeout waiting for jaeger to be ready", nil)
		case <-tick:
			resp, err := http.Get(url)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return nil
				}
			}
		case <-ctx.Done():
			return errors.NewSystemError("context cancelled while waiting for jaeger", ctx.Err())
		}
	}
}
