package testutils

import (
	"context"
	"fmt"
	"gobase/pkg/errors"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

// StartRedisContainer 启动 Redis 容器
func StartRedisContainer(ctx context.Context) (string, error) {
	_, filename, _, _ := runtime.Caller(0)
	dockerComposeFile := filepath.Join(filepath.Dir(filename), "..", "docker-compose.yml")

	// 先清理可能存在的容器
	cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "down", "-v")
	if err := cmd.Run(); err != nil {
		return "", errors.NewSystemError("清理 Redis 容器失败", err)
	}

	// 启动容器
	cmd = exec.Command("docker-compose", "-f", dockerComposeFile, "up", "-d")
	if err := cmd.Run(); err != nil {
		return "", errors.NewSystemError("启动 Redis 容器失败", err)
	}

	// 等待容器健康检查通过
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return "", errors.NewTimeoutError("等待 Redis 容器启动超时", ctx.Err())
		case <-timeout:
			return "", errors.NewTimeoutError("等待 Redis 容器启动超时", nil)
		case <-ticker.C:
			cmd = exec.Command("docker-compose", "-f", dockerComposeFile, "ps", "-q")
			output, err := cmd.Output()
			if err != nil {
				continue
			}
			if len(output) > 0 {
				return "localhost:6379", nil
			}
		}
	}
}

// StopRedisContainer 停止 Redis 容器
func StopRedisContainer() error {
	_, filename, _, _ := runtime.Caller(0)
	dockerComposeFile := filepath.Join(filepath.Dir(filename), "..", "docker-compose.yml")
	cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "down", "-v")
	if err := cmd.Run(); err != nil {
		return errors.NewSystemError("停止 Redis 容器失败", err)
	}
	return nil
}

