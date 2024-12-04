package config

import (
	"github.com/spf13/viper"
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

var globalConfig *Config

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
	return globalConfig
}
