package errors

import (
	"errors"
	"gobase/pkg/errors/types"
)

// ErrorChain 表示错误链
type ErrorChain []types.Error

// GetErrorChain 获取完整的错误链
func GetErrorChain(err error) ErrorChain {
	if err == nil {
		return nil
	}

	chain := make(ErrorChain, 0)
	current := err

	for current != nil {
		if customErr, ok := current.(types.Error); ok {
			chain = append(chain, customErr)
		}
		current = errors.Unwrap(current)
	}

	return chain
}

// FirstError 获取错误链中的第一个错误
func FirstError(err error) types.Error {
	chain := GetErrorChain(err)
	if len(chain) > 0 {
		return chain[0]
	}
	return nil
}

// LastError 获取错误链中的最后一个错误
func LastError(err error) types.Error {
	chain := GetErrorChain(err)
	if len(chain) > 0 {
		return chain[len(chain)-1]
	}
	return nil
}

// RootCause 获取错误的根本原因
func RootCause(err error) error {
	current := err
	for current != nil {
		next := errors.Unwrap(current)
		if next == nil {
			return current
		}
		current = next
	}
	return err
}
