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

func TestQueryExtractor_Extract(t *testing.T) {
	tests := []struct {
		name       string
		paramName  string
		setupQuery func(*http.Request)
		wantToken  string
		wantErr    error
	}{
		{
			name:      "成功提取token",
			paramName: "token",
			setupQuery: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("token", "valid.jwt.token")
				r.URL.RawQuery = q.Encode()
			},
			wantToken: "valid.jwt.token",
			wantErr:   nil,
		},
		{
			name:      "参数不存在",
			paramName: "token",
			setupQuery: func(r *http.Request) {
				// 不添加查询参数
			},
			wantToken: "",
			wantErr:   errors.NewTokenNotFoundError("token not found in query parameters", nil),
		},
		{
			name:      "参数值为空",
			paramName: "token",
			setupQuery: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("token", "")
				r.URL.RawQuery = q.Encode()
			},
			wantToken: "",
			wantErr:   errors.NewTokenNotFoundError("token not found in query parameters", nil),
		},
		{
			name:      "自定义参数名",
			paramName: "jwt_token",
			setupQuery: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("jwt_token", "valid.jwt.token")
				r.URL.RawQuery = q.Encode()
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
			tt.setupQuery(req)
			c.Request = req

			e := extractor.NewQueryExtractor(tt.paramName)

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

func TestNewQueryExtractor(t *testing.T) {
	// Arrange & Act
	e1 := extractor.NewQueryExtractor("")
	e2 := extractor.NewQueryExtractor("jwt_token")

	// Assert
	assert.Equal(t, "token", e1.ParamName, "应该使用默认参数名")
	assert.Equal(t, "jwt_token", e2.ParamName, "应该使用自定义参数名")
}
