package extractor

import (
	"gobase/pkg/errors"

	"github.com/gin-gonic/gin"
)

// CookieExtractor 从Cookie提取token
type CookieExtractor struct {
	// CookieName Cookie名称
	CookieName string
}

// NewCookieExtractor 创建新的Cookie提取器
func NewCookieExtractor(cookieName string) *CookieExtractor {
	if cookieName == "" {
		cookieName = "jwt"
	}
	return &CookieExtractor{
		CookieName: cookieName,
	}
}

// Extract 实现TokenExtractor接口
func (e *CookieExtractor) Extract(c *gin.Context) (string, error) {
	// 从gin.Context获取cookie
	token, err := c.Cookie(e.CookieName)
	if err != nil {
		return "", errors.NewTokenNotFoundError("token cookie not found", nil)
	}

	if token == "" {
		return "", errors.NewTokenInvalidError("token cookie is empty", nil)
	}

	return token, nil
}
