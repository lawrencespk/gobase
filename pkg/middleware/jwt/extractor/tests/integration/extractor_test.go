package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"gobase/pkg/middleware/jwt/extractor"
)

func setupRouter(e extractor.TokenExtractor) (*gin.Engine, func(c *gin.Context)) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// 处理函数
	handler := func(c *gin.Context) {
		token, err := e.Extract(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
	}

	r.GET("/test", handler)
	return r, handler
}

func TestExtractors_Integration(t *testing.T) {
	tests := []struct {
		name           string
		extractor      extractor.TokenExtractor
		setupRequest   func(*http.Request)
		expectedStatus int
		expectedToken  string
	}{
		{
			name:      "Header提取器-成功",
			extractor: extractor.NewHeaderExtractor("Authorization", "Bearer "),
			setupRequest: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer test.token.123")
			},
			expectedStatus: http.StatusOK,
			expectedToken:  "test.token.123",
		},
		{
			name:      "Cookie提取器-成功",
			extractor: extractor.NewCookieExtractor("jwt"),
			setupRequest: func(r *http.Request) {
				r.AddCookie(&http.Cookie{
					Name:  "jwt",
					Value: "test.token.456",
				})
			},
			expectedStatus: http.StatusOK,
			expectedToken:  "test.token.456",
		},
		{
			name:      "Query提取器-成功",
			extractor: extractor.NewQueryExtractor("token"),
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("token", "test.token.789")
				r.URL.RawQuery = q.Encode()
			},
			expectedStatus: http.StatusOK,
			expectedToken:  "test.token.789",
		},
		{
			name: "链式提取器-按顺序成功",
			extractor: extractor.ChainExtractor{
				extractor.NewHeaderExtractor("Authorization", "Bearer "),
				extractor.NewCookieExtractor("jwt"),
				extractor.NewQueryExtractor("token"),
			},
			setupRequest: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer test.token.chain")
			},
			expectedStatus: http.StatusOK,
			expectedToken:  "test.token.chain",
		},
		{
			name: "链式提取器-回退成功",
			extractor: extractor.ChainExtractor{
				extractor.NewHeaderExtractor("Authorization", "Bearer "),
				extractor.NewCookieExtractor("jwt"),
				extractor.NewQueryExtractor("token"),
			},
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("token", "test.token.fallback")
				r.URL.RawQuery = q.Encode()
			},
			expectedStatus: http.StatusOK,
			expectedToken:  "test.token.fallback",
		},
		{
			name: "链式提取器-全部失败",
			extractor: extractor.ChainExtractor{
				extractor.NewHeaderExtractor("Authorization", "Bearer "),
				extractor.NewCookieExtractor("jwt"),
				extractor.NewQueryExtractor("token"),
			},
			setupRequest: func(r *http.Request) {
				// 不设置任何token
			},
			expectedStatus: http.StatusUnauthorized,
			expectedToken:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			router, _ := setupRouter(tt.extractor)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			tt.setupRequest(req)

			// Act
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, response["token"])
			}
		})
	}
}
