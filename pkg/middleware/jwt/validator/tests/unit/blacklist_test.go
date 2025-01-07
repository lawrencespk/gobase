package unit

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/errors"
	"gobase/pkg/middleware/jwt/validator"
)

// mockBlacklist 模拟黑名单实现
type mockBlacklist struct {
	mock.Mock
}

// 修正 Add 方法签名
func (m *mockBlacklist) Add(ctx context.Context, token string, expiration time.Duration) error {
	args := m.Called(ctx, token, expiration)
	return args.Error(0)
}

func (m *mockBlacklist) Remove(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockBlacklist) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	args := m.Called(ctx, token)
	return args.Bool(0), args.Error(1)
}

// 添加 Clear 方法实现接口
func (m *mockBlacklist) Clear(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// 添加 Close 方法实现接口
func (m *mockBlacklist) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestBlacklistValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mockBlacklist)
		setupCtx  func(*gin.Context)
		claims    jwt.Claims
		wantErr   bool
		checkErr  func(error) bool
	}{
		{
			name: "验证成功-token不在黑名单中",
			setupMock: func(m *mockBlacklist) {
				m.On("IsBlacklisted", mock.Anything, "valid.token").Return(false, nil)
			},
			setupCtx: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Set("jwt_token", "valid.token")
			},
			claims:  newTestClaims("test-user", "test-device", "127.0.0.1"),
			wantErr: false,
		},
		{
			name: "token在黑名单中",
			setupMock: func(m *mockBlacklist) {
				m.On("IsBlacklisted", mock.Anything, "revoked.token").Return(true, nil)
			},
			setupCtx: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Set("jwt_token", "revoked.token")
			},
			claims:  newTestClaims("test-user", "test-device", "127.0.0.1"),
			wantErr: true,
			checkErr: func(err error) bool {
				return errors.Is(err, errors.NewTokenRevokedError("", nil))
			},
		},
		{
			name: "黑名单检查失败",
			setupMock: func(m *mockBlacklist) {
				m.On("IsBlacklisted", mock.Anything, "error.token").Return(false, assert.AnError)
			},
			setupCtx: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Set("jwt_token", "error.token")
			},
			claims:  newTestClaims("test-user", "test-device", "127.0.0.1"),
			wantErr: true,
			checkErr: func(err error) bool {
				return errors.Is(err, errors.NewTokenBlacklistError("", nil))
			},
		},
		{
			name:      "上下文中没有token",
			setupMock: func(m *mockBlacklist) {},
			setupCtx: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
			},
			claims:  newTestClaims("test-user", "test-device", "127.0.0.1"),
			wantErr: true,
			checkErr: func(err error) bool {
				return errors.Is(err, errors.NewTokenNotFoundError("", nil))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建mock和验证器
			mockBL := new(mockBlacklist)
			if tt.setupMock != nil {
				tt.setupMock(mockBL)
			}
			v := validator.NewBlacklistValidator(mockBL)

			// 创建测试上下文
			c, _ := gin.CreateTestContext(nil)
			if tt.setupCtx != nil {
				tt.setupCtx(c)
			}

			// 执行验证
			err := v.Validate(c, tt.claims)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err), "错误类型不匹配: %v", err)
				}
			} else {
				assert.NoError(t, err)
			}

			// 验证mock调用
			mockBL.AssertExpectations(t)
		})
	}
}
