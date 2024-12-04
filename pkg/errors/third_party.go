package errors

import (
	"gobase/pkg/errors/codes"
)

// 第三方服务错误码 (2500-2599)

// NewAPIError 创建API错误
func NewAPIError(message string, cause error) error {
	return NewError(codes.APIError, message, cause)
}

// NewServiceError 创建服务错误
func NewServiceError(message string, cause error) error {
	return NewError(codes.ServiceError, message, cause)
}

// NewIntegrationError 创建集成错误
func NewIntegrationError(message string, cause error) error {
	return NewError(codes.IntegrationError, message, cause)
}

// NewDependencyError 创建依赖错误
func NewDependencyError(message string, cause error) error {
	return NewError(codes.DependencyError, message, cause)
}

// NewELKConnectionError 创建 ELK 连接错误
func NewELKConnectionError(message string, cause error) error {
	return NewError(codes.ELKConnectionError, message, cause)
}

// NewELKIndexError 创建 ELK 索引错误
func NewELKIndexError(message string, cause error) error {
	return NewError(codes.ELKIndexError, message, cause)
}

// NewELKQueryError 创建 ELK 查询错误
func NewELKQueryError(message string, cause error) error {
	return NewError(codes.ELKQueryError, message, cause)
}

// NewELKBulkError 创建 ELK 批量操作错误
func NewELKBulkError(message string, cause error) error {
	return NewError(codes.ELKBulkError, message, cause)
}

// NewELKConfigError 创建 ELK 配置错误
func NewELKConfigError(message string, cause error) error {
	return NewError(codes.ELKConfigError, message, cause)
}

// NewELKTimeoutError 创建 ELK 超时错误
func NewELKTimeoutError(message string, cause error) error {
	return NewError(codes.ELKTimeoutError, message, cause)
}
