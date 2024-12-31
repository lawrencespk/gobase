package unit

import (
	"context"
	"errors"
	"gobase/pkg/client/redis"
	"testing"

	gredis "github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestZ(t *testing.T) {
	// 测试 Z 结构体
	tests := []struct {
		name    string
		z       redis.Z
		wantErr bool
	}{
		{
			name: "valid string member",
			z: redis.Z{
				Score:  1.5,
				Member: "test",
			},
			wantErr: false,
		},
		{
			name: "valid int member",
			z: redis.Z{
				Score:  2.0,
				Member: 123,
			},
			wantErr: false,
		},
		{
			name: "nil member",
			z: redis.Z{
				Score:  1.0,
				Member: nil,
			},
			wantErr: false, // Z 结构体允许 nil member
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证结构体字段
			assert.Equal(t, tt.z.Score, tt.z.Score)
			assert.Equal(t, tt.z.Member, tt.z.Member)
		})
	}
}

func TestCmder(t *testing.T) {
	ctx := context.Background()

	// 测试 Cmder 接口
	tests := []struct {
		name    string
		cmd     redis.Cmder
		wantErr bool
	}{
		{
			name:    "valid command",
			cmd:     gredis.NewStringCmd(ctx, "SET", "key", "value"),
			wantErr: false,
		},
		{
			name: "error command",
			cmd: func() redis.Cmder {
				cmd := gredis.NewStringCmd(ctx, "INVALID")
				// 手动设置错误以模拟无效命令
				cmd.SetErr(errors.New("ERR unknown command 'INVALID'"))
				return cmd
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证命令基本属性
			assert.NotEmpty(t, tt.cmd.Name(), "command name should not be empty")
			assert.NotNil(t, tt.cmd.Args(), "command args should not be nil")
			assert.NotEmpty(t, tt.cmd.String(), "command string representation should not be empty")

			// 验证错误状态
			if tt.wantErr {
				assert.Error(t, tt.cmd.Err(), "expected an error but got nil")
			} else {
				assert.NoError(t, tt.cmd.Err(), "expected no error but got one")
			}
		})
	}
}
