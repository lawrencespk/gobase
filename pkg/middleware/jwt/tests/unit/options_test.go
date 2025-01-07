package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jwt "gobase/pkg/middleware/jwt"
	"gobase/pkg/middleware/jwt/extractor"
	"gobase/pkg/middleware/jwt/validator"
)

func TestNewOptions(t *testing.T) {
	// 测试默认选项
	t.Run("默认选项", func(t *testing.T) {
		opts := jwt.NewOptions()
		assert.False(t, opts.EnableTracing)
		assert.False(t, opts.EnableMetrics)
		assert.Equal(t, 60, opts.Timeout)
		assert.NotNil(t, opts.Extractor)
		assert.NotNil(t, opts.Validator)
	})

	// 测试自定义选项
	t.Run("自定义选项", func(t *testing.T) {
		customExtractor := extractor.NewHeaderExtractor("Custom-Auth", "Token ")
		customValidator := validator.NewClaimsValidator()

		opts := jwt.NewOptions(
			jwt.WithTracing(true),
			jwt.WithMetrics(true),
			jwt.WithTimeout(30),
			jwt.WithExtractor(customExtractor),
			jwt.WithValidator(customValidator),
		)

		assert.True(t, opts.EnableTracing)
		assert.True(t, opts.EnableMetrics)
		assert.Equal(t, 30, opts.Timeout)
		assert.Equal(t, customExtractor, opts.Extractor)
		assert.Equal(t, customValidator, opts.Validator)
	})
}

func TestOptionFunctions(t *testing.T) {
	tests := []struct {
		name   string
		option jwt.Option
		verify func(*testing.T, *jwt.Options)
	}{
		{
			name:   "WithTracing",
			option: jwt.WithTracing(true),
			verify: func(t *testing.T, opts *jwt.Options) {
				assert.True(t, opts.EnableTracing)
			},
		},
		{
			name:   "WithMetrics",
			option: jwt.WithMetrics(true),
			verify: func(t *testing.T, opts *jwt.Options) {
				assert.True(t, opts.EnableMetrics)
			},
		},
		{
			name:   "WithTimeout",
			option: jwt.WithTimeout(120),
			verify: func(t *testing.T, opts *jwt.Options) {
				assert.Equal(t, 120, opts.Timeout)
			},
		},
		{
			name:   "WithExtractor",
			option: jwt.WithExtractor(extractor.NewHeaderExtractor("X-Token", "Bearer ")),
			verify: func(t *testing.T, opts *jwt.Options) {
				require.NotNil(t, opts.Extractor)
				_, ok := opts.Extractor.(*extractor.HeaderExtractor)
				assert.True(t, ok)
			},
		},
		{
			name:   "WithValidator",
			option: jwt.WithValidator(validator.NewClaimsValidator()),
			verify: func(t *testing.T, opts *jwt.Options) {
				require.NotNil(t, opts.Validator)
				_, ok := opts.Validator.(*validator.ClaimsValidator)
				assert.True(t, ok)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &jwt.Options{}
			tt.option(opts)
			tt.verify(t, opts)
		})
	}
}
