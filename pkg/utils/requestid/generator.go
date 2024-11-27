package requestid

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Generate 生成唯一的请求ID
// 格式: prefix-timestamp-uuid
func Generate() string {
	timestamp := time.Now().Format("20060102150405")
	uid := uuid.New().String()[:8] // 只使用UUID的前8位
	return fmt.Sprintf("req-%s-%s", timestamp, uid)
}

// ValidateRequestID 验证请求ID格式是否正确
func ValidateRequestID(requestID string) bool {
	// TODO: 实现验证逻辑
	return true
}

// ParseRequestID 解析请求ID，返回其组成部分
func ParseRequestID(requestID string) (timestamp string, uuid string, ok bool) {
	// TODO: 实现解析逻辑
	return "", "", true
}
