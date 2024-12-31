package redis

import (
	"context"
	"fmt"
	"gobase/pkg/logger/types"
	"strings"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

// withOperation 统一的操作包装函数
func (c *client) withOperation(ctx context.Context, operation string, fn func() error) error {
	span, ctx := startSpan(ctx, c.tracer, "redis."+operation)
	defer span.Finish()

	// 将客户端实例添加到上下文中
	ctx = context.WithValue(ctx, clientKey, c)

	c.logger.WithFields(
		types.Field{Key: "operation", Value: operation},
		types.Field{Key: "client_id", Value: fmt.Sprintf("%p", c)},
	).Debug(ctx, "starting redis operation")

	err := withMetrics(ctx, operation, c.options.Registry, fn)
	if err != nil {
		// 统一处理上下文超时错误
		if err == context.DeadlineExceeded {
			return errors.NewError(codes.TimeoutError, "operation timed out", err)
		}
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return errors.NewError(codes.TimeoutError, "operation timed out", err)
		}
	}
	return err
}

// withOperationResult 用于处理有返回值的Redis操作
func (c *client) withOperationResult(ctx context.Context, operation string, fn func() (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := c.withOperation(ctx, operation, func() error {
		var err error
		result, err = fn()
		return err
	})
	// withOperation 已经处理了超时错误，这里直接返回
	return result, err
}
