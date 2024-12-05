package elk

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
}

// RetryableFunc 可重试的函数类型
type RetryableFunc func() error

// WithRetry 包装需要重试的操作
func WithRetry(ctx context.Context, config RetryConfig, operation RetryableFunc, logger logrus.FieldLogger) error {
	var err error
	wait := config.InitialWait

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err = operation(); err == nil {
				return nil
			}

			if attempt == config.MaxRetries {
				break
			}

			logger.WithFields(logrus.Fields{
				"attempt": attempt + 1,
				"error":   err,
				"wait":    wait,
			}).Warn("operation failed, retrying...")

			time.Sleep(wait)

			// 指数退避，但不超过最大等待时间
			wait *= 2
			if wait > config.MaxWait {
				wait = config.MaxWait
			}
		}
	}

	return err
}
