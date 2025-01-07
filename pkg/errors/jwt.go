package errors

import (
	"gobase/pkg/errors/codes"
	"gobase/pkg/errors/types"
)

// NewTokenInvalidError 创建Token无效错误
func NewTokenInvalidError(message string, cause error) types.Error {
	return NewError(codes.TokenInvalid, message, cause)
}

// NewTokenExpiredError 创建Token过期错误
func NewTokenExpiredError(message string, cause error) types.Error {
	return NewError(codes.TokenExpired, message, cause)
}

// NewTokenRevokedError 创建Token已吊销错误
func NewTokenRevokedError(message string, cause error) types.Error {
	return NewError(codes.TokenRevoked, message, cause)
}

// NewTokenNotFoundError 创建Token不存在错误
func NewTokenNotFoundError(message string, cause error) types.Error {
	return NewError(codes.TokenNotFound, message, cause)
}

// NewTokenTypeMismatchError 创建Token类型不匹配错误
func NewTokenTypeMismatchError(message string, cause error) types.Error {
	return NewError(codes.TokenTypeMismatch, message, cause)
}

// NewClaimsMissingError 创建Claims缺失错误
func NewClaimsMissingError(message string, cause error) types.Error {
	return NewError(codes.ClaimsMissing, message, cause)
}

// NewClaimsInvalidError 创建Claims无效错误
func NewClaimsInvalidError(message string, cause error) types.Error {
	return NewError(codes.ClaimsInvalid, message, cause)
}

// NewClaimsExpiredError 创建Claims过期错误
func NewClaimsExpiredError(message string, cause error) types.Error {
	return NewError(codes.ClaimsExpired, message, cause)
}

// NewSignatureInvalidError 创建签名无效错误
func NewSignatureInvalidError(message string, cause error) types.Error {
	return NewError(codes.SignatureInvalid, message, cause)
}

// NewKeyInvalidError 创建密钥无效错误
func NewKeyInvalidError(message string, cause error) types.Error {
	return NewError(codes.KeyInvalid, message, cause)
}

// NewAlgorithmMismatchError 创建算法不匹配错误
func NewAlgorithmMismatchError(message string, cause error) types.Error {
	return NewError(codes.AlgorithmMismatch, message, cause)
}

// NewBindingInvalidError 创建绑定信息无效错误
func NewBindingInvalidError(message string, cause error) types.Error {
	return NewError(codes.BindingInvalid, message, cause)
}

// NewBindingMismatchError 创建绑定信息不匹配错误
func NewBindingMismatchError(message string, cause error) types.Error {
	return NewError(codes.BindingMismatch, message, cause)
}

// NewSessionInvalidError 创建会话无效错误
func NewSessionInvalidError(message string, cause error) types.Error {
	return NewError(codes.SessionInvalid, message, cause)
}

// NewSessionExpiredError 创建会话过期错误
func NewSessionExpiredError(message string, cause error) types.Error {
	return NewError(codes.SessionExpired, message, cause)
}

// NewSessionNotFoundError 创建会话不存在错误
func NewSessionNotFoundError(message string, cause error) types.Error {
	return NewError(codes.SessionNotFound, message, cause)
}

// NewPolicyViolationError 创建违反安全策略错误
func NewPolicyViolationError(message string, cause error) types.Error {
	return NewError(codes.PolicyViolation, message, cause)
}

// NewRotationFailedError 创建密钥轮换失败错误
func NewRotationFailedError(message string, cause error) types.Error {
	return NewError(codes.RotationFailed, message, cause)
}

// NewTokenGenerationError 创建Token生成错误
func NewTokenGenerationError(message string, cause error) types.Error {
	return NewError(codes.TokenGenerationError, message, cause)
}

// NewTokenBlacklistError 创建Token黑名单错误
func NewTokenBlacklistError(message string, cause error) types.Error {
	return NewError(codes.TokenBlacklistError, message, cause)
}

// IsTokenExpiredError 判断是否为Token过期错误
func IsTokenExpiredError(err error) bool {
	if e, ok := err.(types.Error); ok {
		return checkErrorCodeMapping(e.Code(), codes.TokenExpired)
	}
	return false
}

// IsSignatureInvalidError 判断是否为签名无效错误
func IsSignatureInvalidError(err error) bool {
	if e, ok := err.(types.Error); ok {
		return checkErrorCodeMapping(e.Code(), codes.SignatureInvalid)
	}
	return false
}

// IsTokenInvalidError 判断是否为Token无效错误
func IsTokenInvalidError(err error) bool {
	if e, ok := err.(types.Error); ok {
		return checkErrorCodeMapping(e.Code(), codes.TokenInvalid)
	}
	return false
}
