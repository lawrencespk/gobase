package binding

import (
	"context"
	"gobase/pkg/auth/jwt"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/types"
)

// ValidatorOptions 验证器配置选项
type ValidatorOptions struct {
	logger types.Logger
}

// ValidatorOption 验证器选项函数
type ValidatorOption func(*ValidatorOptions)

// WithLogger 设置日志记录器
func WithLogger(logger types.Logger) ValidatorOption {
	return func(opts *ValidatorOptions) {
		opts.logger = logger
	}
}

// defaultValidator 默认验证器实现
type defaultValidator struct {
	store  Store
	logger types.Logger
}

// NewValidator 创建新的验证器实例
func NewValidator(store Store, opts ...ValidatorOption) Validator {
	options := &ValidatorOptions{}
	for _, opt := range opts {
		opt(options)
	}

	return &defaultValidator{
		store:  store,
		logger: options.logger,
	}
}

// ValidateIP 实现 Validator 接口
func (v *defaultValidator) ValidateIP(ctx context.Context, claims jwt.Claims, currentIP string) error {
	tokenID := claims.GetTokenID()
	boundIP, err := v.store.GetIPBinding(ctx, tokenID)
	if err != nil {
		return err
	}

	if boundIP != currentIP {
		if v.logger != nil {
			v.logger.Warn(ctx, "IP binding mismatch",
				types.Field{Key: "token_id", Value: tokenID},
				types.Field{Key: "bound_ip", Value: boundIP},
				types.Field{Key: "current_ip", Value: currentIP},
			)
		}
		return errors.NewError(codes.BindingMismatch, "IP address mismatch", nil)
	}

	return nil
}

// ValidateDevice 实现 Validator 接口
func (v *defaultValidator) ValidateDevice(ctx context.Context, claims jwt.Claims, deviceInfo *DeviceInfo) error {
	tokenID := claims.GetTokenID()
	boundDevice, err := v.store.GetDeviceBinding(ctx, tokenID)
	if err != nil {
		return err
	}

	if !isDeviceMatch(boundDevice, deviceInfo) {
		if v.logger != nil {
			v.logger.Warn(ctx, "Device binding mismatch",
				types.Field{Key: "token_id", Value: tokenID},
				types.Field{Key: "bound_device", Value: boundDevice},
				types.Field{Key: "current_device", Value: deviceInfo},
			)
		}
		return errors.NewError(codes.BindingMismatch, "device mismatch", nil)
	}

	return nil
}

// isDeviceMatch 检查设备是否匹配
func isDeviceMatch(bound, current *DeviceInfo) bool {
	if bound == nil || current == nil {
		return false
	}
	return bound.ID == current.ID &&
		bound.Fingerprint == current.Fingerprint
}
