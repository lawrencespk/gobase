package elk

import (
	"gobase/pkg/config"
	"time"

	"github.com/sirupsen/logrus"
)

// ElkConfig 定义与 Elasticsearch 连接的配置
type ElkConfig struct {
	Addresses []string
	Username  string
	Password  string
	Index     string
	Timeout   time.Duration
}

// DefaultElkConfig 返回默认配置
func DefaultElkConfig() *ElkConfig {
	defaultConfig := &ElkConfig{
		Addresses: []string{"http://localhost:9200"},
		Username:  "",
		Password:  "",
		Index:     "default-index",
		Timeout:   30 * time.Second,
	}

	conf := config.GetConfig()
	if conf == nil || len(conf.ELK.Addresses) == 0 {
		return defaultConfig
	}

	return &ElkConfig{
		Addresses: conf.ELK.Addresses,
		Username:  conf.ELK.Username,
		Password:  conf.ELK.Password,
		Index:     conf.ELK.Index,
		Timeout:   time.Duration(conf.ELK.Timeout) * time.Second,
	}
}

// Config ELK配置
type Config struct {
	// Elasticsearch配置
	Elasticsearch ElkConfig `yaml:"elasticsearch"`
	// Hook配置
	Hook struct {
		Enabled bool        `yaml:"enabled"`
		Levels  []string    `yaml:"levels"`
		Index   string      `yaml:"index"`
		Batch   BatchConfig `yaml:"batch"`
		Retry   RetryConfig `yaml:"retry"`
	} `yaml:"hook"`
}

// BatchConfig 批处理配置
type BatchConfig struct {
	Size       int           `yaml:"size"`
	FlushBytes int64         `yaml:"flush_bytes"`
	Interval   time.Duration `yaml:"interval"`
}

// ConfigureLogrus 配置Logrus使用ELK Hook
func ConfigureLogrus(logger *logrus.Logger, cfg Config) error {
	if !cfg.Hook.Enabled {
		return nil
	}

	// 转换日志级别
	var levels []logrus.Level
	for _, levelStr := range cfg.Hook.Levels {
		level, err := logrus.ParseLevel(levelStr)
		if err != nil {
			return err
		}
		levels = append(levels, level)
	}

	// 创建Hook
	hook, err := NewElkHook(ElkHookOptions{
		Config: &cfg.Elasticsearch,
		Levels: levels,
		Index:  cfg.Hook.Index,
		BatchConfig: &BulkProcessorConfig{
			BatchSize:    cfg.Hook.Batch.Size,
			FlushBytes:   int64(cfg.Hook.Batch.FlushBytes),
			Interval:     cfg.Hook.Batch.Interval,
			RetryCount:   cfg.Hook.Retry.MaxRetries,
			RetryWait:    cfg.Hook.Retry.InitialWait,
			MaxWait:      cfg.Hook.Retry.MaxWait,
			CloseTimeout: 30 * time.Second,
		},
		ErrorLogger: logger,
	})
	if err != nil {
		return err
	}

	logger.AddHook(hook)
	return nil
}
