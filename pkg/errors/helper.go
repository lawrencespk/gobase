package errors

import (
	"errors"
	"fmt"
	"gobase/pkg/errors/codes"
	"gobase/pkg/errors/types"
)

// Is 判断错误链中是否包含指定错误
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As 将错误转换为指定类型
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// IsBusinessError 判断是否为业务错误
func IsBusinessError(err error) bool {
	var e types.Error
	if errors.As(err, &e) {
		code := e.Code()
		return code >= "2000" && code < "3000"
	}
	return false
}

// IsSystemError 判断是否为系统错误
func IsSystemError(err error) bool {
	var e types.Error
	if errors.As(err, &e) {
		code := e.Code()
		return code >= "1000" && code < "2000"
	}
	return false
}

// GetErrorCode 获取错误码
func GetErrorCode(err error) string {
	var e types.Error
	if errors.As(err, &e) {
		return e.Code()
	}
	return codes.SystemError
}

// GetErrorMessage 获取错误信息
func GetErrorMessage(err error) string {
	var e types.Error
	if errors.As(err, &e) {
		return e.Message()
	}
	return err.Error()
}

// GetErrorDetails 获取错误详情
func GetErrorDetails(err error) []interface{} {
	var e types.Error
	if errors.As(err, &e) {
		return e.Details()
	}
	return nil
}

// GetErrorStack 获取错误堆栈
func GetErrorStack(err error) []string {
	var e types.Error
	if errors.As(err, &e) {
		return e.Stack()
	}
	return nil
}

// HasErrorCode 检查错误是否包含指定的错误码
func HasErrorCode(err error, code string) bool {
	chain := GetErrorChain(err)
	for _, e := range chain {
		if e.Code() == code {
			return true
		}
	}
	return false
}

// IsErrorType 检查错误是否属于指定的错误类型范围
func IsErrorType(err error, startCode, endCode string) bool {
	code := GetErrorCode(err)
	return code >= startCode && code < endCode
}

// FormatErrorChain 格式化错误链信息
func FormatErrorChain(err error) string {
	if err == nil {
		return ""
	}

	chain := GetErrorChain(err)
	if len(chain) == 0 {
		return err.Error()
	}

	var result string
	for i, e := range chain {
		if i > 0 {
			result += " -> "
		}
		result += fmt.Sprintf("[%s] %s", e.Code(), e.Message())
	}
	return result
}

// AsError 将标准 error 转换为自定义 Error 类型
func AsError(err error) types.Error {
	if customErr, ok := err.(types.Error); ok {
		return customErr
	}
	return NewSystemError(err.Error(), err).(types.Error)
}
