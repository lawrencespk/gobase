package errors

import (
	"gobase/pkg/errors/codes"
)

// 数据库相关错误码 (2800-2899)

// NewDBConnError 创建数据库连接错误
func NewDBConnError(message string, cause error) error {
	return NewError(codes.DBConnError, message, cause)
}

// NewDBQueryError 创建数据库查询错误
func NewDBQueryError(message string, cause error) error {
	return NewError(codes.DBQueryError, message, cause)
}

// NewDBTransactionError 创建数据库事务错误
func NewDBTransactionError(message string, cause error) error {
	return NewError(codes.DBTransactionError, message, cause)
}

// NewDBDeadlockError 创建数据库死锁错误
func NewDBDeadlockError(message string, cause error) error {
	return NewError(codes.DBDeadlockError, message, cause)
}
