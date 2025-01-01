package redis

import (
	"context"
	"gobase/pkg/errors"
	"io"
	"strings"

	"github.com/go-redis/redis/v8"
)

// 预定义错误变量
var (
	// ErrInvalidConfig 无效的配置错误
	ErrInvalidConfig = errors.NewRedisInvalidConfigError("invalid redis config", nil)

	// ErrConnectionFailed 连接失败错误
	ErrConnectionFailed = errors.NewRedisConnError("failed to connect to redis", nil)

	// ErrOperationFailed 操作失败错误
	ErrOperationFailed = errors.NewRedisCommandError("redis operation failed", nil)

	// ErrKeyNotFound 键不存在错误
	ErrKeyNotFound = errors.NewRedisKeyNotFoundError("redis key not found", nil)

	// ErrPoolTimeout 连接池超时错误
	ErrPoolTimeout = errors.NewRedisPoolExhaustedError("redis pool timeout", nil)

	// ErrConnPool 连接池错误
	ErrConnPool = errors.NewRedisPoolExhaustedError("redis connection pool error", nil)

	// ErrNil 表示键不存在
	ErrNil = errors.NewRedisKeyNotFoundError("redis: nil", nil)
)

// 错误消息常量
const (
	errKeyNotFound    = "key not found"
	errFieldNotFound  = "field not found"
	errTimeout        = "operation timed out"
	errConnFailed     = "connection failed"
	errPipelineFailed = "pipeline operation failed"
)

// handleRedisError 处理Redis错误
func handleRedisError(err error, msg string) error {
	if err == nil {
		return nil
	}

	// 1. 检查是否已经是包装过的错误
	if errors.HasErrorCode(err, "") {
		return err
	}

	// 2. 处理上下文超时错误 (提升优先级，并使用通用的 TimeoutError)
	if err == context.DeadlineExceeded {
		return errors.NewTimeoutError(errTimeout, err)
	}
	if strings.Contains(err.Error(), "context deadline exceeded") {
		return errors.NewTimeoutError(errTimeout, err)
	}

	// 3. 处理 redis.Nil 错误 (键不存在)
	if err == redis.Nil {
		return errors.NewRedisKeyNotFoundError(msg, err)
	}

	// 4. 处理连接错误
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		return errors.NewRedisConnError(errConnFailed, err)
	}

	// 5. 处理超时相关错误 (Redis 特定的超时)
	if strings.Contains(strings.ToLower(err.Error()), "timeout") {
		return errors.NewRedisTimeoutError(errTimeout, err)
	}

	// 6. 处理只读错误
	if strings.Contains(strings.ToLower(err.Error()), "readonly") {
		return errors.NewRedisReadOnlyError(msg, err)
	}

	// 7. 处理认证错误
	if strings.Contains(strings.ToLower(err.Error()), "auth") {
		return errors.NewRedisAuthError(msg, err)
	}

	// 8. 处理内存相关错误
	if strings.Contains(strings.ToLower(err.Error()), "oom") ||
		strings.Contains(strings.ToLower(err.Error()), "out of memory") {
		return errors.NewRedisMaxMemoryError(msg, err)
	}

	// 9. 处理加载错误
	if strings.Contains(strings.ToLower(err.Error()), "loading") {
		return errors.NewRedisLoadingError(msg, err)
	}

	// 10. 处理集群错误
	if strings.Contains(strings.ToLower(err.Error()), "cluster") {
		return errors.NewRedisClusterError(msg, err)
	}

	// 11. 处理复制错误
	if strings.Contains(strings.ToLower(err.Error()), "replication") {
		return errors.NewRedisReplicationError(msg, err)
	}

	// 12. 处理脚本错误
	if strings.Contains(strings.ToLower(err.Error()), "script") {
		return errors.NewRedisScriptError(msg, err)
	}

	// 13. 处理管道错误
	if strings.Contains(strings.ToLower(err.Error()), "pipeline") {
		return errors.NewRedisPipelineError(msg, err)
	}

	// 14. 处理监视错误
	if strings.Contains(strings.ToLower(err.Error()), "watch") {
		return errors.NewRedisWatchError(msg, err)
	}
	// 15. 处理锁错误
	if strings.Contains(strings.ToLower(err.Error()), "lock") {
		return errors.NewRedisLockError(msg, err)
	}

	// 默认返回命令执行错误
	return errors.NewRedisCommandError(msg, err)
}
