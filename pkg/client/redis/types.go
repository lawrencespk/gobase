package redis

import (
	goredis "github.com/go-redis/redis/v8"
)

// Cmder 是 Redis 命令的接口定义
type Cmder interface {
	// Name 返回命令名称
	Name() string
	// Args 返回命令参数
	Args() []interface{}
	// String 返回命令的字符串表示
	String() string
	// Err 返回命令执行的错误
	Err() error
}

// 将 go-redis 的 Cmder 转换为我们的 Cmder
type cmder struct {
	cmd goredis.Cmder
}

func (c *cmder) Name() string {
	return c.cmd.Name()
}

func (c *cmder) Args() []interface{} {
	return c.cmd.Args()
}

func (c *cmder) String() string {
	return c.cmd.String()
}

func (c *cmder) Err() error {
	return c.cmd.Err()
}

// wrapCmder 将 go-redis 的 Cmder 包装为我们的 Cmder
func wrapCmder(cmd goredis.Cmder) Cmder {
	return &cmder{cmd: cmd}
}

// Z 是有序集合的成员
type Z struct {
	Score  float64
	Member interface{}
}
