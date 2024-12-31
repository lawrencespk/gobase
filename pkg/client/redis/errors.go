package redis

import (
	"context"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"io"
	"strings"

	"github.com/go-redis/redis/v8"
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

const (
	errKeyNotFound    = "key not found"
	errFieldNotFound  = "field not found"
	errTimeout        = "operation timed out"
	errConnFailed     = "connection failed"
	errPipelineFailed = "pipeline operation failed"
)

func handleRedisError(err error, msg string) error {
	if err == nil {
		return nil
	}

	// 1. 检查是否已经是包装过的错误
	if errors.HasErrorCode(err, "") {
		return err
	}

	// 2. 处理 redis.Nil 错误 (键不存在)
	if err == redis.Nil {
		return errors.NewError(codes.CacheError, msg, err) // 改为返回 CacheError
	}

	// 3. 处理上下文超时错误
	if err == context.DeadlineExceeded {
		return errors.NewError(codes.TimeoutError, errTimeout, err)
	}
	if strings.Contains(err.Error(), "context deadline exceeded") {
		return errors.NewError(codes.TimeoutError, errTimeout, err)
	}

	// 4. 处理连接错误
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return errors.NewError(codes.CacheError, errConnFailed, err)
	}

	// 5. 处理超时相关错误
	if strings.Contains(strings.ToLower(err.Error()), "timeout") {
		return errors.NewError(codes.TimeoutError, errTimeout, err)
	}

	// 6. 其他错误
	return errors.NewError(codes.CacheError, msg, err)
}
