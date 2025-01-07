package unit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"gobase/pkg/errors"
	"gobase/pkg/middleware/jwt/extractor"
)

func TestCookieExtractor_Extract(t *testing.T) {
	// 设置测试用例
	tests := []struct {
		name        string
		cookieName  string
		setupCookie func(*http.Request)
		wantToken   string
		wantErr     error
	}{
		{
			name:       "成功提取token",
			cookieName: "jwt",
			setupCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{
					Name:  "jwt",
					Value: "valid.jwt.token",
				})
			},
			wantToken: "valid.jwt.token",
			wantErr:   nil,
		},
		{
			name:       "cookie不存在",
			cookieName: "jwt",
			setupCookie: func(r *http.Request) {
				// 不添加cookie
			},
			wantToken: "",
			wantErr:   errors.NewTokenNotFoundError("token cookie not found", nil),
		},
		{
			name:       "cookie值为空",
			cookieName: "jwt",
			setupCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{
					Name:  "jwt",
					Value: "",
				})
			},
			wantToken: "",
			wantErr:   errors.NewTokenInvalidError("token cookie is empty", nil),
		},
		{
			name:       "自定义cookie名称",
			cookieName: "custom_jwt",
			setupCookie: func(r *http.Request) {
				r.AddCookie(&http.Cookie{
					Name:  "custom_jwt",
					Value: "valid.jwt.token",
				})
			},
			wantToken: "valid.jwt.token",
			wantErr:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			req := httptest.NewRequest("GET", "/", nil)
			tt.setupCookie(req)
			c.Request = req

			e := extractor.NewCookieExtractor(tt.cookieName)

			// Act
			gotToken, err := e.Extract(c)

			// Assert
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantToken, gotToken)
			}
		})
	}
}

func TestNewCookieExtractor(t *testing.T) {
	// Arrange & Act
	e1 := extractor.NewCookieExtractor("")
	e2 := extractor.NewCookieExtractor("custom_jwt")

	// Assert
	assert.Equal(t, "jwt", e1.CookieName, "应该使用默认cookie名称")
	assert.Equal(t, "custom_jwt", e2.CookieName, "应该使用自定义cookie名称")
}
