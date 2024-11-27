package errors

import (
	"fmt"
	"runtime"
	"strings"

	"gobase/pkg/errors/types"
)

// baseError 基础错误类型
type baseError struct {
	code    string
	message string
	details []interface{}
	cause   error
	stack   []string
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
