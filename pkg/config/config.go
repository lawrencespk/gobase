package config

import (
	"context"
	"sync"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"gobase/pkg/config/types"
	"gobase/pkg/errors"
)

type Config struct {
	ELK    ELKConfig          `mapstructure:"elk" yaml:"elk"`
	Logger LoggerConfig       `mapstructure:"logger" yaml:"logger"`
	Jaeger types.JaegerConfig `mapstructure:"jaeger" yaml:"jaeger"`
}

type ELKConfig struct {
	Addresses []string   `mapstructure:"addresses"`
	Username  string     `mapstructure:"username"`
	Password  string     `mapstructure:"password"`
	Index     string     `mapstructure:"index"`
	Timeout   int        `mapstructure:"timeout"`
	Bulk      BulkConfig `mapstructure:"bulk"`
}

type BulkConfig struct {
	BatchSize  int    `mapstructure:"batchSize"`
	FlushBytes int    `mapstructure:"flushBytes"`
	Interval   string `mapstructure:"interval"`
}

type LoggerConfig struct {
	Development      bool           `mapstructure:"development"`
	Level            string         `mapstructure:"level"`
	Format           string         `mapstructure:"format"`
	ReportCaller     bool           `mapstructure:"reportCaller"`
	TimeFormat       string         `mapstructure:"timeFormat"`
	DisableConsole   bool           `mapstructure:"disableConsole"`
	OutputPaths      []string       `mapstructure:"outputPaths"`
	ErrorOutputPaths []string       `mapstructure:"errorOutputPaths"`
	MaxAge           string         `mapstructure:"maxAge"`
	RotationTime     string         `mapstructure:"rotationTime"`
	MaxSize          int64          `mapstructure:"maxSize"`
	Rotation         RotationConfig `mapstructure:"rotation"`
	Compress         CompressConfig `mapstructure:"compress"`
	Cleanup          CleanupConfig  `mapstructure:"cleanup"`
	Async            AsyncConfig    `mapstructure:"async"`
	Recovery         RecoveryConfig `mapstructure:"recovery"`
	Queue            QueueConfig    `mapstructure:"queue"`
	Elk              struct {
		Enable bool `mapstructure:"enable"`
	} `mapstructure:"elk"`
	Output string `mapstructure:"output" yaml:"output"`
}

type RotationConfig struct {
	Enable       bool   `mapstructure:"enable"`
	Filename     string `mapstructure:"filename"`
	MaxSize      int    `mapstructure:"maxSize"`
	MaxAge       int    `mapstructure:"maxAge"`
	RotationTime string `mapstructure:"rotationTime"`
	MaxBackups   int    `mapstructure:"maxBackups"`
	Compress     bool   `mapstructure:"compress"`
}

type CompressConfig struct {
	Enable       bool   `mapstructure:"enable"`
	Algorithm    string `mapstructure:"algorithm"`
	Level        int    `mapstructure:"level"`
	DeleteSource bool   `mapstructure:"deleteSource"`
	Interval     string `mapstructure:"interval"`
}

type CleanupConfig struct {
	Enable     bool   `mapstructure:"enable"`
	MaxBackups int    `mapstructure:"maxBackups"`
	MaxAge     int    `mapstructure:"maxAge"`
	Interval   string `mapstructure:"interval"`
}

type AsyncConfig struct {
	Enable        bool   `mapstructure:"enable"`
	BufferSize    int    `mapstructure:"bufferSize"`
	FlushInterval string `mapstructure:"flushInterval"`
	BlockOnFull   bool   `mapstructure:"blockOnFull"`
	DropOnFull    bool   `mapstructure:"dropOnFull"`
	FlushOnExit   bool   `mapstructure:"flushOnExit"`
}

type RecoveryConfig struct {
	Enable           bool   `mapstructure:"enable"`
	MaxRetries       int    `mapstructure:"maxRetries"`
	RetryInterval    string `mapstructure:"retryInterval"`
	EnableStackTrace bool   `mapstructure:"enableStackTrace"`
	MaxStackSize     int    `mapstructure:"maxStackSize"`
}

