package unit

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/middleware/jwt/validator"
)

// mockValidator 模拟验证器实现
type mockValidator struct {
	mock.Mock
}

func (m *mockValidator) Validate(c *gin.Context, claims jwt.Claims) error {
	args := m.Called(c, claims)
	return args.Error(0)
}

func TestValidatorFunc(t *testing.T) {
	t.Run("ValidatorFunc实现", func(t *testing.T) {
		called := false
		f := validator.ValidatorFunc(func(c *gin.Context, claims jwt.Claims) error {
			called = true
			return nil
		})

		err := f.Validate(nil, nil)
		assert.NoError(t, err)
		assert.True(t, called)
	})
}

func TestChainValidator_Validate(t *testing.T) {
	tests := []struct {
		name       string
		validators []validator.TokenValidator
		setupMocks func([]*mockValidator)
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "空验证器链",
			validators: nil,
			wantErr:    false,
		},
		{
			name: "单个验证器成功",
			validators: []validator.TokenValidator{
				&mockValidator{},
			},
			setupMocks: func(mocks []*mockValidator) {
				mocks[0].On("Validate", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "多个验证器全部成功",
			validators: []validator.TokenValidator{
				&mockValidator{},
				&mockValidator{},
			},
			setupMocks: func(mocks []*mockValidator) {
				mocks[0].On("Validate", mock.Anything, mock.Anything).Return(nil)
				mocks[1].On("Validate", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "验证器链中途失败",
			validators: []validator.TokenValidator{
				&mockValidator{},
				&mockValidator{},
			},
			setupMocks: func(mocks []*mockValidator) {
				mocks[0].On("Validate", mock.Anything, mock.Anything).Return(nil)
				mocks[1].On("Validate", mock.Anything, mock.Anything).Return(assert.AnError)
			},
			wantErr: true,
			errMsg:  assert.AnError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建mock验证器
			mocks := make([]*mockValidator, len(tt.validators))
			validators := make([]validator.TokenValidator, len(tt.validators))
			for i := range tt.validators {
				mocks[i] = new(mockValidator)
				validators[i] = mocks[i]
			}

			// 设置mock期望
			if tt.setupMocks != nil {
				tt.setupMocks(mocks)
			}

			// 创建链式验证器
			chain := validator.ChainValidator(validators)

			// 执行验证
			err := chain.Validate(nil, nil)

			// 验证结果
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// 验证所有mock调用
			for _, m := range mocks {
				m.AssertExpectations(t)
			}
		})
	}
}
