package config

import (
	"sync"

	"github.com/spf13/viper"

	"gobase/pkg/errors"
)

type Config struct {
	ELK    ELKConfig    `mapstructure:"elk"`
	Logger LoggerConfig `mapstructure:"logger"`
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

	return nil
}