type QueueConfig struct {
	MaxSize         int    `mapstructure:"maxSize"`
	BatchSize       int    `mapstructure:"batchSize"`
	Workers         int    `mapstructure:"workers"`
	FlushInterval   string `mapstructure:"flushInterval"`
	RetryCount      int    `mapstructure:"retryCount"`
	RetryInterval   string `mapstructure:"retryInterval"`
	MaxBatchWait    string `mapstructure:"maxBatchWait"`
	ShutdownTimeout string `mapstructure:"shutdownTimeout"`
}

var (
	globalConfig *Config
	configMu     sync.RWMutex
	configFile   string
)

func Init() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("configs")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return err
	}

	globalConfig = config
	return nil
}

func GetConfig() *Config {
	configMu.RLock()
	defer configMu.RUnlock()
	return globalConfig
}

func SetConfig(cfg *Config) {
	configMu.Lock()
	globalConfig = cfg
	configMu.Unlock()
}

// LoadConfig 加载配置文件
func LoadConfig() error {
	if configFile != "" {
		// 如果设置了具体的配置文件路径，直接使用
		viper.SetConfigFile(configFile)
	} else {
		// 默认配置
		viper.SetConfigName("config") // 配置文件名(不带扩展名)
		viper.SetConfigType("yaml")   // 配置文件类型
		viper.AddConfigPath("config") // 配置文件路径
	}

	if err := viper.ReadInConfig(); err != nil {
		return errors.NewConfigError("failed to read config file", err)
	}

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return errors.NewConfigError("failed to unmarshal config", err)
	}

	// 验证配置
	if err := ValidateConfig(config); err != nil {
		return err // ValidateConfig 已经返回了正确的错误类型
	}

	SetConfig(config)
	return nil
}

// SetConfigPath 设置配置文件路径
func SetConfigPath(path string) {
	configFile = path
}

// ValidateConfig 验证配置
func ValidateConfig(cfg *Config) error {
	if len(cfg.ELK.Addresses) == 0 {
		return errors.NewConfigError("elk addresses is empty", nil)
	}
	if cfg.ELK.Username == "" {
		return errors.NewConfigError("elk username is empty", nil)
	}
	if cfg.ELK.Password == "" {
		return errors.NewConfigError("elk password is empty", nil)
	}
	if cfg.ELK.Index == "" {
		return errors.NewConfigError("elk index is empty", nil)
	}
	if cfg.ELK.Timeout <= 0 {
		return errors.NewConfigError("elk timeout must be greater than 0", nil)
	}

	// 验证Bulk配置
	if cfg.ELK.Bulk.BatchSize <= 0 {
		return errors.NewConfigError("elk bulk batch size must be greater than 0", nil)
	}
	if cfg.ELK.Bulk.FlushBytes <= 0 {
		return errors.NewConfigError("elk bulk flush bytes must be greater than 0", nil)
	}
	if cfg.ELK.Bulk.Interval == "" {
		return errors.NewConfigError("elk bulk interval is empty", nil)
	}

	// 添加 Jaeger 配置验证
	if cfg.Jaeger.Enable {
		if cfg.Jaeger.ServiceName == "" {
			return errors.NewConfigError("jaeger service name is empty", nil)
		}
		if cfg.Jaeger.Agent.Host == "" {
			return errors.NewConfigError("jaeger agent host is empty", nil)
		}
		if cfg.Jaeger.Agent.Port == "" {
			return errors.NewConfigError("jaeger agent port is empty", nil)
		}
		if cfg.Jaeger.Collector.Endpoint == "" {
			return errors.NewConfigError("jaeger collector endpoint is empty", nil)
		}
	}

	return nil
}

// NewConfig 创建新的配置实例并设置默认值
func NewConfig() *Config {
	return &Config{
		ELK: ELKConfig{
			Timeout: 30,
			Bulk: BulkConfig{
				BatchSize:  1000,
				FlushBytes: 5 * 1024 * 1024, // 5MB
				Interval:   "5s",
			},
		},
		Logger: LoggerConfig{
			Level:  "info",
			Output: "console",
		},
	}
}

