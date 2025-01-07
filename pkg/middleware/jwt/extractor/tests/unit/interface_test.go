package unit

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"gobase/pkg/errors"
	"gobase/pkg/middleware/jwt/extractor"
)

// MockExtractor 模拟提取器
type MockExtractor struct {
	token string
	err   error
}

func (e *MockExtractor) Extract(c *gin.Context) (string, error) {
	return e.token, e.err
}

func TestChainExtractor_Extract(t *testing.T) {
	tests := []struct {
		name       string
		extractors extractor.ChainExtractor
		wantToken  string
		wantErr    error
	}{
		{
			name: "第一个提取器成功",
			extractors: extractor.ChainExtractor{
				&MockExtractor{token: "token1", err: nil},
				&MockExtractor{token: "token2", err: nil},
			},
			wantToken: "token1",
			wantErr:   nil,
		},
		{
			name: "第一个失败第二个成功",
			extractors: extractor.ChainExtractor{
				&MockExtractor{token: "", err: errors.NewTokenNotFoundError("not found", nil)},
				&MockExtractor{token: "token2", err: nil},
			},
			wantToken: "token2",
			wantErr:   nil,
		},
		{
			name: "全部提取器失败",
			extractors: extractor.ChainExtractor{
				&MockExtractor{token: "", err: errors.NewTokenNotFoundError("not found 1", nil)},
				&MockExtractor{token: "", err: errors.NewTokenNotFoundError("not found 2", nil)},
			},
			wantToken: "",
			wantErr:   errors.NewTokenNotFoundError("not found 2", nil),
		},
		{
			name:       "空提取器链",
			extractors: extractor.ChainExtractor{},
			wantToken:  "",
			wantErr:    errors.NewTokenNotFoundError("no token extractors configured", nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)

			// Act
			gotToken, err := tt.extractors.Extract(c)

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
