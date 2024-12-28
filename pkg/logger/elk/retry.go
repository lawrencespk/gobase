package elk

import (
	"context"
	"gobase/pkg/errors"
	"time"

	"github.com/sirupsen/logrus"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries  int           `yaml:"max_retries"`  // 最大重试次数（不包括初始尝试）
	InitialWait time.Duration `yaml:"initial_wait"` // 首次重试前的等待时间
	MaxWait     time.Duration `yaml:"max_wait"`     // 最大重试等待时间
}

// RetryableFunc 可重试的函数类型
type RetryableFunc func() error

// WithRetry 包装需要重试的操作
func WithRetry(ctx context.Context, config RetryConfig, operation RetryableFunc, logger logrus.FieldLogger) error {
	var err error
	attempts := 0
	wait := config.InitialWait

	logger.WithFields(logrus.Fields{
		"maxRetries":  config.MaxRetries,
		"maxAttempts": config.MaxRetries + 1,
	}).Debug("Starting retry mechanism")

	for attempts <= config.MaxRetries {
		logger.WithFields(logrus.Fields{
			"currentAttempt": attempts + 1,
			"maxAttempts":    config.MaxRetries + 1,
		}).Debug("Executing operation")

		select {
		case <-ctx.Done():
			return errors.NewSystemError("operation cancelled or timed out", ctx.Err())
		default:
		}

		err = operation()
		if err == nil {
			logger.WithField("finalAttempt", attempts+1).Info("operation succeeded")
			return nil
		}

		attempts++

		if attempts > config.MaxRetries {
			logger.WithFields(logrus.Fields{
				"finalAttempts": attempts,
				"maxRetries":    config.MaxRetries,
			}).Info("reached maximum retry limit")
			break
		}

		logger.WithFields(logrus.Fields{
			"attempts":   attempts,
			"maxRetries": config.MaxRetries,
			"error":      err,
		}).Info("operation failed")

		logger.WithFields(logrus.Fields{
			"attempt": attempts,
			"error":   err,
			"wait":    wait,
		}).Warn("operation failed, retrying...")

		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return errors.NewSystemError("operation cancelled or timed out during retry", ctx.Err())
		case <-timer.C:
		}

		wait *= 2
		if wait > config.MaxWait {
			wait = config.MaxWait
		}
	}

	return err
}
