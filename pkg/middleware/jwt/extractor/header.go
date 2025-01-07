package extractor

import (
	"strings"

	"gobase/pkg/errors"

	"github.com/gin-gonic/gin"
)

// HeaderExtractor 从请求头提取token
type HeaderExtractor struct {
	// Header 请求头名称
	Header string
	// Prefix token前缀(如Bearer)
	Prefix string
}

// NewHeaderExtractor 创建新的Header提取器
func NewHeaderExtractor(header, prefix string) *HeaderExtractor {
	if header == "" {
		header = "Authorization"
	}
	return &HeaderExtractor{
		Header: header,
		Prefix: prefix,
	}
}

// Extract 实现TokenExtractor接口
func (e *HeaderExtractor) Extract(c *gin.Context) (string, error) {
	// 从gin.Context获取header值
	header := c.GetHeader(e.Header)
	if header == "" {
		return "", errors.NewTokenNotFoundError("token not found in header", nil)
	}

	// 如果设置了前缀,验证并移除前缀
	if e.Prefix != "" {
		if !strings.HasPrefix(header, e.Prefix) {
			return "", errors.NewTokenInvalidError("invalid token prefix", nil)
		}
		header = strings.TrimPrefix(header, e.Prefix)
	}

	// 移除空格
	token := strings.TrimSpace(header)
	if token == "" {
		return "", errors.NewTokenInvalidError("token is empty", nil)
	}

	return token, nil
}
