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

func TestHeaderExtractor_Extract(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		prefix      string
		setupHeader func(*http.Request)
		wantToken   string
		wantErr     error
	}{
		{
			name:   "成功提取token",
			header: "Authorization",
			prefix: "Bearer ",
			setupHeader: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer valid.jwt.token")
			},
			wantToken: "valid.jwt.token",
			wantErr:   nil,
		},
		{
			name:   "header不存在",
			header: "Authorization",
			prefix: "Bearer ",
			setupHeader: func(r *http.Request) {
				// 不设置header
			},
			wantToken: "",
			wantErr:   errors.NewTokenNotFoundError("token not found in header", nil),
		},
		{
			name:   "前缀不匹配",
			header: "Authorization",
			prefix: "Bearer ",
			setupHeader: func(r *http.Request) {
				r.Header.Set("Authorization", "Basic valid.jwt.token")
			},
			wantToken: "",
			wantErr:   errors.NewTokenInvalidError("invalid token prefix", nil),
		},
		{
			name:   "token为空",
			header: "Authorization",
			prefix: "Bearer ",
			setupHeader: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer ")
			},
			wantToken: "",
			wantErr:   errors.NewTokenInvalidError("token is empty", nil),
		},
		{
			name:   "自定义header",
			header: "X-JWT-Token",
			prefix: "",
			setupHeader: func(r *http.Request) {
				r.Header.Set("X-JWT-Token", "valid.jwt.token")
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
			tt.setupHeader(req)
			c.Request = req

			e := extractor.NewHeaderExtractor(tt.header, tt.prefix)

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

func TestNewHeaderExtractor(t *testing.T) {
	// Arrange & Act
	e1 := extractor.NewHeaderExtractor("", "")
	e2 := extractor.NewHeaderExtractor("X-JWT-Token", "JWT ")

	// Assert
	assert.Equal(t, "Authorization", e1.Header, "应该使用默认header名称")
	assert.Equal(t, "", e1.Prefix, "应该使用空前缀")
	assert.Equal(t, "X-JWT-Token", e2.Header, "应该使用自定义header名称")
	assert.Equal(t, "JWT ", e2.Prefix, "应该使用自定义前缀")
}
