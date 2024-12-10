package context

import (
	"fmt"
	"gobase/pkg/context/types"
	"reflect"
)

var (
	// 常用的验证规则集
	RequiredUserContext  = []string{types.KeyUserID, types.KeyUserName}
	RequiredTraceContext = []string{types.KeyRequestID, types.KeyTraceID}
	RequiredBasicContext = []string{types.KeyRequestID, types.KeyClientIP}
)

// ValidateContext 验证上下文是否包含必要信息
func ValidateContext(ctx types.Context, required ...string) error {
	metadata := ctx.GetMetadata()
	for _, key := range required {
		if _, ok := metadata[key]; !ok {
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

// ValidateMetadata 验证元数据
func ValidateMetadata(ctx types.Context, key string, expectedType reflect.Type) error {
	// 获取元数据值
	value := ctx.GetValue(key)
	if value == nil {
		return fmt.Errorf("metadata key %s not found", key)
	}

	// 验证类型
	actualType := reflect.TypeOf(value)
	if actualType != expectedType {
		return fmt.Errorf("metadata key %s has wrong type, expected %v but got %v",
			key, expectedType, actualType)
	}

	return nil
}

// ValidateRequiredMetadata 验证必需的元数据
func ValidateRequiredMetadata(ctx types.Context, keys ...string) error {
	for _, key := range keys {
		if value := ctx.GetValue(key); value == nil {
			return fmt.Errorf("required metadata key %s not found", key)
		}
	}
	return nil
}

// ValidateMetadataType 验证元数据类型
func ValidateMetadataType(ctx types.Context, validations map[string]reflect.Type) error {
	for key, expectedType := range validations {
		value := ctx.GetValue(key)
		if value == nil {
			return fmt.Errorf("metadata key %s not found", key)
		}

		actualType := reflect.TypeOf(value)
		if actualType != expectedType {
			return fmt.Errorf("metadata key %s has wrong type, expected %v but got %v",
				key, expectedType, actualType)
		}
	}
	return nil
}
