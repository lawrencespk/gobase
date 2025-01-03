package redis

import (
	"context"
	"time"
)

// Client Redis客户端接口定义
type Client interface {
	// 基础操作
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) (int64, error)

	// Hash操作
	HGet(ctx context.Context, key, field string) (string, error)
	HSet(ctx context.Context, key string, values ...interface{}) (int64, error)
	HDel(ctx context.Context, key string, fields ...string) (int64, error)

	// List操作
	LPush(ctx context.Context, key string, values ...interface{}) (int64, error)
	LPop(ctx context.Context, key string) (string, error)

	// Set操作
	SAdd(ctx context.Context, key string, members ...interface{}) (int64, error)
	SRem(ctx context.Context, key string, members ...interface{}) (int64, error)

	// ZSet操作
	ZAdd(ctx context.Context, key string, members ...*Z) (int64, error)
	ZRem(ctx context.Context, key string, members ...interface{}) (int64, error)

	// 事务操作
	TxPipeline() Pipeline

	// Lua脚本
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)

	// 连接管理
	Close() error
	Ping(ctx context.Context) error

	// 监控相关
	PoolStats() *PoolStats

	// 缓存管理
	Exists(ctx context.Context, key string) (bool, error)

	// 连接池管理
	Pool() Pool

	// Publish 发布消息到指定的频道
	Publish(ctx context.Context, channel string, message interface{}) error

	// Publish/Subscribe 操作
	Subscribe(ctx context.Context, channels ...string) PubSub
}

// Cache Redis缓存接口
type Cache interface {
	// 基础操作
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Del(ctx context.Context, key string) error

	// 批量操作
	MGet(ctx context.Context, keys ...string) ([]interface{}, error)
	MSet(ctx context.Context, pairs ...interface{}) error

	// 原子操作
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Incr(ctx context.Context, key string) (int64, error)

	// 缓存管理
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) (bool, error)

	// 关闭
	Close() error
}

// Pipeline 管道接口
type Pipeline interface {
	// 基础操作
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) (int64, error)

	// Hash操作
	HSet(ctx context.Context, key string, values ...interface{}) (int64, error)
	HGet(ctx context.Context, key, field string) (string, error)
	HDel(ctx context.Context, key string, fields ...string) (int64, error)

	// Set操作
	SAdd(ctx context.Context, key string, members ...interface{}) (int64, error)
	SRem(ctx context.Context, key string, members ...interface{}) (int64, error)

	// ZSet操作
	ZAdd(ctx context.Context, key string, members ...*Z) (int64, error)
	ZRem(ctx context.Context, key string, members ...interface{}) (int64, error)

	// 管道控制
	Exec(ctx context.Context) ([]Cmder, error)
	Close() error

	// 过期时间操作
	ExpireAt(ctx context.Context, key string, tm time.Time) (bool, error)
}

// PubSub 发布订阅接口
type PubSub interface {
	// 接收消息
	ReceiveMessage(ctx context.Context) (*Message, error)
	// 关闭订阅
	Close() error
}

// Message 消息结构
type Message struct {
	Channel string
	Pattern string
	Payload string
}
