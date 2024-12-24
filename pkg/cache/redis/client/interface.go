package client

import (
	"context"
	"time"
)

// Client Redis客户端接口
type Client interface {
	// 基础操作
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error

	// 计数器操作
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)

	// Hash操作
	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key string, values ...interface{}) error

	// List操作
	LPush(ctx context.Context, key string, values ...interface{}) error
	LPop(ctx context.Context, key string) (string, error)

	// Set操作
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SRem(ctx context.Context, key string, members ...interface{}) error

	// ZSet操作
	ZAdd(ctx context.Context, key string, members ...interface{}) error
	ZRem(ctx context.Context, key string, members ...interface{}) error

	// Lua脚本操作
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)

	// 事务操作
	TxPipeline() Pipeline

	// 连接池操作
	Close() error
}

// Pipeline Redis管道接口
type Pipeline interface {
	// 执行管道命令
	Exec(ctx context.Context) error
	// 丢弃管道命令
	Discard() error
}
