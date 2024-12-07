package integration

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/config"
	"gobase/pkg/config/types"
)

// TestConfig 测试配置结构
type TestConfig struct {
	App struct {
		Server struct {
			Host string `json:"host"`
			Port int    `json:"port"`
		} `json:"server"`
		Database struct {
			Host     string        `json:"host"`
			Port     int           `json:"port"`
			User     string        `json:"user"`
			Password string        `json:"password"`
			DBName   string        `json:"dbname"`
			Timeout  time.Duration `json:"timeout"`
		} `json:"database"`
	} `json:"app"`
}

func TestNewManager(t *testing.T) {
	tests := []struct {
		name    string
		opts    []types.ConfigOption
		wantErr bool
	}{
		{
			name: "default config",
			opts: []types.ConfigOption{
				types.WithConfigFile("../../../../config/config.yaml"),
			},
			wantErr: false,
		},
		{
			name: "file not found",
			opts: []types.ConfigOption{
				types.WithConfigFile("not_exists.yaml"),
			},
			wantErr: true,
		},
		{
			name: "with environment",
			opts: []types.ConfigOption{
				types.WithConfigFile("../../../../config/config.yaml"),
				types.WithEnvironment("test"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := config.NewManager(tt.opts...)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, manager)
			defer manager.Close()
		})
	}
}

func TestManager_GetValues_Integration(t *testing.T) {
	// 创建配置管理器，指定配置文件路径
	manager, err := config.NewManager(func(options *types.ConfigOptions) {
		options.ConfigFile = "../../../../config/config.yaml" // 从 pkg/config/tests/integration 到项目根目录
	})
	require.NoError(t, err)
	defer manager.Close()

	t.Run("get app info", func(t *testing.T) {
		assert.Equal(t, "gobase", manager.GetString("app.name"))
		assert.Equal(t, "1.0.0", manager.GetString("app.version"))
		assert.Equal(t, "development", manager.GetString("app.mode"))
	})

	t.Run("get server config", func(t *testing.T) {
		assert.Equal(t, "0.0.0.0", manager.GetString("server.host"))
		assert.Equal(t, 8080, manager.GetInt("server.port"))
		assert.Equal(t, 30*time.Second, manager.GetDuration("server.timeout"))
	})

	t.Run("get database config", func(t *testing.T) {
		assert.Equal(t, "104.238.161.243", manager.GetString("database.postgres.host"))
		assert.Equal(t, 9187, manager.GetInt("database.postgres.port"))
		assert.Equal(t, "stagingserv", manager.GetString("database.postgres.user"))
	})
}

func TestManager_Parse_Integration(t *testing.T) {
	manager, err := config.NewManager(func(options *types.ConfigOptions) {
		options.ConfigFile = "../../../../config/config.yaml"
	})
	require.NoError(t, err)
	defer manager.Close()

	t.Run("parse server config", func(t *testing.T) {
		type ServerConfig struct {
			Host    string         `json:"host"`
			Port    int            `json:"port"`
			Timeout types.Duration `json:"timeout"`
		}

		var config ServerConfig
		err := manager.Parse("server", &config)
		require.NoError(t, err)
		assert.Equal(t, "0.0.0.0", config.Host)
		assert.Equal(t, 8080, config.Port)
		assert.Equal(t, 30*time.Second, config.Timeout.Duration)
	})
}

func TestManager_WithEnv(t *testing.T) {
	// 设置测试环境变量
	os.Setenv("GOBASE_APP_SERVER_PORT", "9090")
	defer os.Unsetenv("GOBASE_APP_SERVER_PORT")

	// 创建测试配置管理器
	manager, err := config.NewManager(
		types.WithConfigFile("../../../../config/config.yaml"),
		types.WithEnableEnv(true),
		types.WithEnvPrefix("GOBASE"),
	)
	require.NoError(t, err)
	defer manager.Close()

	t.Run("override with env", func(t *testing.T) {
		value := manager.GetInt("app.server.port")
		assert.Equal(t, 9090, value)
	})
}

func TestManager_Watch(t *testing.T) {
	manager, err := config.NewManager(func(options *types.ConfigOptions) {
		options.ConfigFile = "../../../../config/config.yaml"
	})
	require.NoError(t, err)
	defer manager.Close()

	t.Run("watch config change", func(t *testing.T) {
		done := make(chan bool, 1)
		var newValue interface{}

		err := manager.Watch("server.port", func(key string, value interface{}) {
			newValue = value
			done <- true
		})
		require.NoError(t, err)

		go manager.Set("server.port", 8081)

		select {
		case <-done:
			assert.Equal(t, 8081, newValue)
		case <-time.After(100 * time.Millisecond):
			t.Fatal("watch timeout")
		}
	})

	time.Sleep(100 * time.Millisecond)
}
