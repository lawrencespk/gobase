package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gobase/pkg/cache/memory"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *memory.Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &memory.Config{
				MaxEntries:      1000,
				CleanupInterval: time.Second,
				DefaultTTL:      time.Hour,
			},
			wantErr: false,
		},
		{
			name: "zero max entries",
			config: &memory.Config{
				MaxEntries:      0,
				CleanupInterval: time.Second,
				DefaultTTL:      time.Hour,
			},
			wantErr: true,
		},
		{
			name: "negative max entries",
			config: &memory.Config{
				MaxEntries:      -1,
				CleanupInterval: time.Second,
				DefaultTTL:      time.Hour,
			},
			wantErr: true,
		},
		{
			name: "zero cleanup interval",
			config: &memory.Config{
				MaxEntries:      1000,
				CleanupInterval: 0,
				DefaultTTL:      time.Hour,
			},
			wantErr: true,
		},
		{
			name: "negative cleanup interval",
			config: &memory.Config{
				MaxEntries:      1000,
				CleanupInterval: -time.Second,
				DefaultTTL:      time.Hour,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
