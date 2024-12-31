package redis

import (
	"github.com/go-redis/redis/v8"
)

// Z 是有序集合的成员
type Z struct {
	Score  float64
	Member interface{}
}

// Cmder 是 Redis 命令接口
type Cmder interface {
	// Name 返回命令名称
	Name() string
	// Args 返回命令参数
	Args() []interface{}
	// String 返回命令字符串表示
	String() string
	// Err 返回命令执行错误
	Err() error
}

// 确保 redis.Cmder 实现了 Cmder 接口
var _ Cmder = redis.Cmder(nil)
