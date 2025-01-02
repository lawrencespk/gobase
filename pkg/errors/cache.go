package errors

import (
	"gobase/pkg/errors/codes"
)

// 缓存相关错误码 (2700-2799)

// NewCacheMissError 创建缓存未命中错误
func NewCacheMissError(message string, cause error) error {
	return NewError(codes.CacheMissError, message, cause)
}

// NewCacheExpiredError 创建缓存已过期错误
func NewCacheExpiredError(message string, cause error) error {
	return NewError(codes.CacheExpiredError, message, cause)
}

// NewCacheFullError 创建缓存已满错误
func NewCacheFullError(message string, cause error) error {
	return NewError(codes.CacheFullError, message, cause)
}

// NewCacheNotFoundError 创建缓存层级不存在错误
func NewCacheNotFoundError(message string, cause error) error {
	return NewError(codes.CacheNotFoundError, message, cause)
}
