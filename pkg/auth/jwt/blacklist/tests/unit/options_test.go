package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gobase/pkg/auth/jwt/blacklist"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
)

func TestOptions(t *testing.T) {
	t.Run("DefaultOptions", func(t *testing.T) {
		opts := blacklist.DefaultOptions()
		assert.NotNil(t, opts)
		assert.Equal(t, time.Hour*24, opts.DefaultExpiration)
		assert.Equal(t, time.Hour, opts.CleanupInterval)
	})

	t.Run("WithLogger", func(t *testing.T) {
		log, err := logger.NewLogger(logger.WithLevel(types.InfoLevel))
		assert.NoError(t, err)

		opts := blacklist.DefaultOptions()
		blacklist.WithLogger(log)(opts)
		assert.Equal(t, log, opts.Logger)
	})

	t.Run("WithDefaultExpiration", func(t *testing.T) {
		expiration := time.Hour * 48
		opts := blacklist.DefaultOptions()
		blacklist.WithDefaultExpiration(expiration)(opts)
		assert.Equal(t, expiration, opts.DefaultExpiration)
	})

	t.Run("WithCleanupInterval", func(t *testing.T) {
		interval := time.Minute * 30
		opts := blacklist.DefaultOptions()
		blacklist.WithCleanupInterval(interval)(opts)
		assert.Equal(t, interval, opts.CleanupInterval)
	})

	t.Run("WithMetrics", func(t *testing.T) {
		opts := blacklist.DefaultOptions()
		blacklist.WithMetrics(true)(opts)
		assert.True(t, opts.EnableMetrics)

		blacklist.WithMetrics(false)(opts)
		assert.False(t, opts.EnableMetrics)
	})

	t.Run("OptionValidation", func(t *testing.T) {
		tests := []struct {
			name    string
			option  blacklist.Option
			wantErr bool
		}{
			{
				name:    "有效的过期时间",
				option:  blacklist.WithDefaultExpiration(time.Hour),
				wantErr: false,
			},
			{
				name:    "无效的过期时间",
				option:  blacklist.WithDefaultExpiration(-time.Hour),
				wantErr: true,
			},
			{
				name:    "有效的清理间隔",
				option:  blacklist.WithCleanupInterval(time.Minute),
				wantErr: false,
			},
			{
				name:    "无效的清理间隔",
				option:  blacklist.WithCleanupInterval(-time.Minute),
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				opts := blacklist.DefaultOptions()
				err := tt.option(opts)

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("MultipleOptions", func(t *testing.T) {
		log, err := logger.NewLogger(logger.WithLevel(types.InfoLevel))
		assert.NoError(t, err)

		opts := blacklist.DefaultOptions()
		err = blacklist.ApplyOptions(opts,
			blacklist.WithLogger(log),
			blacklist.WithDefaultExpiration(time.Hour*48),
			blacklist.WithCleanupInterval(time.Minute*30),
			blacklist.WithMetrics(true),
		)

		assert.NoError(t, err)
		assert.Equal(t, log, opts.Logger)
		assert.Equal(t, time.Hour*48, opts.DefaultExpiration)
		assert.Equal(t, time.Minute*30, opts.CleanupInterval)
		assert.True(t, opts.EnableMetrics)
	})
}
