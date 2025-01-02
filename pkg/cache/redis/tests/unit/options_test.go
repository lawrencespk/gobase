package unit

import (
	"testing"

	"gobase/pkg/cache/redis"

	"github.com/stretchr/testify/assert"
)

func TestNewCacheOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    redis.Options
		wantErr bool
	}{
		{
			name: "nil client",
			opts: redis.Options{
				Client: nil,
			},
			wantErr: true,
		},
		{
			name: "with logger",
			opts: redis.Options{
				Client: &mockRedisClient{},
				Logger: &mockLogger{},
			},
			wantErr: false,
		},
		{
			name: "minimal valid options",
			opts: redis.Options{
				Client: &mockRedisClient{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache, err := redis.NewCache(tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cache)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cache)
			}
		})
	}
}
