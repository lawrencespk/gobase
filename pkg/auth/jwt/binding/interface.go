package binding

import (
	"context"
	"gobase/pkg/auth/jwt"
)

// Validator 绑定验证器接口
type Validator interface {
	// ValidateIP 验证IP绑定
	ValidateIP(ctx context.Context, claims jwt.Claims, currentIP string) error

	// ValidateDevice 验证设备绑定
	ValidateDevice(ctx context.Context, claims jwt.Claims, deviceInfo *DeviceInfo) error
}

// Store 绑定存储接口
type Store interface {
	// SaveIPBinding 保存IP绑定
	SaveIPBinding(ctx context.Context, userID, tokenID, ip string) error

	// SaveDeviceBinding 保存设备绑定
	SaveDeviceBinding(ctx context.Context, userID, tokenID string, device *DeviceInfo) error

	// GetIPBinding 获取IP绑定
	GetIPBinding(ctx context.Context, tokenID string) (string, error)

	// GetDeviceBinding 获取设备绑定
	GetDeviceBinding(ctx context.Context, tokenID string) (*DeviceInfo, error)

	// DeleteBinding 删除绑定
	DeleteBinding(ctx context.Context, tokenID string) error

	// Close 关闭存储
	Close() error
}

// DeviceInfo 设备信息
type DeviceInfo struct {
	// 设备ID
	ID string `json:"id"`
	// 设备类型
	Type string `json:"type"`
	// 设备名称
	Name string `json:"name"`
	// 操作系统
	OS string `json:"os"`
	// 浏览器
	Browser string `json:"browser"`
	// 设备指纹
	Fingerprint string `json:"fingerprint"`
}
