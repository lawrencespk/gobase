package errors

import (
	"gobase/pkg/errors/codes"
)

// NewDataAccessError 创建数据访问错误
func NewDataAccessError(message string, cause error) error {
	return NewError(codes.DataAccessError, message, cause)
}

// NewDataCreateError 创建数据创建错误
func NewDataCreateError(message string, cause error) error {
	return NewError(codes.DataCreateError, message, cause)
}

// NewDataUpdateError 创建数据更新错误
func NewDataUpdateError(message string, cause error) error {
	return NewError(codes.DataUpdateError, message, cause)
}

// NewDataDeleteError 创建数据删除错误
func NewDataDeleteError(message string, cause error) error {
	return NewError(codes.DataDeleteError, message, cause)
}

// NewDataQueryError 创建数据查询错误
func NewDataQueryError(message string, cause error) error {
	return NewError(codes.DataQueryError, message, cause)
}

// NewDataConvertError 创建数据转换错误
func NewDataConvertError(message string, cause error) error {
	return NewError(codes.DataConvertError, message, cause)
}

// NewDataValidateError 创建数据验证错误
func NewDataValidateError(message string, cause error) error {
	return NewError(codes.DataValidateError, message, cause)
}

// NewDataCorruptError 创建数据损坏错误
func NewDataCorruptError(message string, cause error) error {
	return NewError(codes.DataCorruptError, message, cause)
}
