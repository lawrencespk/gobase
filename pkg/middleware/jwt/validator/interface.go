package validator

import (
	"github.com/gin-gonic/gin"

	"gobase/pkg/auth/jwt"
)

// TokenValidator 定义token验证器接口
type TokenValidator interface {
	// Validate 验证token
	Validate(c *gin.Context, claims jwt.Claims) error
}

// ValidatorFunc 定义token验证器函数类型
type ValidatorFunc func(c *gin.Context, claims jwt.Claims) error

// Validate 实现TokenValidator接口
func (f ValidatorFunc) Validate(c *gin.Context, claims jwt.Claims) error {
	return f(c, claims)
}

// ChainValidator 链式验证器,按顺序执行多个验证器
type ChainValidator []TokenValidator

// Validate 实现TokenValidator接口
func (v ChainValidator) Validate(c *gin.Context, claims jwt.Claims) error {
	for _, validator := range v {
		if err := validator.Validate(c, claims); err != nil {
			return err
		}
	}
	return nil
}
