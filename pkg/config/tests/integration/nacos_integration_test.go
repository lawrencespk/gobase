package integration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gobase/pkg/config/source/nacos"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// NacosConfig 定义配置结构体
type NacosConfig struct {
	Endpoints     []string
	Namespace     string
	Group         string
	DataID        string
	Username      string
	Password      string
	TimeoutMs     uint64
	LogDir        string
	CacheDir      string
	LogLevel      string
	Scheme        string
	ConfigType    string
	RetryTimes    int
	RetryInterval time.Duration
}

// validateConfig 验证配置
func validateConfig(cfg *NacosConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if len(cfg.Endpoints) == 0 {
		return fmt.Errorf("endpoints is required")
	}
	if cfg.DataID == "" {
		return fmt.Errorf("dataId is required")
	}
	if cfg.Group == "" {
		return fmt.Errorf("group is required")
	}
	return nil
}

// isConfigNotExistError 判断是否为配置不存在错误
func isConfigNotExistError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "config data not exist")
}

// isClientNotConnectedError 判断是否为客户端未连接错误
func isClientNotConnectedError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "client not connected")
}

func TestNacosIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 加载并验证配置
	cfg := loadNacosConfig(t)
	if err := validateConfig(cfg); err != nil {
		t.Fatalf("Invalid config: %v", err)
	}

	// 使用系统临时目录而不是测试临时目录
	tempDir := os.TempDir()
	logDir := filepath.Join(tempDir, fmt.Sprintf("nacos-test-log-%d", time.Now().UnixNano()))
	cacheDir := filepath.Join(tempDir, fmt.Sprintf("nacos-test-cache-%d", time.Now().UnixNano()))

	require.NoError(t, os.MkdirAll(logDir, 0755))
	require.NoError(t, os.MkdirAll(cacheDir, 0755))

	t.Logf("Created temp directories - log: %s, cache: %s", logDir, cacheDir)

	// 创建客户端
	client := setupNacosClient(t, cfg, logDir, cacheDir)

	// 首先发布初始配置
	testConfig := `
test:
  key: value
  nested:
    key: nestedValue
`
	success, err := client.PublishConfig(vo.ConfigParam{
		DataId:  cfg.DataID,
		Group:   cfg.Group,
		Content: testConfig,
		Type:    cfg.ConfigType,
	})
	require.NoError(t, err)
	require.True(t, success, "Failed to publish initial test config")
	t.Log("Published initial test config")

	// 等待配置生效
	time.Sleep(time.Second)

	// 然后等待客户端就绪
	if err := waitForClientReady(t, client, cfg); err != nil {
		t.Fatalf("Failed to wait for client ready: %v", err)
	}
	t.Log("Client is ready")

	// 创建 NacosSource
	opts := &nacos.Options{
		Endpoints:     cfg.Endpoints,
		NamespaceID:   cfg.Namespace,
		Group:         cfg.Group,
		DataID:        cfg.DataID,
		Username:      cfg.Username,
		Password:      cfg.Password,
		TimeoutMs:     cfg.TimeoutMs,
		LogDir:        cfg.LogDir,
		CacheDir:      cfg.CacheDir,
		LogLevel:      cfg.LogLevel,
		Scheme:        cfg.Scheme,
		ConfigType:    cfg.ConfigType,
		RetryTimes:    cfg.RetryTimes,
		RetryInterval: cfg.RetryInterval,
	}

	source, err := nacos.NewSourceWithClient(client, opts)
	require.NoError(t, err)

	// 修改清理逻辑
	cleanup := func() {
		t.Log("Starting cleanup...")

		// 1. 停止所有活动操作
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := source.StopWatch(ctx); err != nil {
			t.Logf("Warning: Failed to stop watch: %v", err)
		}

		// 2. 关闭 source（包括客户端和日志）
		if err := source.Close(ctx); err != nil {
			t.Logf("Warning: Failed to close source: %v", err)
		}

		// 3. 等待资源释放
		time.Sleep(time.Second)

		// 4. 删除配置（使用新的临时客户端）
		if err := deleteConfig(t, cfg); err != nil {
			t.Logf("Warning: Failed to delete config: %v", err)
		}

		t.Log("Cleanup completed")
	}

	t.Cleanup(cleanup)

	// 运行子测试
	t.Run("ConcurrentOperations", func(t *testing.T) {
		testConcurrentOperations(t, client, cfg)
	})

	time.Sleep(time.Second)

	t.Run("ErrorRecovery", func(t *testing.T) {
		testErrorRecovery(t, client, cfg)
	})

	// 在测试结束前关闭日志文件
	t.Cleanup(func() {
		t.Log("Starting cleanup...")

		// 1. 只保留关闭客户端
		if client != nil {
			client.CloseClient()
		}

		t.Log("Cleanup completed")
	})
}

