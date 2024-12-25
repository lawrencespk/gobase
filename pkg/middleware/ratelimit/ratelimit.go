package ratelimit

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metric"
	"gobase/pkg/ratelimit/core"
)

var (
	// RequestsTotal 记录所有限流请求
	RequestsTotal = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "ratelimit",
		Name:      "requests_total",
		Help:      "Total number of requests handled by rate limiter",
	}).WithLabels("key", "result") // result: allowed/rejected/error

	// RejectedTotal 记录被限流的请求
	RejectedTotal = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "ratelimit",
		Name:      "rejected_total",
		Help:      "Total number of requests rejected by rate limiter",
	}).WithLabels("key")

	// LimiterLatency 记录限流器处理延迟
	LimiterLatency = metric.NewHistogram(metric.HistogramOpts{
		Namespace: "gobase",
		Subsystem: "ratelimit",
		Name:      "latency_seconds",
		Help:      "Latency of rate limiter operations",
		Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}).WithLabels("key", "operation") // operation: allow/wait/total

	// WaitDuration 记录等待时间分布
	WaitDuration = metric.NewHistogram(metric.HistogramOpts{
		Namespace: "gobase",
		Subsystem: "ratelimit",
		Name:      "wait_duration_seconds",
		Help:      "Distribution of waiting time in wait mode",
		Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
	}).WithLabels("key")

	// RetryCount 记录重试次数分布
	RetryCount = metric.NewHistogram(metric.HistogramOpts{
		Namespace: "gobase",
		Subsystem: "ratelimit",
		Name:      "retry_count",
		Help:      "Distribution of retry attempts",
		Buckets:   []float64{0, 1, 2, 3, 4, 5},
	}).WithLabels("key", "result") // result: success/failure

	// ErrorsTotal 记录错误类型分布
	ErrorsTotal = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "ratelimit",
		Name:      "errors_total",
		Help:      "Total number of errors by type",
	}).WithLabels("key", "type") // type: timeout/connection/other
)

// ResponseFormatter 定义响应格式化接口
type ResponseFormatter interface {
	FormatError(code string, message string) interface{}
}

// DefaultResponseFormatter 默认的响应格式化器
type DefaultResponseFormatter struct{}

func (f *DefaultResponseFormatter) FormatError(code string, message string) interface{} {
	return gin.H{
		"code":    code,
		"message": message,
	}
}

// RetryStrategy 重试策略配置
type RetryStrategy struct {
	// 最大重试次数
	MaxAttempts int
	// 重试间隔
	RetryInterval time.Duration
	// 是否使用指数退避
	UseExponentialBackoff bool
	// 最大重试间隔（当使用指数退避时）
	MaxRetryInterval time.Duration
}

// ErrorMessages 自定义错误消息
type ErrorMessages struct {
	// 限流检查失败时的错误消息
	CheckFailedMessage string
	// 请求被限流时的错误消息
	LimitExceededMessage string
}

// StatusCodes 自定义状态码
type StatusCodes struct {
	// 限流检查失败时的状态码
	CheckFailed int
	// 请求被限流时的状态码
	LimitExceeded int
}

// Config 限流中间件配置
type Config struct {
	// 限流器实例
	Limiter core.Limiter
	// 限流键生成函数
	KeyFunc func(*gin.Context) string
	// 限流阈值
	Limit int64
	// 时间窗口
	Window time.Duration
	// 是否启用等待模式(true:等待直到允许通过, false:直接返回429)
	WaitMode bool
	// 等待模式下的超时时间
	WaitTimeout time.Duration
	// 响应格式化器
	Formatter ResponseFormatter
	// 重试策略
	Retry *RetryStrategy
	// 自定义错误消息
	Messages *ErrorMessages
	// 自定义状态码
	Status *StatusCodes
}

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		KeyFunc: func(c *gin.Context) string {
			return c.ClientIP() // 默认使用客户端IP作为限流键
		},
		Limit:       100,         // 默认100个请求
		Window:      time.Minute, // 默认1分钟窗口
		WaitMode:    false,       // 默认不等待
		WaitTimeout: time.Second * 5,
		Formatter:   &DefaultResponseFormatter{},
		Retry: &RetryStrategy{
			MaxAttempts:           3,
			RetryInterval:         time.Millisecond * 100,
			UseExponentialBackoff: true,
			MaxRetryInterval:      time.Second * 2,
		},
		Messages: &ErrorMessages{
			CheckFailedMessage:   "rate limit check failed",
			LimitExceededMessage: "too many requests",
		},
		Status: &StatusCodes{
			CheckFailed:   http.StatusInternalServerError,
			LimitExceeded: http.StatusTooManyRequests,
		},
		// Limiter 必须由用户提供，因为它需要依赖外部存储
	}
}

