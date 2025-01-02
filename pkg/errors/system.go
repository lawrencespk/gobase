package errors

import (
	"gobase/pkg/errors/codes"
)

// NewSystemError 创建系统错误
func NewSystemError(message string, cause error) error {
	return NewError(codes.SystemError, message, cause)
}

// NewConfigError 创建配置错误
func NewConfigError(message string, cause error) error {
	return NewError(codes.ConfigError, message, cause)
}

// NewNetworkError 创建网络错误
func NewNetworkError(message string, cause error) error {
	return NewError(codes.NetworkError, message, cause)
}

// NewDatabaseError 创建数据库错误
func NewDatabaseError(message string, cause error) error {
	return NewError(codes.DatabaseError, message, cause)
}

// NewCacheError 创建缓存错误
func NewCacheError(message string, cause error) error {
	return NewError(codes.CacheError, message, cause)
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(message string, cause error) error {
	return NewError(codes.TimeoutError, message, cause)
}

// NewValidationError 创建验证错误
func NewValidationError(message string, cause error) error {
	return NewError(codes.ValidationError, message, cause)
}

// NewSerializationError 创建序列化错误
func NewSerializationError(message string, cause error) error {
	return NewError(codes.SerializationError, message, cause)
}

// NewThirdPartyError 创建第三方服务错误
func NewThirdPartyError(message string, cause error) error {
	return NewError(codes.ThirdPartyError, message, cause)
}

// NewInitializeError 创建初始化错误
func NewInitializeError(message string, cause error) error {
	return NewError(codes.InitializeError, message, cause)
}

// NewShutdownError 创建关闭错误
func NewShutdownError(message string, cause error) error {
	return NewError(codes.ShutdownError, message, cause)
}

// NewMemoryError 创建内存错误
func NewMemoryError(message string, cause error) error {
	return NewError(codes.MemoryError, message, cause)
}

// NewDiskError 创建磁盘错误
func NewDiskError(message string, cause error) error {
	return NewError(codes.DiskError, message, cause)
}

// NewResourceExhaustedError 创建资源耗尽错误
func NewResourceExhaustedError(message string, cause error) error {
	return NewError(codes.ResourceExhausted, message, cause)
}

// NewInitializationError 创建初始化错误
func NewInitializationError(message string, cause error) error {
	return NewError(codes.InitializeError, message, cause)
}

// NewCacheNotFoundError 创建缓存未找到错误
func NewCacheNotFoundError(message string, cause error) error {
	return NewError(codes.CacheMissError, message, cause)
}
