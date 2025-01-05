package crypto

import (
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/errors/types"
)

// NewKeyInvalidError 创建密钥无效错误
func NewKeyInvalidError(message string, cause error) types.Error {
	return errors.NewError(codes.KeyInvalid, message, cause)
}

// NewAlgorithmMismatchError 创建算法不匹配错误
func NewAlgorithmMismatchError(message string, cause error) types.Error {
	return errors.NewError(codes.AlgorithmMismatch, message, cause)
}

// NewSignatureInvalidError 创建签名无效错误
func NewSignatureInvalidError(message string, cause error) types.Error {
	return errors.NewError(codes.SignatureInvalid, message, cause)
}

// NewRotationFailedError 创建密钥轮换失败错误
func NewRotationFailedError(message string, cause error) types.Error {
	return errors.NewError(codes.RotationFailed, message, cause)
}

// IsKeyInvalidError 判断是否为密钥无效错误
func IsKeyInvalidError(err error) bool {
	return errors.Is(err, errors.NewError(codes.KeyInvalid, "", nil))
}

// IsAlgorithmMismatchError 判断是否为算法不匹配错误
func IsAlgorithmMismatchError(err error) bool {
	return errors.Is(err, errors.NewError(codes.AlgorithmMismatch, "", nil))
}

// IsSignatureInvalidError 判断是否为签名无效错误
func IsSignatureInvalidError(err error) bool {
	return errors.Is(err, errors.NewError(codes.SignatureInvalid, "", nil))
}

// IsRotationFailedError 判断是否为密钥轮换失败错误
func IsRotationFailedError(err error) bool {
	return errors.Is(err, errors.NewError(codes.RotationFailed, "", nil))
}
