package extractor

import (
	"gobase/pkg/errors"

	"github.com/gin-gonic/gin"
)

// QueryExtractor 从URL查询参数提取token
type QueryExtractor struct {
	// ParamName 参数名称
	ParamName string
}

// NewQueryExtractor 创建新的Query提取器
func NewQueryExtractor(paramName string) *QueryExtractor {
	if paramName == "" {
		paramName = "token"
	}
	return &QueryExtractor{
		ParamName: paramName,
	}
}

// Extract 实现TokenExtractor接口
func (e *QueryExtractor) Extract(c *gin.Context) (string, error) {
	// 从gin.Context获取查询参数
	token := c.Query(e.ParamName)
	if token == "" {
		return "", errors.NewTokenNotFoundError("token not found in query parameters", nil)
	}
	return token, nil
}
