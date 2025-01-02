package unit

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/cache/redis"

	"github.com/stretchr/testify/assert"
)

func TestSerialization(t *testing.T) {
	ctx := context.Background()
	mockClient := newMockRedisClient()
	cache, err := redis.NewCache(redis.Options{
		Client: mockClient,
	})
	assert.NoError(t, err)

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "nil value",
			value:   nil,
			wantErr: true,
		},
		{
			name: "complex struct",
			value: struct {
				Int    int
				String string
				Map    map[string]interface{}
			}{
				Int:    42,
				String: "test",
				Map: map[string]interface{}{
					"key": "value",
				},
			},
			wantErr: false,
		},
		{
			name:    "circular reference",
			value:   createCircularRef(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cache.Set(ctx, "test_key", tt.value, time.Minute)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.value != nil {
					assert.Contains(t, err.Error(), "encountered a cycle")
				}
			} else {
				assert.NoError(t, err)

				value, err := cache.Get(ctx, "test_key")
				assert.NoError(t, err)
				assert.NotNil(t, value)

				if tt.name == "complex struct" {
					data, ok := value.(map[string]interface{})
					assert.True(t, ok)
					assert.Equal(t, float64(42), data["Int"])
					assert.Equal(t, "test", data["String"])
					assert.Equal(t, "value", data["Map"].(map[string]interface{})["key"])
				}
			}
		})
	}
}

func createCircularRef() interface{} {
	type Node struct {
		Next *Node
	}
	node := &Node{}
	node.Next = node
	return node
}
