package extractor

import (
	"gobase/pkg/errors"

	"github.com/gin-gonic/gin"
)

// TokenExtractor 定义token提取器接口
type TokenExtractor interface {
	// Extract 从gin.Context中提取token
	Extract(c *gin.Context) (string, error)
}

// ExtractorFunc 定义token提取器函数类型
type ExtractorFunc func(c *gin.Context) (string, error)

// Extract 实现TokenExtractor接口
func (f ExtractorFunc) Extract(c *gin.Context) (string, error) {
	return f(c)
}

// ChainExtractor 链式提取器,按顺序尝试多个提取器
type ChainExtractor []TokenExtractor

// Extract 实现TokenExtractor接口
func (e ChainExtractor) Extract(c *gin.Context) (string, error) {
	if len(e) == 0 {
		return "", errors.NewTokenNotFoundError("no token extractors configured", nil)
	}

	var lastErr error
	for _, extractor := range e {
		token, err := extractor.Extract(c)
		if err == nil {
			return token, nil
		}
		lastErr = err
	}

	// 直接返回最后一个提取器的错误，而不是包装它
	return "", lastErr
}
