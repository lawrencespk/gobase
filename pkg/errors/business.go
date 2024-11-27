package errors

import (
	"gobase/pkg/errors/codes"
)

// 业务通用错误 (2000-2099)
// NewInvalidParamsError 创建无效参数错误
func NewInvalidParamsError(message string, cause error) error {
	return NewError(codes.InvalidParams, message, cause)
}

// NewUnauthorizedError 创建未授权错误
func NewUnauthorizedError(message string, cause error) error {
	return NewError(codes.Unauthorized, message, cause)
}

// NewForbiddenError 创建禁止访问错误
func NewForbiddenError(message string, cause error) error {
	return NewError(codes.Forbidden, message, cause)
}

// NewNotFoundError 创建资源不存在错误
func NewNotFoundError(message string, cause error) error {
	return NewError(codes.NotFound, message, cause)
}

// NewAlreadyExistsError 创建资源已存在错误
func NewAlreadyExistsError(message string, cause error) error {
	return NewError(codes.AlreadyExists, message, cause)
}

// NewBadRequestError 创建请求错误
func NewBadRequestError(message string, cause error) error {
	return NewError(codes.BadRequest, message, cause)
}

// NewTooManyRequestsError 创建请求过多错误
func NewTooManyRequestsError(message string, cause error) error {
	return NewError(codes.TooManyRequests, message, cause)
}

// NewOperationFailedError 创建操作失败错误
func NewOperationFailedError(message string, cause error) error {
	return NewError(codes.OperationFailed, message, cause)
}

// NewDataConflictError 创建数据冲突错误
func NewDataConflictError(message string, cause error) error {
	return NewError(codes.DataConflict, message, cause)
}

// NewServiceUnavailableError 创建服务不可用错误
func NewServiceUnavailableError(message string, cause error) error {
	return NewError(codes.ServiceUnavailable, message, cause)
}

// NewRequestTimeoutError 创建请求超时错误
func NewRequestTimeoutError(message string, cause error) error {
	return NewError(codes.RequestTimeout, message, cause)
}

// NewInvalidTokenError 创建无效令牌错误
func NewInvalidTokenError(message string, cause error) error {
	return NewError(codes.InvalidToken, message, cause)
}

// NewTokenExpiredError 创建令牌过期错误
func NewTokenExpiredError(message string, cause error) error {
	return NewError(codes.TokenExpired, message, cause)
}

// NewInvalidSignatureError 创建无效签名错误
func NewInvalidSignatureError(message string, cause error) error {
	return NewError(codes.InvalidSignature, message, cause)
}

// 用户相关错误 (2100-2199)
func NewUserNotFoundError(message string, cause error) error {
	return NewError(codes.UserNotFound, message, cause)
}

// NewUserDisabledError 创建用户禁用错误
func NewUserDisabledError(message string, cause error) error {
	return NewError(codes.UserDisabled, message, cause)
}

// NewUserLockedError 创建用户锁定错误
func NewUserLockedError(message string, cause error) error {
	return NewError(codes.UserLocked, message, cause)
}

// NewInvalidPasswordError 创建无效密码错误
func NewInvalidPasswordError(message string, cause error) error {
	return NewError(codes.InvalidPassword, message, cause)
}

// NewPasswordExpiredError 创建密码过期错误
func NewPasswordExpiredError(message string, cause error) error {
	return NewError(codes.PasswordExpired, message, cause)
}

// NewInvalidUsernameError 创建无效用户名错误
func NewInvalidUsernameError(message string, cause error) error {
	return NewError(codes.InvalidUsername, message, cause)
}

// NewDuplicateUsernameError 创建用户名重复错误
func NewDuplicateUsernameError(message string, cause error) error {
	return NewError(codes.DuplicateUsername, message, cause)
}

// 权限相关错误 (2200-2299)
// NewNoPermissionError 创建无权限错误
func NewNoPermissionError(message string, cause error) error {
	return NewError(codes.NoPermission, message, cause)
}

// NewRoleNotFoundError 创建角色不存在错误
func NewRoleNotFoundError(message string, cause error) error {
	return NewError(codes.RoleNotFound, message, cause)
}

// NewInvalidRoleError 创建无效角色错误
func NewInvalidRoleError(message string, cause error) error {
	return NewError(codes.InvalidRole, message, cause)
}

// NewPermissionDeniedError 创建权限拒绝错误
func NewPermissionDeniedError(message string, cause error) error {
	return NewError(codes.PermissionDenied, message, cause)
}

// 文件操作错误 (2300-2399)
// NewFileNotFoundError 创建文件不存在错误
func NewFileNotFoundError(message string, cause error) error {
	return NewError(codes.FileNotFound, message, cause)
}

// NewFileUploadError 创建文件上传错误
func NewFileUploadError(message string, cause error) error {
	return NewError(codes.FileUploadError, message, cause)
}

// NewFileDownloadError 创建文件下载错误
func NewFileDownloadError(message string, cause error) error {
	return NewError(codes.FileDownloadError, message, cause)
}

// NewInvalidFileTypeError 创建无效文件类型错误
func NewInvalidFileTypeError(message string, cause error) error {
	return NewError(codes.InvalidFileType, message, cause)
}

// NewFileTooLargeError 创建文件过大错误
func NewFileTooLargeError(message string, cause error) error {
	return NewError(codes.FileTooLarge, message, cause)
}

// 通信相关错误 (2400-2499)
// NewMessageError 创建消息错误
func NewMessageError(message string, cause error) error {
	return NewError(codes.MessageError, message, cause)
}

// NewInvalidFormatError 创建无效格式错误
func NewInvalidFormatError(message string, cause error) error {
	return NewError(codes.InvalidFormat, message, cause)
}

// NewInvalidProtocolError 创建无效协议错误
func NewInvalidProtocolError(message string, cause error) error {
	return NewError(codes.InvalidProtocol, message, cause)
}

// NewEncryptionError 创建加密错误
func NewEncryptionError(message string, cause error) error {
	return NewError(codes.EncryptionError, message, cause)
}

// NewDecryptionError 创建解密错误
func NewDecryptionError(message string, cause error) error {
	return NewError(codes.DecryptionError, message, cause)
}