// doWithRetry 执行带重试的操作
func doWithRetry(c *gin.Context, key string, strategy *RetryStrategy, operation func() (bool, error)) (bool, error) {
	if strategy == nil {
		return operation()
	}

	log := logger.GetLogger().WithFields(
		types.Field{Key: "module", Value: "middleware"},
		types.Field{Key: "component", Value: "ratelimit"},
		types.Field{Key: "key", Value: key},
	)

	var lastErr error
	interval := strategy.RetryInterval
	attempts := 0
	start := time.Now()

	for attempt := 0; attempt < strategy.MaxAttempts; attempt++ {
		attempts++
		operationStart := time.Now()
		allowed, err := operation()

		log.Debug(c, "rate limit operation attempt",
			types.Field{Key: "attempt", Value: attempts},
			types.Field{Key: "allowed", Value: allowed},
			types.Field{Key: "error", Value: err},
			types.Field{Key: "duration", Value: time.Since(operationStart)},
		)

		if err == nil && allowed {
			RetryCount.WithLabelValues(key, "success").Observe(float64(attempts))
			log.Info(c, "rate limit operation succeeded after retries",
				types.Field{Key: "attempts", Value: attempts},
				types.Field{Key: "total_duration", Value: time.Since(start)},
			)
			return true, nil
		}

		if err != nil {
			lastErr = err
			errType := "other"
			if err == context.DeadlineExceeded {
				errType = "timeout"
				ErrorsTotal.WithLabelValues(key, "timeout").Inc()
			} else if strings.Contains(err.Error(), "connection") {
				errType = "connection"
				ErrorsTotal.WithLabelValues(key, "connection").Inc()
			} else {
				ErrorsTotal.WithLabelValues(key, "other").Inc()
			}

			log.Warn(c, "rate limit operation failed",
				types.Field{Key: "attempt", Value: attempts},
				types.Field{Key: "error_type", Value: errType},
				types.Field{Key: "error", Value: err},
			)
		}

		if attempt == strategy.MaxAttempts-1 {
			break
		}

		nextInterval := interval
		if strategy.UseExponentialBackoff {
			nextInterval = time.Duration(float64(interval) * 2)
			if nextInterval > strategy.MaxRetryInterval {
				nextInterval = strategy.MaxRetryInterval
			}
		}

		log.Debug(c, "scheduling retry",
			types.Field{Key: "attempt", Value: attempts},
			types.Field{Key: "next_interval", Value: nextInterval},
		)

		select {
		case <-c.Request.Context().Done():
			RetryCount.WithLabelValues(key, "failure").Observe(float64(attempts))
			log.Warn(c, "rate limit operation cancelled",
				types.Field{Key: "attempts", Value: attempts},
				types.Field{Key: "total_duration", Value: time.Since(start)},
			)
			return false, c.Request.Context().Err()
		case <-time.After(nextInterval):
			interval = nextInterval
			continue
		}
	}

	RetryCount.WithLabelValues(key, "failure").Observe(float64(attempts))
	log.Error(c, "rate limit operation failed after all retries",
		types.Field{Key: "attempts", Value: attempts},
		types.Field{Key: "total_duration", Value: time.Since(start)},
		types.Field{Key: "error", Value: lastErr},
	)

	if lastErr != nil {
		return false, lastErr
	}
	return false, nil
}

