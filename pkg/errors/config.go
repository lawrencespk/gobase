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

// NewConfigParseError 创建配置解析错误
func NewConfigParseError(message string, cause error) error {
	return NewError(codes.ConfigParseError, message, cause)
}

// NewConfigWatchError 创建配置监听错误
func NewConfigWatchError(message string, cause error) error {
	return NewError(codes.ConfigWatchError, message, cause)
}

// NewConfigValidateError 创建配置验证错误
func NewConfigValidateError(message string, cause error) error {
	return NewError(codes.ConfigValidateError, message, cause)
}

// NewConfigTypeError 创建配置类型错误
func NewConfigTypeError(message string, cause error) error {
	return NewError(codes.ConfigTypeError, message, cause)
}

// NewConfigProviderError 创建配置提供者错误
func NewConfigProviderError(message string, cause error) error {
	return NewError(codes.ConfigProviderError, message, cause)
}

// NewConfigCloseError 创建配置关闭错误
func NewConfigCloseError(message string, cause error) error {
	return NewError(codes.ConfigCloseError, message, cause)
}

// NewConfigKeyNotFoundError 创建配置键不存在错误
func NewConfigKeyNotFoundError(message string, cause error) error {
	return NewError(codes.ConfigKeyNotFoundError, message, cause)
}

// NewConfigNotLoadedError 创建配置未加载错误
func NewConfigNotLoadedError(message string, cause error) error {
	return NewError(codes.ConfigNotLoadedError, message, cause)
}