// waitForClientReady 等待客户端就绪
func waitForClientReady(t *testing.T, client config_client.IConfigClient, cfg *NacosConfig) error {
	t.Helper()

	maxRetries := 20
	retryInterval := 500 * time.Millisecond

	for i := 1; i <= maxRetries; i++ {
		t.Logf("Attempting to connect to Nacos server (attempt %d/%d)", i, maxRetries)

		content, err := client.GetConfig(vo.ConfigParam{
			DataId: cfg.DataID,
			Group:  cfg.Group,
		})

		if err == nil {
			t.Logf("Successfully connected to Nacos server, config content: %s", content)
			return nil
		}

		// 使用错误类型检查函数
		if isConfigNotExistError(err) {
			t.Logf("Connection attempt failed: %v", err)
			if i == maxRetries {
				return fmt.Errorf("timeout waiting for client to be ready after %d retries", i-1)
			}
			time.Sleep(retryInterval)
			continue
		}

		if isClientNotConnectedError(err) {
			t.Logf("Client not connected: %v", err)
			time.Sleep(retryInterval)
			continue
		}

		return fmt.Errorf("failed to connect: %v", err)
	}

	return fmt.Errorf("failed to connect after maximum retries")
}

// loadNacosConfig 加载测试配置
func loadNacosConfig(t *testing.T) *NacosConfig {
	t.Helper()

	cfg := &NacosConfig{
		Endpoints:     []string{"127.0.0.1:8848"},
		Namespace:     "public",
		Group:         "DEFAULT_GROUP",
		DataID:        fmt.Sprintf("test-config-%d", time.Now().UnixNano()),
		Username:      "nacos",
		Password:      "nacos",
		TimeoutMs:     5000,
		LogLevel:      "debug",
		Scheme:        "http",
		ConfigType:    "yaml",
		RetryTimes:    3,
		RetryInterval: time.Second,
	}

	t.Logf("Created test config: %+v", cfg)
	return cfg
}

// setupNacosClient 设置Nacos客户端
func setupNacosClient(t *testing.T, cfg *NacosConfig, logDir, cacheDir string) config_client.IConfigClient {
	t.Helper()

	clientConfig := constant.ClientConfig{
		NamespaceId:         cfg.Namespace,
		TimeoutMs:           cfg.TimeoutMs,
		NotLoadCacheAtStart: true,
		LogDir:              logDir,
		CacheDir:            cacheDir,
		LogLevel:            "debug",
		Username:            cfg.Username,
		Password:            cfg.Password,
		ContextPath:         "/nacos",
		LogRollingConfig: &constant.ClientLogRollingConfig{
			MaxAge:     1,  // 日志最大保存时间(天)
			MaxSize:    10, // 日志文件最大大小(MB)
			MaxBackups: 2,  // 最大备份数
		},
	}

	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      "127.0.0.1",
			Port:        8848,
			ContextPath: "/nacos",
			Scheme:      "http",
		},
	}

	client, err := clients.CreateConfigClient(map[string]interface{}{
		"serverConfigs": serverConfigs,
		"clientConfig":  clientConfig,
	})
	require.NoError(t, err)
	return client
}

