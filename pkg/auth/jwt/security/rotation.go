package security

import (
	"context"
	"sync"
	"time"

	"gobase/pkg/auth/jwt/crypto"
	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metric"
)

// KeyRotator 密钥轮换管理器
type KeyRotator struct {
	keyManager *crypto.KeyManager
	policy     *Policy
	stopCh     chan struct{}
	mutex      sync.RWMutex
	logger     types.Logger
	metrics    *metric.Counter
}

// NewKeyRotator 创建密钥轮换管理器
func NewKeyRotator(keyManager *crypto.KeyManager, policy *Policy, logger types.Logger) (*KeyRotator, error) {
	if keyManager == nil {
		return nil, errors.NewConfigInvalidError("key manager is required", nil)
	}
	if policy == nil {
		return nil, errors.NewConfigInvalidError("policy is required", nil)
	}
	if logger == nil {
		return nil, errors.NewConfigInvalidError("logger is required", nil)
	}

	r := &KeyRotator{
		keyManager: keyManager,
		policy:     policy,
		stopCh:     make(chan struct{}),
		logger:     logger,
	}

	// 初始化监控指标
	r.metrics = metric.NewCounter(metric.CounterOpts{
		Namespace: "gobase",
		Subsystem: "jwt_security",
		Name:      "key_rotations_total",
		Help:      "Total number of key rotation operations",
	})

	return r, nil
}

// Start 启动密钥轮换
func (r *KeyRotator) Start(ctx context.Context) error {
	if !r.policy.EnableRotation {
		r.logger.Info(ctx, "key rotation is disabled")
		return nil
	}

	go r.rotationLoop(ctx)
	r.logger.Info(ctx, "key rotation started",
		types.Field{
			Key:   "interval",
			Value: r.policy.RotationInterval,
		},
	)
	return nil
}

// Stop 停止密钥轮换
func (r *KeyRotator) Stop() {
	close(r.stopCh)
	r.logger.Info(context.Background(), "key rotation stopped")
}

// rotationLoop 密钥轮换循环
func (r *KeyRotator) rotationLoop(ctx context.Context) {
	ticker := time.NewTicker(r.policy.RotationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := r.rotate(ctx); err != nil {
				r.logger.Error(ctx, "failed to rotate keys",
					types.Field{
						Key:   "error",
						Value: err,
					},
				)
			}
		case <-r.stopCh:
			return
		case <-ctx.Done():
			return
		}
	}
}

// rotate 执行密钥轮换
func (r *KeyRotator) rotate(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.metrics != nil {
		defer r.metrics.Inc()
	}

	if err := r.keyManager.RotateKeys(ctx); err != nil {
		return errors.NewRotationFailedError("failed to rotate keys", err)
	}

	r.logger.Info(ctx, "successfully rotated keys")
	return nil
}
