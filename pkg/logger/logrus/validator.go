package logrus

import (
	"fmt"
	"strings"

	"gobase/pkg/errors"
)

// ValidateOptions 验证日志选项
func ValidateOptions(opts *Options) error {
	// 基本选项验证
	if err := validateBasicOptions(opts); err != nil {
		return errors.NewValidationError("basic options validation failed", err) // 基本选项验证失败
	}

	// 压缩配置验证
	if opts.CompressConfig.Enable {
		if err := validateCompressConfig(&opts.CompressConfig); err != nil {
			return errors.NewValidationError("compress config validation failed", err) // 压缩配置验证失败
		}
	}

	// 清理配置验证
	if opts.CleanupConfig.Enable {
		if err := validateCleanupConfig(&opts.CleanupConfig); err != nil {
			return errors.NewValidationError("cleanup config validation failed", err) // 清理配置验证失败
		}
	}

	// 异步配置验证
	if opts.AsyncConfig.Enable {
		if err := validateAsyncConfig(&opts.AsyncConfig); err != nil {
			return errors.NewValidationError("async config validation failed", err) // 异步配置验证失败
		}
	}

	// 恢复配置验证
	if opts.RecoveryConfig.Enable {
		if err := validateRecoveryConfig(&opts.RecoveryConfig); err != nil {
			return errors.NewValidationError("recovery config validation failed", err) // 恢复配置验证失败
		}
	}

	// 配置冲突检查
	if err := validateConfigConflicts(opts); err != nil {
		return errors.NewValidationError("config conflicts validation failed", err) // 配置冲突检查失败
	}

	return nil
}

// validateBasicOptions 验证基本选项
func validateBasicOptions(opts *Options) error {
	// 验证日志级别
	if !ValidateLevel(opts.Level.String()) {
		return errors.NewValidationError(fmt.Sprintf("invalid log level: %s", opts.Level), nil) // 无效的日志级别
	}

	// 验证输出路径
	if len(opts.OutputPaths) == 0 {
		return errors.NewValidationError("output paths cannot be empty", nil) // 输出路径不能为空
	}

	// 验证队列配置 - 只在必要时验证
	if opts.QueueConfig != (QueueConfig{}) { // 只有当 QueueConfig 被设置时才验证
		if opts.QueueConfig.MaxSize <= 0 {
			return errors.NewValidationError("max size must be positive", nil) // 最大大小必须为正数
		}
		if opts.QueueConfig.BatchSize <= 0 {
			return errors.NewValidationError("batch size must be positive", nil) // 批处理大小必须为正数
		}
		if opts.QueueConfig.Workers <= 0 {
			return errors.NewValidationError("workers count must be positive", nil) // 工作协程数量必须为正数
		}
		if opts.QueueConfig.FlushInterval <= 0 {
			return errors.NewValidationError("flush interval must be positive", nil) // 刷新间隔必须为正数
		}
	}

	return nil
}

// validateCompressConfig 验证压缩配置
func validateCompressConfig(config *CompressConfig) error {
	if !isValidCompressAlgorithm(config.Algorithm) {
		return errors.NewValidationError(fmt.Sprintf("invalid compress algorithm: %s", config.Algorithm), nil) // 无效的压缩算法
	}
	if config.Level < 0 || config.Level > 9 {
		return errors.NewValidationError(fmt.Sprintf("invalid compress level: %d", config.Level), nil) // 无效的压缩级别
	}
	return nil
}

// validateCleanupConfig 验证清理配置
func validateCleanupConfig(config *CleanupConfig) error {
	if config.MaxAge < 0 {
		return errors.NewValidationError("max age must be non-negative", nil) // 最大年龄必须为非负数
	}
	if config.MaxBackups < 0 {
		return errors.NewValidationError("max backups must be non-negative", nil) // 最大备份数必须为非负数
	}
	if config.Interval <= 0 {
		return errors.NewValidationError("cleanup interval must be positive", nil) // 清理间隔必须为正数
	}
	return nil
}

// validateAsyncConfig 验证异步配置
func validateAsyncConfig(config *AsyncConfig) error {
	if config.BufferSize <= 0 {
		return errors.NewValidationError("buffer size must be positive", nil) // 缓冲区大小必须为正数
	}
	if config.FlushInterval <= 0 {
		return errors.NewValidationError("flush interval must be positive", nil) // 刷新间隔必须为正数
	}
	return nil
}

// validateRecoveryConfig 验证恢复配置
func validateRecoveryConfig(config *RecoveryConfig) error {
	if config.MaxRetries < 0 {
		return errors.NewValidationError("max retries must be non-negative", nil) // 最大重试次数必须为非负数
	}
	if config.RetryInterval <= 0 {
		return errors.NewValidationError("retry interval must be positive", nil) // 重试间隔必须为正数
	}
	if config.MaxStackSize < 0 {
		return errors.NewValidationError("max stack size must be non-negative", nil) // 最大堆栈大小必须为非负数
	}
	return nil
}

// validateConfigConflicts 验证配置冲突
func validateConfigConflicts(opts *Options) error {
	// 异步和恢复配置不能同时启用
	if opts.AsyncConfig.Enable && opts.RecoveryConfig.Enable {
		return errors.NewValidationError("async and recovery configs cannot be enabled simultaneously", nil) // 异步和恢复配置不能同时启用
	}
	return nil
}

// ValidateLevel 验证日志级别
func ValidateLevel(level string) bool {
	// 有效的日志级别
	validLevels := map[string]bool{
		"debug": true, // 调试级别
		"info":  true, // 信息级别
		"warn":  true, // 警告级别
		"error": true, // 错误级别
		"fatal": true, // 严重级别
	}
	return validLevels[strings.ToLower(level)]
}

// isValidCompressAlgorithm 验证压缩算法
func isValidCompressAlgorithm(algorithm string) bool {
	// 有效的压缩算法
	validAlgorithms := map[string]bool{
		"gzip": true, // gzip 压缩算法
		"zlib": true, // zlib 压缩算法
	}
	return validAlgorithms[strings.ToLower(algorithm)]
}
