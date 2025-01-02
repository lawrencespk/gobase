package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
	"gobase/pkg/trace/jaeger"

	"github.com/go-redis/redis/v8"
)

// contextKey 是上下文键类型
type contextKey string

// clientKey 是Redis客户端在上下文中的键
const clientKey contextKey = "redis_client"

// withOperation 统一的操作包装函数
func (c *client) withOperation(ctx context.Context, operation string, fn func() error) error {
	var span *jaeger.Span
	var err error

	// 只在启用了追踪且tracer不为空时创建span
	if c.options.EnableTracing && c.tracer != nil {
		span, err = jaeger.NewSpan(
			"redis."+operation,
			jaeger.WithParent(ctx),
			jaeger.WithTag("db.type", "redis"),
			jaeger.WithTag("db.operation", operation),
		)
		if err != nil {
			c.logger.WithError(err).Error(ctx, "failed to create span")
		} else {
			defer span.Finish()
			ctx = span.Context()
		}
	}

	// 将客户端实例添加到上下文中
	ctx = context.WithValue(ctx, clientKey, c)

	c.logger.WithFields(
		types.Field{Key: "operation", Value: operation},
		types.Field{Key: "client_id", Value: fmt.Sprintf("%p", c)},
	).Debug(ctx, "starting redis operation")

	// 检查 Collector 而不是 Registry
	var metrics *pipelineMetrics
	if c.options.EnableMetrics && c.options.Collector != nil {
		metrics = newPipelineMetrics(c.options.MetricsNamespace)
	}

	// 使用正确的 metrics 类型调用 withMetrics
	if metrics != nil {
		return withMetrics(ctx, operation, metrics, fn)
	}

	// 如果未启用指标收集，直接执行函数
	err = fn()
	if err != nil {
		// 如果有span且发生错误，记录错误
		if span != nil {
			span.SetError(err)
		}

		// 1. 如果是上下文超时，返回超时错误
		if err == context.DeadlineExceeded ||
			strings.Contains(err.Error(), "context deadline exceeded") {
			return errors.NewTimeoutError("operation timed out", err)
		}

		// 2. 如果已经是包装过的错误，直接返回
		if errors.HasErrorCode(err, "") {
			return err
		}

		// 3. 处理特定的 Redis 错误
		if strings.Contains(err.Error(), "NOAUTH") || strings.Contains(err.Error(), "AUTH") {
			return errors.NewRedisAuthError("authentication failed", err)
		}

		// 处理连接池错误
		if strings.Contains(err.Error(), "pool exhausted") {
			return errors.NewRedisPoolExhaustedError("connection pool exhausted", err)
		}

		// 处理连接错误
		if strings.Contains(err.Error(), "connection refused") ||
			strings.Contains(err.Error(), "network is down") {
			return errors.NewRedisConnError("connection failed", err)
		}

		// 处理只读错误
		if isReadOnlyError(err) {
			return errors.NewRedisReadOnlyError("redis instance is read-only", err)
		}

		// 处理集群错误
		if isClusterDownError(err) {
			return errors.NewRedisClusterError("cluster is down", err)
		}

		// 处理加载错误
		if isLoadingError(err) {
			return errors.NewRedisLoadingError("redis is loading the dataset in memory", err)
		}
	}
	return err
}

// withMetrics 包装Redis操作并记录监控指标
func withMetrics(ctx context.Context, operation string, metrics *pipelineMetrics, fn func() error) error {
	startTime := time.Now()
	err := fn()
	duration := time.Since(startTime)

	// 记录执行时间，添加操作类型标签
	metrics.executionLatency.WithLabelValues(operation).Observe(duration.Seconds())

	// 记录错误计数，添加操作类型和错误类型标签
	if err != nil {
		errorType := "unknown"
		if errors.HasErrorCode(err, "") {
			errorType = errors.GetErrorCode(err)
		}
		metrics.errorTotal.WithLabelValues(operation, errorType).Inc()
	}

	// 记录命令计数，添加操作类型标签
	metrics.commandsTotal.WithLabelValues(operation).Inc()

	// 记录到日志（如果上下文中有 logger）
	if logger, ok := ctx.Value("logger").(types.Logger); ok {
		logger.WithFields(
			types.Field{Key: "operation", Value: operation},
			types.Field{Key: "duration_ms", Value: duration.Milliseconds()},
			types.Field{Key: "error", Value: err != nil},
		).Debug(ctx, "redis operation metrics")
	}

	return err
}

// withOperationResult 用于处理有返回值的Redis操作
func (c *client) withOperationResult(ctx context.Context, operation string, fn func() (interface{}, error)) (interface{}, error) {
	var result interface{}
	err := c.withOperation(ctx, operation, func() error {
		var err error
		result, err = fn()
		if err != nil {
			// 处理键不存在的情况
			if err == redis.Nil {
				return errors.NewRedisKeyNotFoundError("key not found", err)
			}
			return err
		}
		return nil
	})
	return result, err
}
