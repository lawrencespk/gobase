package types

import (
	"context"
	"net/http"
)

// Blacklist 定义黑名单接口
type Blacklist interface {
	IsBlacklisted(ctx context.Context, token string) bool
}

// DeviceValidator 定义设备验证接口
type DeviceValidator interface {
	Validate(ctx context.Context, token string, headers http.Header) error
}

// IPValidator 定义IP验证接口
type IPValidator interface {
	Validate(ctx context.Context, token string, remoteAddr string) error
}

// SessionManager 定义会话管理接口
type SessionManager interface {
	ValidateSession(ctx context.Context, claims Claims) error
}

// SecurityChecker 定义安全检查接口
type SecurityChecker interface {
	Check(ctx context.Context, token string) error
}

// EventBus 定义事件总线接口
type EventBus interface {
	Publish(ctx context.Context, topic string, data interface{}) error
}

// Claims 定义JWT声明接口
type Claims interface {
	GetSubject() (string, error)
}
