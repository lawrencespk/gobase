package errors

import (
	"gobase/pkg/errors/codes"
)

// Redis相关错误 (3300-3399)

// NewRedisConnError 创建Redis连接错误
func NewRedisConnError(message string, cause error) error {
	return NewError(codes.RedisConnError, message, cause)
}

// NewRedisAuthError 创建Redis认证错误
func NewRedisAuthError(message string, cause error) error {
	return NewError(codes.RedisAuthError, message, cause)
}

// NewRedisTimeoutError 创建Redis超时错误
func NewRedisTimeoutError(message string, cause error) error {
	return NewError(codes.RedisTimeoutError, message, cause)
}

// NewRedisClusterError 创建Redis集群错误
func NewRedisClusterError(message string, cause error) error {
	return NewError(codes.RedisClusterError, message, cause)
}

// NewRedisReadOnlyError 创建Redis只读错误
func NewRedisReadOnlyError(message string, cause error) error {
	return NewError(codes.RedisReadOnlyError, message, cause)
}

// NewRedisCommandError 创建Redis命令执行错误
func NewRedisCommandError(message string, cause error) error {
	return NewError(codes.RedisCommandError, message, cause)
}

// NewRedisPipelineError 创建Redis管道操作错误
func NewRedisPipelineError(message string, cause error) error {
	return NewError(codes.RedisPipelineError, message, cause)
}

// NewRedisPoolExhaustedError 创建Redis连接池耗尽错误
func NewRedisPoolExhaustedError(message string, cause error) error {
	return NewError(codes.RedisPoolExhaustedError, message, cause)
}

// NewRedisReplicationError 创建Redis复制错误
func NewRedisReplicationError(message string, cause error) error {
	return NewError(codes.RedisReplicationError, message, cause)
}

// NewRedisScriptError 创建Redis脚本执行错误
func NewRedisScriptError(message string, cause error) error {
	return NewError(codes.RedisScriptError, message, cause)
}

// NewRedisWatchError 创建Redis监视错误
func NewRedisWatchError(message string, cause error) error {
	return NewError(codes.RedisWatchError, message, cause)
}

// NewRedisLockError 创建Redis锁操作错误
func NewRedisLockError(message string, cause error) error {
	return NewError(codes.RedisLockError, message, cause)
}

// NewRedisMaxMemoryError 创建Redis内存超限错误
func NewRedisMaxMemoryError(message string, cause error) error {
	return NewError(codes.RedisMaxMemoryError, message, cause)
}

// NewRedisLoadingError 创建Redis加载数据错误
func NewRedisLoadingError(message string, cause error) error {
	return NewError(codes.RedisLoadingError, message, cause)
}

// NewRedisInvalidConfigError 创建Redis配置无效错误
func NewRedisInvalidConfigError(message string, cause error) error {
	return NewError(codes.RedisInvalidConfigError, message, cause)
}

// NewRedisKeyNotFoundError 创建Redis键不存在错误
func NewRedisKeyNotFoundError(message string, cause error) error {
	return NewError(codes.RedisKeyNotFoundError, message, cause)
}

// NewRedisKeyExpiredError 创建Redis键已过期错误
func NewRedisKeyExpiredError(message string, cause error) error {
	return NewError(codes.RedisKeyExpiredError, message, cause)
}
