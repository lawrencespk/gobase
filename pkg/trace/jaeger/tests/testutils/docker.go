package testutils

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// StartJaegerContainer 启动 Jaeger 容器
func StartJaegerContainer() error {
	// 获取当前文件所在目录
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filepath.Dir(filename))
	dockerComposeFile := filepath.Join(testDir, "integration", "docker-compose.yml")

	// 启动 docker-compose
	cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "up", "-d")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start jaeger container: %v", err)
	}

	// 等待服务就绪
	time.Sleep(5 * time.Second)
	return nil
}

// StopJaegerContainer 停止并清理 Jaeger 容器
func StopJaegerContainer() error {
	// 获取当前文件所在目录
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filepath.Dir(filename))
	dockerComposeFile := filepath.Join(testDir, "integration", "docker-compose.yml")

	// 停止并删除容器
	cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "down")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop jaeger container: %v", err)
	}

	return nil
}
