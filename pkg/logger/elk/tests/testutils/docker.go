package testutils

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

// DockerEnvironment 管理测试用的Docker环境
type DockerEnvironment struct {
	projectName string
	composeFile string
}

// NewDockerEnvironment 创建新的Docker环境管理器
func NewDockerEnvironment(projectName, composeFile string) *DockerEnvironment {
	return &DockerEnvironment{
		projectName: projectName,
		composeFile: composeFile,
	}
}

// StartEnvironment 启动Docker环境
func (d *DockerEnvironment) StartEnvironment(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker-compose",
		"-p", d.projectName,
		"-f", d.composeFile,
		"up", "-d")

	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.WrapWithCode(err, codes.ThirdPartyError,
			fmt.Sprintf("failed to start docker environment: %s", string(output)))
	}

	return d.waitForServices(ctx)
}

// StopEnvironment 停止Docker环境
func (d *DockerEnvironment) StopEnvironment(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "docker-compose",
		"-p", d.projectName,
		"-f", d.composeFile,
		"down", "-v")

	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.WrapWithCode(err, codes.ThirdPartyError,
			fmt.Sprintf("failed to stop docker environment: %s", string(output)))
	}

	return nil
}

// waitForServices 等待服务就绪
func (d *DockerEnvironment) waitForServices(ctx context.Context) error {
	// 等待 Elasticsearch 就绪
	for i := 0; i < 30; i++ {
		cmd := exec.CommandContext(ctx, "curl", "-s", "http://localhost:9200/_cluster/health")
		if err := cmd.Run(); err == nil {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return errors.NewError(codes.TimeoutError, "timeout waiting for services to be ready", nil)
}
