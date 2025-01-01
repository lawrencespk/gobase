package sync

import (
	"context"
	"time"

	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

// Locker 分布式锁接口
type Locker interface {
	// Lock 获取锁
	Lock(ctx context.Context) error

	// TryLock 尝试获取锁
	TryLock(ctx context.Context) (bool, error)

	// Unlock 释放锁
	Unlock(ctx context.Context) error
}

// NewLocker 创建分布式锁
func NewLocker(client redis.Cache, key string) Locker {
	return &redisLocker{
		client: client,
		key:    "jwt:lock:" + key,
	}
}

type redisLocker struct {
	client redis.Cache
	key    string
}

// Lock 获取锁
func (l *redisLocker) Lock(ctx context.Context) error {
	success, err := l.client.SetNX(ctx, l.key, "1", 30*time.Second)
	if err != nil {
		return errors.NewError(codes.RedisLockError, "failed to acquire lock", err)
	}
	if !success {
		return errors.NewError(codes.RedisLockError, "lock is held by another process", nil)
	}
	return nil
}

// TryLock 尝试获取锁
func (l *redisLocker) TryLock(ctx context.Context) (bool, error) {
	success, err := l.client.SetNX(ctx, l.key, "1", 30*time.Second)
	if err != nil {
		return false, errors.NewError(codes.RedisLockError, "failed to try lock", err)
	}
	return success, nil
}

// Unlock 释放锁
func (l *redisLocker) Unlock(ctx context.Context) error {
	err := l.client.Del(ctx, l.key)
	if err != nil {
		return errors.NewError(codes.RedisLockError, "failed to release lock", err)
	}
	return nil
}
