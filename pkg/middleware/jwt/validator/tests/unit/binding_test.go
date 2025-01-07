package unit

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/errors"
	"gobase/pkg/middleware/jwt/validator"
)

// TestClaims 用于测试的Claims实现
type TestClaims struct {
	*jwt.StandardClaims
}

// newTestClaims 创建测试用Claims
func newTestClaims(userID, deviceID, ipAddress string) *TestClaims {
	return &TestClaims{
		StandardClaims: jwt.NewStandardClaims(
			jwt.WithUserID(userID),
			jwt.WithDeviceID(deviceID),
			jwt.WithIPAddress(ipAddress),
			jwt.WithTokenType(jwt.AccessToken),
			jwt.WithExpiresAt(time.Now().Add(time.Hour)),
			jwt.WithTokenID("test-token"),
		),
	}
}

func TestBindingValidator_Validate(t *testing.T) {
	tests := []struct {
		name      string
		validator *validator.BindingValidator
		claims    jwt.Claims
		setupCtx  func(*gin.Context)
		wantErr   bool
		checkErr  func(error) bool // 使用 errors.Is 检查错误类型
	}{
		{
			name:      "验证成功",
			validator: validator.NewBindingValidator(),
			claims: newTestClaims(
				"test-user",
				"test-device",
				"127.0.0.1",
			),
			setupCtx: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.RemoteAddr = "127.0.0.1:8080"
			},
			wantErr: false,
		},
		{
			name:      "设备ID为空",
			validator: validator.NewBindingValidator(),
			claims: newTestClaims(
				"test-user",
				"",
				"127.0.0.1",
			),
			setupCtx: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.RemoteAddr = "127.0.0.1:8080"
			},
			wantErr: true,
			checkErr: func(err error) bool {
				return errors.Is(err, errors.NewBindingInvalidError("", nil))
			},
		},
		{
			name:      "IP地址为空",
			validator: validator.NewBindingValidator(),
			claims: newTestClaims(
				"test-user",
				"test-device",
				"",
			),
			setupCtx: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.RemoteAddr = "127.0.0.1:8080"
			},
			wantErr: true,
			checkErr: func(err error) bool {
				return errors.Is(err, errors.NewBindingInvalidError("", nil))
			},
		},
		{
			name:      "IP地址不匹配",
			validator: validator.NewBindingValidator(),
			claims: newTestClaims(
				"test-user",
				"test-device",
				"192.168.1.1",
			),
			setupCtx: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.RemoteAddr = "127.0.0.1:8080"
			},
			wantErr: true,
			checkErr: func(err error) bool {
				return errors.Is(err, errors.NewBindingMismatchError("", nil))
			},
		},
		{
			name: "禁用设备ID验证",
			validator: validator.NewBindingValidator(
				validator.WithDeviceIDValidation(false),
			),
			claims: newTestClaims(
				"test-user",
				"",
				"127.0.0.1",
			),
			setupCtx: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.RemoteAddr = "127.0.0.1:8080"
			},
			wantErr: false,
		},
		{
			name: "禁用IP地址验证",
			validator: validator.NewBindingValidator(
				validator.WithIPAddressValidation(false),
			),
			claims: newTestClaims(
				"test-user",
				"test-device",
				"",
			),
			setupCtx: func(c *gin.Context) {
				c.Request = httptest.NewRequest("GET", "/", nil)
				c.Request.RemoteAddr = "127.0.0.1:8080"
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试上下文
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tt.setupCtx != nil {
				tt.setupCtx(c)
			}

			// 执行验证
			err := tt.validator.Validate(c, tt.claims)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkErr != nil {
					assert.True(t, tt.checkErr(err), "错误类型不匹配: %v", err)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBindingValidatorOptions(t *testing.T) {
	tests := []struct {
		name    string
		options []validator.BindingValidatorOption
		check   func(*validator.BindingValidator) bool
	}{
		{
			name: "禁用设备ID验证",
			options: []validator.BindingValidatorOption{
				validator.WithDeviceIDValidation(false),
			},
			check: func(v *validator.BindingValidator) bool {
				return !v.ValidateDeviceID()
			},
		},
		{
			name: "禁用IP地址验证",
			options: []validator.BindingValidatorOption{
				validator.WithIPAddressValidation(false),
			},
			check: func(v *validator.BindingValidator) bool {
				return !v.ValidateIPAddress()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.NewBindingValidator(tt.options...)
			assert.True(t, tt.check(v))
		})
	}
}
