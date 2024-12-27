package redis

import (
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

// ErrInvalidConfig 无效的配置错误
var ErrInvalidConfig = errors.NewError(codes.CacheError, "invalid redis config", nil)

// ErrConnectionFailed 连接失败错误
var ErrConnectionFailed = errors.NewError(codes.CacheError, "failed to connect to redis", nil)

// ErrOperationFailed 操作失败错误
var ErrOperationFailed = errors.NewError(codes.CacheError, "redis operation failed", nil)

// ErrKeyNotFound 键不存在错误
var ErrKeyNotFound = errors.NewNotFoundError("redis key not found", nil)

// ErrPoolTimeout 连接池超时错误
var ErrPoolTimeout = errors.NewError(codes.CacheError, "redis pool timeout", nil)

// ErrConnPool 连接池错误
var ErrConnPool = errors.NewError(codes.CacheError, "redis connection pool error", nil)
