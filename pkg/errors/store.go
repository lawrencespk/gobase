package errors

import (
	"gobase/pkg/errors/codes"
	"gobase/pkg/errors/types"
)

// NewStoreNotFoundError 创建存储数据不存在错误
func NewStoreNotFoundError(message string, cause error) types.Error {
	return NewError(codes.StoreErrNotFound, message, cause)
}

// IsStoreNotFoundError 判断是否为存储数据不存在错误
func IsStoreNotFoundError(err error) bool {
	if e, ok := err.(types.Error); ok {
		return checkErrorCodeMapping(e.Code(), codes.StoreErrNotFound)
	}
	return false
}
