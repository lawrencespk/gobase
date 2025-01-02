package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"gobase/pkg/cache/redis/ratelimit"
	mockRedis "gobase/pkg/cache/redis/ratelimit/tests/mock"
)

func TestStore_Eval(t *testing.T) {
	tests := []struct {
		name    string
		script  string
		keys    []string
		args    []interface{}
		want    interface{}
		wantErr bool
		mock    func(*mockRedis.MockRedisClient)
	}{
		{
			name:   "success",
			script: "return KEYS[1]",
			keys:   []string{"test-key"},
			args:   []interface{}{1},
			want:   "test-key",
			mock: func(m *mockRedis.MockRedisClient) {
				m.On("Eval",
					mock.Anything,
					"return KEYS[1]",
					[]string{"test-key"},
					[]interface{}{1},
				).Return("test-key", nil)
			},
		},
		{
			name:    "error",
			script:  "invalid script",
			keys:    []string{"test-key"},
			args:    []interface{}{1},
			wantErr: true,
			mock: func(m *mockRedis.MockRedisClient) {
				m.On("Eval",
					mock.Anything,
					"invalid script",
					[]string{"test-key"},
					[]interface{}{1},
				).Return(nil, assert.AnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mockRedis.MockRedisClient)
			if tt.mock != nil {
				tt.mock(mockClient)
			}

			s := ratelimit.NewStore(mockClient)
			got, err := s.Eval(context.Background(), tt.script, tt.keys, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
			mockClient.AssertExpectations(t)
		})
	}
}

func TestStore_Del(t *testing.T) {
	tests := []struct {
		name    string
		keys    []string
		wantErr bool
		mock    func(*mockRedis.MockRedisClient)
	}{
		{
			name: "success single key",
			keys: []string{"key1"},
			mock: func(m *mockRedis.MockRedisClient) {
				m.On("Del",
					mock.Anything,
					[]string{"key1"},
				).Return(int64(1), nil)
			},
		},
		{
			name: "success multiple keys",
			keys: []string{"key1", "key2"},
			mock: func(m *mockRedis.MockRedisClient) {
				m.On("Del",
					mock.Anything,
					[]string{"key1", "key2"},
				).Return(int64(2), nil)
			},
		},
		{
			name:    "error",
			keys:    []string{"key1"},
			wantErr: true,
			mock: func(m *mockRedis.MockRedisClient) {
				m.On("Del",
					mock.Anything,
					[]string{"key1"},
				).Return(int64(0), assert.AnError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := new(mockRedis.MockRedisClient)
			if tt.mock != nil {
				tt.mock(mockClient)
			}

			s := ratelimit.NewStore(mockClient)
			err := s.Del(context.Background(), tt.keys...)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			mockClient.AssertExpectations(t)
		})
	}
}
