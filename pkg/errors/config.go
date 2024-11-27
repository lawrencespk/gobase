package errors

import (
	"gobase/pkg/errors/codes"
)

// 配置相关错误码 (2900-2999)

// NewConfigNotFoundError 创建配置未找到错误
func NewConfigNotFoundError(message string, cause error) error {
	return NewError(codes.ConfigNotFound, message, cause)
}

// NewConfigInvalidError 创建配置无效错误
func NewConfigInvalidError(message string, cause error) error {
	return NewError(codes.ConfigInvalid, message, cause)
}

// NewConfigUpdateError 创建配置更新错误
func NewConfigUpdateError(message string, cause error) error {
	return NewError(codes.ConfigUpdateError, message, cause)
}
