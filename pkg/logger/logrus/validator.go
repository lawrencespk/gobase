package logrus

import (
	"errors"
	"fmt"
	"strings"
)

// ValidateOptions 验证日志选项
func ValidateOptions(opts *Options) error {
	// 基本选项验证
	if err := validateBasicOptions(opts); err != nil {
		return fmt.Errorf("basic options validation failed: %v", err)
	}

	// 压缩配置验证
	if opts.CompressConfig.Enable {
		if err := validateCompressConfig(&opts.CompressConfig); err != nil {
			return fmt.Errorf("compress config validation failed: %v", err)
		}
	}

	// 清理配置验证
	if opts.CleanupConfig.Enable {
		if err := validateCleanupConfig(&opts.CleanupConfig); err != nil {
			return fmt.Errorf("cleanup config validation failed: %v", err)
		}
	}

	// 异步配置验证
	if opts.AsyncConfig.Enable {
		if err := validateAsyncConfig(&opts.AsyncConfig); err != nil {
			return fmt.Errorf("async config validation failed: %v", err)
		}
	}

	// 恢复配置验证
	if opts.RecoveryConfig.Enable {
		if err := validateRecoveryConfig(&opts.RecoveryConfig); err != nil {
			return fmt.Errorf("recovery config validation failed: %v", err)
		}
	}

	// 配置冲突检查
	if err := validateConfigConflicts(opts); err != nil {
		return fmt.Errorf("config conflicts validation failed: %v", err)
	}

	return nil
}

// validateBasicOptions 验证基本选项
func validateBasicOptions(opts *Options) error {
	// 验证日志级别
	if !ValidateLevel(opts.Level.String()) {
		return fmt.Errorf("invalid log level: %s", opts.Level)
	}

	// 验证输出路径
	if len(opts.OutputPaths) == 0 {
		return errors.New("output paths cannot be empty")
	}

	// 验证队列配置 - 只在必要时验证
	if opts.QueueConfig != (QueueConfig{}) { // 只有当 QueueConfig 被设置时才验证
		if opts.QueueConfig.MaxSize <= 0 {
			return errors.New("max size must be positive")
		}
		if opts.QueueConfig.BatchSize <= 0 {
			return errors.New("batch size must be positive")
		}
		if opts.QueueConfig.Workers <= 0 {
			return errors.New("workers count must be positive")
		}
		if opts.QueueConfig.FlushInterval <= 0 {
			return errors.New("flush interval must be positive")
		}
	}

	return nil
}

// validateCompressConfig 验证压缩配置
func validateCompressConfig(config *CompressConfig) error {
	if !isValidCompressAlgorithm(config.Algorithm) {
		return fmt.Errorf("invalid compress algorithm: %s", config.Algorithm)
	}
	if config.Level < 0 || config.Level > 9 {
		return fmt.Errorf("invalid compress level: %d", config.Level)
	}
	return nil
}

// validateCleanupConfig 验证清理配置
func validateCleanupConfig(config *CleanupConfig) error {
	if config.MaxAge < 0 {
		return fmt.Errorf("max age must be non-negative")
	}
	if config.MaxBackups < 0 {
		return fmt.Errorf("max backups must be non-negative")
	}
	if config.Interval <= 0 {
		return fmt.Errorf("cleanup interval must be positive")
	}
	return nil
}

// validateAsyncConfig 验证异步配置
func validateAsyncConfig(config *AsyncConfig) error {
	if config.BufferSize <= 0 {
		return fmt.Errorf("buffer size must be positive")
	}
	if config.FlushInterval <= 0 {
		return fmt.Errorf("flush interval must be positive")
	}
	return nil
}

// validateRecoveryConfig 验证恢复配置
func validateRecoveryConfig(config *RecoveryConfig) error {
	if config.MaxRetries < 0 {
		return fmt.Errorf("max retries must be non-negative")
	}
	if config.RetryInterval <= 0 {
		return fmt.Errorf("retry interval must be positive")
	}
	if config.MaxStackSize < 0 {
		return fmt.Errorf("max stack size must be non-negative")
	}
	return nil
}

// validateConfigConflicts 验证配置冲突
func validateConfigConflicts(opts *Options) error {
	if opts.AsyncConfig.Enable && opts.RecoveryConfig.Enable {
		return fmt.Errorf("async and recovery configs cannot be enabled simultaneously")
	}
	return nil
}

// ValidateLevel 验证日志级别
func ValidateLevel(level string) bool {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	return validLevels[strings.ToLower(level)]
}

// isValidCompressAlgorithm 验证压缩算法
func isValidCompressAlgorithm(algorithm string) bool {
	validAlgorithms := map[string]bool{
		"gzip": true,
		"zlib": true,
	}
	return validAlgorithms[strings.ToLower(algorithm)]
}
