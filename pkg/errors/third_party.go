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