// StartRedisClusterContainers 启动Redis集群容器
func StartRedisClusterContainers() ([]string, error) {
	// 获取 docker-compose 文件的绝对路径
	_, filename, _, _ := runtime.Caller(0)
	testDir := filepath.Dir(filename)
	composeFile, err := filepath.Abs(filepath.Join(testDir, "../docker-compose-cluster.yml"))
	if err != nil {
		return nil, errors.NewSystemError("无法获取 docker-compose 文件绝对路径: "+err.Error(), err)
	}

	fmt.Printf("Docker Compose 文件路径: %s\n", composeFile)
	fmt.Printf("工作目录: %s\n", testDir)

	// 检查文件是否存在
	if _, err = os.Stat(composeFile); err != nil {
		return nil, errors.NewSystemError("docker-compose 文件不存在: "+composeFile, err)
	}

	// 使用更长的超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 强制停止并删除所有相关容器和网络
	fmt.Println("强制停止并删除所有Redis相关容器...")

	// 1. 停止并删除所有相关容器
	stopCmd := exec.CommandContext(ctx, "docker", "ps", "-aq", "--filter", "name=redis", "--format", "{{.ID}}")
	out, err := stopCmd.CombinedOutput()
	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		containerIDs := strings.Fields(string(out))
		killCmd := exec.CommandContext(ctx, "docker", append([]string{"kill"}, containerIDs...)...)
		killCmd.Run() // 忽略错误

		rmCmd := exec.CommandContext(ctx, "docker", append([]string{"rm", "-f"}, containerIDs...)...)

		rmCmd.Run() // 忽略错误
	}

	// 2. 清理所有相关网络
	fmt.Println("清理所有 Docker 网络...")
	networkCmd := exec.CommandContext(ctx, "docker", "network", "ls", "--filter", "name=redis-cluster-net", "--format", "{{.ID}}")
	out, err = networkCmd.CombinedOutput()
	if err == nil && len(strings.TrimSpace(string(out))) > 0 {
		networkIDs := strings.Fields(string(out))
		for _, id := range networkIDs {
			rmNetCmd := exec.CommandContext(ctx, "docker", "network", "rm", id)
			rmNetCmd.Run() // 忽略错误
		}
	}

	// 3. 使用 docker-compose down 进行清理
	fmt.Println("清理 docker-compose 项目...")
	downCmd := exec.CommandContext(ctx, "docker-compose", "-f", composeFile, "down", "-v", "--remove-orphans")
	if out, err = downCmd.CombinedOutput(); err != nil {
		fmt.Printf("清理 docker-compose 项目时出错: %s\n", string(out))
		// 继续执行，不返回错误
	}

	// 4. 等待一段时间确保资源完全释放
	time.Sleep(2 * time.Second)

	// 修改 docker-compose.yml 中的网络配置
	fmt.Println("修改网络配置...")
	content, err := os.ReadFile(composeFile)
	if err != nil {
		return nil, errors.NewSystemError("读取 docker-compose 文件失败: "+err.Error(), err)
	}

	// 使用随机的网络子网
	rand.Seed(time.Now().UnixNano())
	subnet := fmt.Sprintf("172.%d.0.0/16", rand.Intn(31)+1) // 1-31 避免与常用网段冲突

	// 替换网络配置
	newContent := strings.Replace(string(content),
		"subnet: 172.28.0.0/16",
		fmt.Sprintf("subnet: %s", subnet),
		1)

	// 更新容器IP地址
	baseIP := strings.Split(subnet, ".")[1]
	newContent = strings.Replace(newContent,
		"172.28.0.11",
		fmt.Sprintf("172.%s.0.11", baseIP),
		-1)
	newContent = strings.Replace(newContent,
		"172.28.0.12",
		fmt.Sprintf("172.%s.0.12", baseIP),
		-1)
	newContent = strings.Replace(newContent,
		"172.28.0.13",
		fmt.Sprintf("172.%s.0.13", baseIP),
		-1)

	// 写入临时文件
	tempFile := filepath.Join(testDir, "temp-docker-compose.yml")
	if err := os.WriteFile(tempFile, []byte(newContent), 0644); err != nil {
		return nil, errors.NewSystemError("写入临时配置文件失败: "+err.Error(), err)
	}
	defer os.Remove(tempFile)

	// 使用新的配置文件启动容器
	fmt.Println("正在启动Redis集群容器...")
	startCmd := exec.Command("docker-compose", "-f", tempFile, "--project-directory", testDir, "up", "-d", "--force-recreate", "--remove-orphans")
	startCmd.Dir = testDir
	out, err = startCmd.CombinedOutput()
	fmt.Printf("启动输出:\n%s\n", string(out))
	if err != nil {
		return nil, errors.NewSystemError(fmt.Sprintf("启动 Redis 集群容器失败: %v\n输出: %s", err, string(out)), err)
	}

	// 等待容器启动和集群初始化
	fmt.Println("等待容器启动和集群初始化...")
	maxRetries := 30
	var clusterInfoOut []byte
	var clusterNodesOut []byte

	for i := 0; i < maxRetries; i++ {
		time.Sleep(2 * time.Second)

		// 检查集群状态
		clusterInfoCmd := exec.Command("docker", "exec", "redis-node-1", "redis-cli", "-p", "7001", "cluster", "info")
		clusterInfoOut, err = clusterInfoCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("检查集群状态失败 (尝试 %d/%d): %s\n", i+1, maxRetries, string(clusterInfoOut))
			continue
		}

		// 检查集群节点
		clusterNodesCmd := exec.Command("docker", "exec", "redis-node-1", "redis-cli", "-p", "7001", "cluster", "nodes")
		clusterNodesOut, err = clusterNodesCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("检查集群节点失败 (尝试 %d/%d): %s\n", i+1, maxRetries, string(clusterNodesOut))
			continue
		}

		// 检查是否所有节点都已就绪
		if strings.Contains(string(clusterNodesOut), "connected") {
			fmt.Println("集群初始化完成")

			// 打印集群信息
			fmt.Println("\n集群状态:")
			fmt.Println(string(clusterInfoOut))
			fmt.Println("\n集群节点:")
			fmt.Println(string(clusterNodesOut))

			break
		}

		fmt.Printf("等待集群初始化 (尝试 %d/%d)...\n", i+1, maxRetries)
	}

	// 等待集群完全稳定
	time.Sleep(5 * time.Second)

	// 更新客户端配置以使用正确的地址
	fmt.Println("获取集群节点地址...")
	getNodesCmd := exec.Command("docker", "exec", "redis-node-1", "redis-cli", "-p", "7001", "cluster", "nodes")
	out, err = getNodesCmd.CombinedOutput()
	if err != nil {
		return nil, errors.NewSystemError("获取集群节点信息失败: "+string(out), err)
	}

	// 解析节点地址
	var addrs []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.Contains(line, "master") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				addr := fields[1]
				// 处理地址格式：172.25.0.11:7001@17001 -> localhost:7001
				if strings.Contains(addr, ":") {
					parts := strings.Split(addr, "@")[0] // 首先移除 @17001 部分
					hostPort := strings.Split(parts, ":")
					if len(hostPort) == 2 {
						// 使用 localhost 和端口映射
						addr = fmt.Sprintf("localhost:%s", hostPort[1])
						addrs = append(addrs, addr)
					}
				}
			}
		}
	}

	if len(addrs) == 0 {
		return nil, errors.NewSystemError("未找到可用的集群节点", nil)
	}

	// 确保地址列表是固定顺序的
	sort.Strings(addrs)

	fmt.Printf("集群节点地址: %v\n", addrs)
	return addrs, nil
}