// Clone 深拷贝配置
func (c *Config) Clone() *Config {
	if c == nil {
		return nil
	}

	copied := &Config{
		ELK: ELKConfig{
			Addresses: make([]string, len(c.ELK.Addresses)),
			Username:  c.ELK.Username,
			Password:  c.ELK.Password,
			Index:     c.ELK.Index,
			Timeout:   c.ELK.Timeout,
			Bulk: BulkConfig{
				BatchSize:  c.ELK.Bulk.BatchSize,
				FlushBytes: c.ELK.Bulk.FlushBytes,
				Interval:   c.ELK.Bulk.Interval,
			},
		},
		Logger: LoggerConfig{
			Level:  c.Logger.Level,
			Output: c.Logger.Output,
		},
	}

	// 深拷贝切片
	copy(copied.ELK.Addresses, c.ELK.Addresses)

	return copied
}

// Merge 合并配置
func (c *Config) Merge(other *Config) *Config {
	if other == nil {
		return c.Clone()
	}

	merged := c.Clone()

	// 合并 ELK 配置
	if len(other.ELK.Addresses) > 0 {
		merged.ELK.Addresses = make([]string, len(other.ELK.Addresses))
		copy(merged.ELK.Addresses, other.ELK.Addresses)
	}
	if other.ELK.Username != "" {
		merged.ELK.Username = other.ELK.Username
	}
	if other.ELK.Password != "" {
		merged.ELK.Password = other.ELK.Password
	}
	if other.ELK.Index != "" {
		merged.ELK.Index = other.ELK.Index
	}
	if other.ELK.Timeout > 0 {
		merged.ELK.Timeout = other.ELK.Timeout
	}

	// 合并 Bulk 配置
	if other.ELK.Bulk.BatchSize > 0 {
		merged.ELK.Bulk.BatchSize = other.ELK.Bulk.BatchSize
	}
	if other.ELK.Bulk.FlushBytes > 0 {
		merged.ELK.Bulk.FlushBytes = other.ELK.Bulk.FlushBytes
	}
	if other.ELK.Bulk.Interval != "" {
		merged.ELK.Bulk.Interval = other.ELK.Bulk.Interval
	}

	// 合并 Logger 配置
	if other.Logger.Level != "" {
		merged.Logger.Level = other.Logger.Level
	}
	if other.Logger.Output != "" {
		merged.Logger.Output = other.Logger.Output
	}

	return merged
}

// MarshalYAML 序列化配置到YAML
func MarshalYAML(cfg *Config) ([]byte, error) {
	if cfg == nil {
		return nil, errors.NewConfigError("config is nil", nil)
	}
	return yaml.Marshal(cfg)
}

// UnmarshalYAML 从YAML反序列化配置
func UnmarshalYAML(data []byte, cfg *Config) error {
	if cfg == nil {
		return errors.NewConfigError("config is nil", nil)
	}
	return yaml.Unmarshal(data, cfg)
}

// Watch 监听配置变更
func Watch(ctx context.Context, cfg *types.Config, onChange func()) error {
	// 实现配置监听逻辑
	// 可以使用 viper 的 WatchConfig 功能
	return nil
}

// ToTypesConfig 将 *Config 转换为 *types.Config
func (c *Config) ToTypesConfig() *types.Config {
	if c == nil {
		return nil
	}
	return &types.Config{
		Jaeger: c.Jaeger,
		Grafana: types.GrafanaConfig{
			Dashboards: struct {
				HTTP      string `json:"http" yaml:"http"`
				Logger    string `json:"logger" yaml:"logger"`
				Runtime   string `json:"runtime" yaml:"runtime"`
				System    string `json:"system" yaml:"system"`
				Redis     string `json:"redis" yaml:"redis"`
				RateLimit string `json:"rate_limit" yaml:"rate_limit"`
			}{},
			Alerts: struct {
				Rules     string `json:"rules" yaml:"rules"`
				Logger    string `json:"logger" yaml:"logger"`
				Redis     string `json:"redis" yaml:"redis"`
				HTTP      string `json:"http" yaml:"http"`
				Runtime   string `json:"runtime" yaml:"runtime"`
				RateLimit string `json:"rate_limit" yaml:"rate_limit"`
			}{},
		},
	}
}
