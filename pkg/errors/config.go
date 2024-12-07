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

// NewConfigLoadError 创建配置加载错误
func NewConfigLoadError(message string, cause error) error {
	return NewError(codes.ConfigLoadError, message, cause)
}

// NewNacosConnectError 创建Nacos连接错误
func NewNacosConnectError(message string, cause error) error {
	return NewError(codes.NacosConnectError, message, cause)
}

// NewNacosAuthError 创建Nacos认证错误
func NewNacosAuthError(message string, cause error) error {
	return NewError(codes.NacosAuthError, message, cause)
}

// NewNacosConfigError 创建Nacos配置错误
func NewNacosConfigError(message string, cause error) error {
	return NewError(codes.NacosConfigError, message, cause)
}

// NewNacosWatchError 创建Nacos监听错误
func NewNacosWatchError(message string, cause error) error {
	return NewError(codes.NacosWatchError, message, cause)
}

// NewNacosPublishError 创建Nacos发布错误
func NewNacosPublishError(message string, cause error) error {
	return NewError(codes.NacosPublishError, message, cause)
}

// NewNacosNamespaceError 创建Nacos命名空间错误
func NewNacosNamespaceError(message string, cause error) error {
	return NewError(codes.NacosNamespaceError, message, cause)
}

// NewNacosGroupError 创建Nacos分组错误
func NewNacosGroupError(message string, cause error) error {
	return NewError(codes.NacosGroupError, message, cause)
}

// NewNacosDataIDError 创建Nacos数据ID错误
func NewNacosDataIDError(message string, cause error) error {
	return NewError(codes.NacosDataIDError, message, cause)
}

// NewNacosTimeoutError 创建Nacos超时错误
func NewNacosTimeoutError(message string, cause error) error {
	return NewError(codes.NacosTimeoutError, message, cause)
}

// NewNacosOperationError 创建Nacos操作错误
func NewNacosOperationError(message string, cause error) error {
	return NewError(codes.NacosOperationError, message, cause)
}
