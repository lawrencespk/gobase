package validator

import (
	"gobase/pkg/auth/jwt"
	"gobase/pkg/errors"

	"github.com/gin-gonic/gin"
)

// BindingValidator 设备绑定验证器
type BindingValidator struct {
	// 是否验证设备ID
	validateDeviceID bool
	// 是否验证IP地址
	validateIPAddress bool
}

// NewBindingValidator 创建新的绑定验证器
func NewBindingValidator(opts ...BindingValidatorOption) *BindingValidator {
	v := &BindingValidator{
		validateDeviceID:  true,
		validateIPAddress: true,
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

// BindingValidatorOption 定义绑定验证器选项
type BindingValidatorOption func(*BindingValidator)

// WithDeviceIDValidation 设置是否验证设备ID
func WithDeviceIDValidation(validate bool) BindingValidatorOption {
	return func(v *BindingValidator) {
		v.validateDeviceID = validate
	}
}

// WithIPAddressValidation 设置是否验证IP地址
func WithIPAddressValidation(validate bool) BindingValidatorOption {
	return func(v *BindingValidator) {
		v.validateIPAddress = validate
	}
}

// Validate 实现TokenValidator接口
func (v *BindingValidator) Validate(c *gin.Context, claims jwt.Claims) error {
	// 验证设备ID
	if v.validateDeviceID {
		deviceID := claims.GetDeviceID()
		if deviceID == "" {
			return errors.NewBindingInvalidError("device id is required", nil)
		}
		// TODO: 可以添加更多设备ID验证逻辑
	}

	// 验证IP地址
	if v.validateIPAddress {
		ipAddress := claims.GetIPAddress()
		if ipAddress == "" {
			return errors.NewBindingInvalidError("ip address is required", nil)
		}
		clientIP := c.ClientIP()
		if ipAddress != clientIP {
			return errors.NewBindingMismatchError("ip address mismatch", nil)
		}
	}

	return nil
}

func (v *BindingValidator) ValidateDeviceID() bool {
	return v.validateDeviceID
}

func (v *BindingValidator) ValidateIPAddress() bool {
	return v.validateIPAddress
}
