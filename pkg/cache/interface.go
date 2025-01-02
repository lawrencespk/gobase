package cache

import (
	"context"
	"time"
)

// Level 缓存级别
type Level int

const (
	// L1Cache 一级缓存(本地内存)
	L1Cache Level = iota + 1
	// L2Cache 二级缓存(Redis)
	L2Cache
)

// Cache 缓存接口
type Cache interface {
	// Get 获取缓存
	Get(ctx context.Context, key string) (interface{}, error)

	// Set 设置缓存
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Delete 删除缓存
	Delete(ctx context.Context, key string) error

	// Clear 清空缓存
	Clear(ctx context.Context) error

	// GetLevel 获取缓存级别
	GetLevel() Level
}

// MultiLevelCache 多级缓存接口
type MultiLevelCache interface {
	Cache

	// GetFromLevel 从指定级别获取缓存
	GetFromLevel(ctx context.Context, key string, level Level) (interface{}, error)

	// SetToLevel 设置缓存到指定级别
	SetToLevel(ctx context.Context, key string, value interface{}, expiration time.Duration, level Level) error

	// DeleteFromLevel 从指定级别删除缓存
	DeleteFromLevel(ctx context.Context, key string, level Level) error

	// ClearLevel 清空指定级别的缓存
	ClearLevel(ctx context.Context, level Level) error

	// Warmup 缓存预热
	Warmup(ctx context.Context, keys []string) error
}