// StopRedisClusterContainers 停止Redis集群容器
func StopRedisClusterContainers() error {
	_, filename, _, _ := runtime.Caller(0)
	dockerComposeFile := filepath.Join(filepath.Dir(filename), "..", "docker-compose-cluster.yml")
	cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "down", "-v")
	if err := cmd.Run(); err != nil {
		return errors.NewSystemError("停止 Redis 集群容器失败", err)
	}
	return nil
}

// WaitForRedisCluster 等待Redis集群就绪
func WaitForRedisCluster(t *testing.T) {
	maxRetries := 30
	retryInterval := time.Second

	for i := 0; i < maxRetries; i++ {
		// 尝试连接到集群的任一节点
		client := redis.NewClient(&redis.Options{
			Addr: "172.25.0.11:7001",
		})

		_, err := client.Ping(context.Background()).Result()
		if err == nil {
			client.Close()
			return
		}

		client.Close()
		time.Sleep(retryInterval)
	}

	t.Fatal("Redis cluster failed to start")
}

// StartRedisSingleContainer 启动单节点 Redis 容器
func StartRedisSingleContainer() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return StartRedisContainer(ctx)
}

// CleanupRedisContainers 清理 Redis 容器
func CleanupRedisContainers() error {
	return StopRedisContainer()
}

// RestartRedisContainer 重启 Redis 容器
func RestartRedisContainer() error {
	_, filename, _, _ := runtime.Caller(0)
	dockerComposeFile := filepath.Join(filepath.Dir(filename), "..", "docker-compose.yml")

	// 执行 docker-compose restart
	cmd := exec.Command("docker-compose", "-f", dockerComposeFile, "restart", "redis")
	if out, err := cmd.CombinedOutput(); err != nil {
		return errors.NewSystemError(fmt.Sprintf("重启 Redis 容器失败: %s", string(out)), err)
	}

	// 等待容器重启完成
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.NewTimeoutError("等待 Redis 容器重启超时", ctx.Err())
		case <-ticker.C:
			// 检查容器状态
			cmd = exec.Command("docker-compose", "-f", dockerComposeFile, "ps", "-q", "redis")
			output, err := cmd.Output()
			if err != nil {
				continue
			}

			if len(output) > 0 {
				// 给容器一些额外时间来完全初始化
				time.Sleep(2 * time.Second)
				return nil
			}
		}
	}
}
