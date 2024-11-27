package context

import (
	"fmt"
	"gobase/pkg/context/types"
)

var (
	// 常用的验证规则集
	RequiredUserContext  = []string{types.KeyUserID, types.KeyUserName}
	RequiredTraceContext = []string{types.KeyRequestID, types.KeyTraceID}
	RequiredBasicContext = []string{types.KeyRequestID, types.KeyClientIP}
)

// ValidateContext 验证上下文是否包含必要信息
func ValidateContext(ctx types.Context, required ...string) error {
	for _, key := range required {
		if _, ok := ctx.GetMetadata(key); !ok {
			return fmt.Errorf("missing required context key: %s", key)
		}
	}
	return nil
}

// ValidateUserContext 验证用户上下文
func ValidateUserContext(ctx types.Context) error {
	return ValidateContext(ctx, RequiredUserContext...)
}

// ValidateTraceContext 验证追踪上下文
func ValidateTraceContext(ctx types.Context) error {
	return ValidateContext(ctx, RequiredTraceContext...)
}