// RateLimit 创建限流中间件
func RateLimit(cfg *Config) gin.HandlerFunc {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// 验证必要的配置
	if cfg.Limiter == nil {
		panic("ratelimit middleware requires a limiter instance")
	}

	// 确保有默认的KeyFunc
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = DefaultConfig().KeyFunc
	}

	// 确保有默认的Formatter
	if cfg.Formatter == nil {
		cfg.Formatter = DefaultConfig().Formatter
	}

	// 确保有默认的Messages
	if cfg.Messages == nil {
		cfg.Messages = DefaultConfig().Messages
	}

	// 确保有默认的Status
	if cfg.Status == nil {
		cfg.Status = DefaultConfig().Status
	}

	// 确保有默认的WaitTimeout
	if cfg.WaitTimeout <= 0 {
		cfg.WaitTimeout = DefaultConfig().WaitTimeout
	}

	return func(c *gin.Context) {
		start := time.Now()
		key := cfg.KeyFunc(c)

		log := logger.GetLogger().WithFields(
			types.Field{Key: "module", Value: "middleware"},
			types.Field{Key: "component", Value: "ratelimit"},
			types.Field{Key: "key", Value: key},
			types.Field{Key: "limit", Value: cfg.Limit},
			types.Field{Key: "window", Value: cfg.Window},
			types.Field{Key: "wait_mode", Value: cfg.WaitMode},
		)

		log.Debug(c, "starting rate limit check")

		var allowed bool
		var err error

		operation := func() (bool, error) {
			operationStart := time.Now()
			if cfg.WaitMode {
				waitStart := time.Now()
				ctx, cancel := context.WithTimeout(c.Request.Context(), cfg.WaitTimeout)
				defer cancel()

				err := cfg.Limiter.Wait(ctx, key, cfg.Limit, cfg.Window)
				waitDuration := time.Since(waitStart)
				WaitDuration.WithLabelValues(key).Observe(waitDuration.Seconds())

				log.Debug(c, "wait operation completed",
					types.Field{Key: "duration", Value: waitDuration},
					types.Field{Key: "error", Value: err},
				)

				if err == context.DeadlineExceeded {
					return false, nil
				}
				return err == nil, err
			}

			allowed, err := cfg.Limiter.Allow(c, key, cfg.Limit, cfg.Window)
			log.Debug(c, "allow operation completed",
				types.Field{Key: "duration", Value: time.Since(operationStart)},
				types.Field{Key: "allowed", Value: allowed},
				types.Field{Key: "error", Value: err},
			)
			return allowed, err
		}

		allowed, err = doWithRetry(c, key, cfg.Retry, operation)
		duration := time.Since(start)
		LimiterLatency.WithLabelValues(key, "total").Observe(duration.Seconds())

		if err != nil {
			log.Error(c, "rate limit check failed",
				types.Field{Key: "error", Value: err},
				types.Field{Key: "duration", Value: duration},
			)

			RequestsTotal.WithLabelValues(key, "error").Inc()
			err = errors.NewError(codes.RateLimitError, cfg.Messages.CheckFailedMessage, err)
			c.Error(err)
			c.AbortWithStatusJSON(cfg.Status.CheckFailed,
				cfg.Formatter.FormatError(codes.RateLimitError, cfg.Messages.CheckFailedMessage))
			return
		}

		if !allowed {
			log.Info(c, "rate limit exceeded",
				types.Field{Key: "duration", Value: duration},
			)

			RequestsTotal.WithLabelValues(key, "rejected").Inc()
			RejectedTotal.WithLabelValues(key).Inc()
			err = errors.NewError(codes.TooManyRequests, cfg.Messages.LimitExceededMessage, nil)
			c.Error(err)
			c.AbortWithStatusJSON(cfg.Status.LimitExceeded,
				cfg.Formatter.FormatError(codes.TooManyRequests, cfg.Messages.LimitExceededMessage))
			return
		}

		log.Info(c, "request allowed",
			types.Field{Key: "duration", Value: duration},
		)

		RequestsTotal.WithLabelValues(key, "allowed").Inc()
		c.Next()
	}
}