func deleteConfig(t *testing.T, cfg *NacosConfig) error {
	// 创建临时客户端来删除配置
	client := setupNacosClient(t, cfg, "", "")
	defer client.CloseClient()

	maxRetries := 3
	retryInterval := 500 * time.Millisecond

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		success, err := client.DeleteConfig(vo.ConfigParam{
			DataId: cfg.DataID,
			Group:  cfg.Group,
		})

		if err == nil && success {
			t.Logf("Successfully deleted config after %d attempts", i+1)
			return nil
		}

		lastErr = err
		if i < maxRetries-1 {
			t.Logf("Retry %d: Failed to delete config: %v", i+1, err)
			// 如果是客户端关闭错误，等待更长时间
			if isClientNotConnectedError(err) {
				time.Sleep(retryInterval * 2)
			} else {
				time.Sleep(retryInterval)
			}
		}
	}

	return fmt.Errorf("failed to delete config after %d retries: %v", maxRetries, lastErr)
}

// testConcurrentOperations 并发操作测试
func testConcurrentOperations(t *testing.T, client config_client.IConfigClient, cfg *NacosConfig) {
	done := make(chan struct{})
	go func() {
		defer close(done)

		var wg sync.WaitGroup
		concurrency := 10
		operations := 10

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				for j := 0; j < operations; j++ {
					content := fmt.Sprintf("value-%d-%d", index, j)
					success, err := client.PublishConfig(vo.ConfigParam{
						DataId:  cfg.DataID,
						Group:   cfg.Group,
						Content: content,
					})
					require.NoError(t, err)
					require.True(t, success)
					time.Sleep(50 * time.Millisecond)
				}
			}(i)
		}
		wg.Wait()
	}()

	select {
	case <-done:
		t.Log("Concurrent operations completed successfully")
	case <-time.After(30 * time.Second):
		t.Fatal("Concurrent operations test timeout")
	}
}

// testErrorRecovery 错误恢复测试
func testErrorRecovery(t *testing.T, client config_client.IConfigClient, cfg *NacosConfig) {
	testValue := "test-error-recovery"

	// 发布有效配置
	success, err := client.PublishConfig(vo.ConfigParam{
		DataId:  cfg.DataID,
		Group:   cfg.Group,
		Content: fmt.Sprintf("test:\n  key: %s", testValue),
		Type:    cfg.ConfigType,
	})
	require.NoError(t, err)
	require.True(t, success)

	time.Sleep(time.Second)
	t.Log("Published valid config")

	// 验证配置
	content, err := client.GetConfig(vo.ConfigParam{
		DataId: cfg.DataID,
		Group:  cfg.Group,
	})
	require.NoError(t, err)

	var config map[string]interface{}
	err = yaml.Unmarshal([]byte(content), &config)
	require.NoError(t, err)

	testMap, ok := config["test"].(map[string]interface{})
	require.True(t, ok)
	actualValue, ok := testMap["key"].(string)
	require.True(t, ok)
	require.Equal(t, testValue, actualValue)

	t.Log("Verified valid config")

	// 发布无效配置
	_, err = client.PublishConfig(vo.ConfigParam{
		DataId:  cfg.DataID,
		Group:   cfg.Group,
		Content: "invalid yaml: :",
		Type:    cfg.ConfigType,
	})
	require.NoError(t, err)

	time.Sleep(time.Second)
	t.Log("Published invalid config")

	// 恢复有效配置
	success, err = client.PublishConfig(vo.ConfigParam{
		DataId:  cfg.DataID,
		Group:   cfg.Group,
		Content: fmt.Sprintf("test:\n  key: %s", testValue),
		Type:    cfg.ConfigType,
	})
	require.NoError(t, err)
	require.True(t, success)

	time.Sleep(time.Second)
	t.Log("Recovered valid config")

	// 验证恢复
	content, err = client.GetConfig(vo.ConfigParam{
		DataId: cfg.DataID,
		Group:  cfg.Group,
	})
	require.NoError(t, err)

	err = yaml.Unmarshal([]byte(content), &config)
	require.NoError(t, err)

	testMap, ok = config["test"].(map[string]interface{})
	require.True(t, ok)
	actualValue, ok = testMap["key"].(string)
	require.True(t, ok)
	require.Equal(t, testValue, actualValue)

	t.Log("Verified recovered config")
}
