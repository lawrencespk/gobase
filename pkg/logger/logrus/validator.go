package logrus

import (
	"compress/gzip"
	"fmt"
	"gobase/pkg/logger/types"
	"time"
)

// ValidateOptions 验证所有配置选项
func ValidateOptions(opts *Options) error {
	if opts == nil {
		return fmt.Errorf("options cannot be nil")
	}

	// 验证基础配置
	if err := validateBasicOptions(opts); err != nil {
		return fmt.Errorf("basic options validation failed: %v", err)
	}

	// 验证压缩配置
	if err := validateCompressConfig(&opts.CompressConfig); err != nil {
		return fmt.Errorf("compress config validation failed: %v", err)
	}

	// 验证清理配置
	if err := validateCleanupConfig(&opts.CleanupConfig); err != nil {
		return fmt.Errorf("cleanup config validation failed: %v", err)
	}

	// 验证异步配置
	if err := validateAsyncConfig(&opts.AsyncConfig); err != nil {
		return fmt.Errorf("async config validation failed: %v", err)
	}

	// 验证恢复配置
	if err := validateRecoveryConfig(&opts.RecoveryConfig); err != nil {
		return fmt.Errorf("recovery config validation failed: %v", err)
	}

	// 验证配置冲突
	if err := validateConfigConflicts(opts); err != nil {
		return fmt.Errorf("config conflicts found: %v", err)
	}

	return nil
}

// validateBasicOptions 验证基础配置
func validateBasicOptions(opts *Options) error {
	// 验证日志级别
	if !isValidLogLevel(opts.Level) {
		return fmt.Errorf("invalid log level: %v", opts.Level)
	}

	// 验证输出路径
	if len(opts.OutputPaths) == 0 {
		return fmt.Errorf("output paths cannot be empty")
	}

	// 验证文件大小限制
	if opts.MaxSize <= 0 {
		return fmt.Errorf("max size must be positive")
	}

	return nil
}

// validateCompressConfig 验证压缩配置
func validateCompressConfig(config *CompressConfig) error {
	if !config.Enable {
		return nil
	}

	// 验证压缩算法
	if config.Algorithm != "" && config.Algorithm != "gzip" {
		return fmt.Errorf("unsupported compression algorithm: %s", config.Algorithm)
	}

	// 验证压缩级别
	if config.Level < gzip.NoCompression || config.Level > gzip.BestCompression {
		return fmt.Errorf("invalid compression level: %d", config.Level)
	}

	// 验证压缩间隔
	if config.Interval < time.Second {
		return fmt.Errorf("compress interval must be at least 1 second")
	}

	return nil
}

// validateCleanupConfig 验证清理配置
func validateCleanupConfig(config *CleanupConfig) error {
	if !config.Enable {
		return nil
	}

	// 验证最大备份数
	if config.MaxBackups <= 0 {
		return fmt.Errorf("max backups must be positive")
	}

	// 验证最大保留天数
	if config.MaxAge <= 0 {
		return fmt.Errorf("max age must be positive")
	}

	// 验证清理间隔
	if config.Interval < time.Minute {
		return fmt.Errorf("cleanup interval must be at least 1 minute")
	}

	return nil
}

// validateAsyncConfig 验证异步配置
func validateAsyncConfig(config *AsyncConfig) error {
	if !config.Enable {
		return nil
	}

	// 验证缓冲区大小
	if config.BufferSize <= 0 {
		return fmt.Errorf("buffer size must be positive")
	}

	// 验证刷新间隔
	if config.FlushInterval < time.Millisecond {
		return fmt.Errorf("flush interval must be at least 1 millisecond")
	}

	return nil
}

// validateRecoveryConfig 验证恢复配置
func validateRecoveryConfig(config *RecoveryConfig) error {
	if !config.Enable {
		return nil
	}

	// 验证最大重试次数
	if config.MaxRetries <= 0 {
		return fmt.Errorf("max retries must be positive")
	}

	// 验证重试间隔
	if config.RetryInterval < time.Millisecond {
		return fmt.Errorf("retry interval must be at least 1 millisecond")
	}

	// 验证堆栈大小
	if config.EnableStackTrace && config.MaxStackSize <= 0 {
		return fmt.Errorf("max stack size must be positive when stack trace is enabled")
	}

	return nil
}

// validateConfigConflicts 验证配置冲突
func validateConfigConflicts(opts *Options) error {
	// 检查异步写入和错误恢复的冲突
	if opts.AsyncConfig.Enable && opts.RecoveryConfig.Enable {
		if opts.AsyncConfig.BlockOnFull && opts.RecoveryConfig.MaxRetries > 0 {
			return fmt.Errorf("blocking async write conflicts with retry mechanism")
		}
	}

	// 检查压缩和清理的冲突
	if opts.CompressConfig.Enable && opts.CleanupConfig.Enable {
		if opts.CompressConfig.DeleteSource && opts.CleanupConfig.MaxBackups > 0 {
			return fmt.Errorf("compress delete source conflicts with cleanup max backups")
		}
	}

	return nil
}

// isValidLogLevel 验证日志级别是否有效
func isValidLogLevel(level types.Level) bool {
	switch level {
	case types.DebugLevel, types.InfoLevel, types.WarnLevel, types.ErrorLevel, types.FatalLevel:
		return true
	default:
		return false
	}
}
