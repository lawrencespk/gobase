package errors

import (
	"fmt"
	"gobase/pkg/errors/types"
)

// Wrap 包装错误并添加上下文信息
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	// 如果已经是自定义错误，保留原始错误码
	if customErr, ok := err.(types.Error); ok {
		return NewError(customErr.Code(), fmt.Sprintf("%s: %s", message, customErr.Message()), err)
	}

	// 其他错误包装为系统错误
	return NewSystemError(fmt.Sprintf("%s: %v", message, err), err)
}

// WrapWithCode 使用指定错误码包装错误
func WrapWithCode(err error, code string, message string) error {
	if err == nil {
		return nil
	}
	return NewError(code, fmt.Sprintf("%s: %v", message, err), err)
}

// Wrapf 包装错误并使用格式化的消息
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	message := fmt.Sprintf(format, args...)

	if customErr, ok := err.(types.Error); ok {
		return NewError(customErr.Code(), fmt.Sprintf("%s: %s", message, customErr.Message()), err)
	}

	return NewSystemError(fmt.Sprintf("%s: %v", message, err), err)
}

// WrapWithCodef 使用指定错误码和格式化消息包装错误
func WrapWithCodef(err error, code string, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	message := fmt.Sprintf(format, args...)
	return NewError(code, fmt.Sprintf("%s: %v", message, err), err)
}
