package ratelimit_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gobase/pkg/middleware/ratelimit"
	mocklimiter "gobase/pkg/middleware/ratelimit/tests/mock"
)

func TestRateLimit_Config(t *testing.T) {
	tests := []struct {
		name        string
		config      *ratelimit.Config
		shouldPanic bool
	}{
		{
			name: "should panic when limiter is nil",
			config: &ratelimit.Config{
				Limit:  100,
				Window: time.Minute,
			},
			shouldPanic: true,
		},
		{
			name: "should panic when limit is zero",
			config: &ratelimit.Config{
				Limiter: new(mocklimiter.MockLimiter),
				Window:  time.Minute,
			},
			shouldPanic: true,
		},
		{
			name: "should panic when window is zero",
			config: &ratelimit.Config{
				Limiter: new(mocklimiter.MockLimiter),
				Limit:   100,
			},
			shouldPanic: true,
		},
		{
			name: "should not panic with valid config",
			config: &ratelimit.Config{
				Limiter: new(mocklimiter.MockLimiter),
				Limit:   100,
				Window:  time.Minute,
			},
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.Panics(t, func() {
					ratelimit.RateLimit(tt.config)
				})
			} else {
				assert.NotPanics(t, func() {
					ratelimit.RateLimit(tt.config)
				})
			}
		})
	}
}
