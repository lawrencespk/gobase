package errors

import (
	"fmt"
	"runtime"
	"strings"

	"gobase/pkg/errors/codes"
	"gobase/pkg/errors/types"
)

// 错误码映射关系
var errorCodeMappings = map[string][]string{
	codes.CacheError: {
		codes.RedisConnError,
		codes.RedisAuthError,
		codes.RedisTimeoutError,
		codes.RedisClusterError,
		codes.RedisReadOnlyError,
		codes.RedisCommandError,
		codes.RedisPipelineError,
		codes.RedisPoolExhaustedError,
		codes.RedisReplicationError,
		codes.RedisScriptError,
		codes.RedisWatchError,
		codes.RedisLockError,
		codes.RedisMaxMemoryError,
		codes.RedisLoadingError,
		codes.RedisInvalidConfigError,
	},
	codes.NotFound: {
		codes.RedisKeyNotFoundError,
		codes.RedisKeyExpiredError,
	},
}

// checkErrorCodeMapping 检查错误码是否匹配或在映射关系中
func checkErrorCodeMapping(errorCode, targetCode string) bool {
	// 直接匹配
	if errorCode == targetCode {
		return true
	}

	// 检查映射关系
	if mappedCodes, exists := errorCodeMappings[targetCode]; exists {
		for _, mappedCode := range mappedCodes {
			if errorCode == mappedCode {
				return true
			}
		}
	}

	return false
}

// baseError 基础错误类型
type baseError struct {
	code    string
	message string
	details []interface{}
	cause   error
	stack   []string
}

// Is 实现错误比较
func (e *baseError) Is(target error) bool {
	// 尝试将目标错误转换为我们的错误类型
	if t, ok := target.(*baseError); ok {
		// 比较错误码
		return checkErrorCodeMapping(e.code, t.code)
	}
	return false
}

// NewError 创建新的错误
func NewError(code string, message string, cause error) types.Error {
	e := &baseError{
		code:    code,
		message: message,
		cause:   cause,
		stack:   make([]string, 0),
	}
	e.captureStack()
	return e
}

func (e *baseError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s (code=%s): %v",
			e.message, e.code, e.cause)
	}
	return fmt.Sprintf("%s (code=%s)", e.message, e.code)
}

func (e *baseError) Code() string {
	return e.code
}

func (e *baseError) Message() string {
	return e.message
}

func (e *baseError) Details() []interface{} {
	return e.details
}

func (e *baseError) Unwrap() error {
	return e.cause
}

// WithDetails 添加错误详情
func (e *baseError) WithDetails(details ...interface{}) types.Error {
	e.details = append(e.details, details...)
	return e
}

// Stack 获取堆栈信息
func (e *baseError) Stack() []string {
	return e.stack
}

// captureStack 捕获堆栈信息
func (e *baseError) captureStack() {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.File, "runtime/") {
			e.stack = append(e.stack, fmt.Sprintf("%s:%d %s",
				frame.File, frame.Line, frame.Function))
		}
		if !more {
			break
		}
	}
}
