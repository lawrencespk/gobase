package mock

import (
	"context"
	"time"

	"gobase/pkg/client/redis"
)

// MockPipeline 实现 Pipeline 接口
type MockPipeline struct {
	err error
}

// 基础操作
func (p *MockPipeline) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if p.err != nil {
		return p.err
	}
	return nil
}

func (p *MockPipeline) Get(ctx context.Context, key string) (string, error) {
	if p.err != nil {
		return "", p.err
	}
	return "", nil
}

func (p *MockPipeline) Del(ctx context.Context, keys ...string) (int64, error) {
	if p.err != nil {
		return 0, p.err
	}
	return 1, nil
}

// Hash操作
func (p *MockPipeline) HSet(ctx context.Context, key string, values ...interface{}) (int64, error) {
	if p.err != nil {
		return 0, p.err
	}
	return 1, nil
}

func (p *MockPipeline) HGet(ctx context.Context, key, field string) (string, error) {
	if p.err != nil {
		return "", p.err
	}
	return "", nil
}

func (p *MockPipeline) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	if p.err != nil {
		return 0, p.err
	}
	return 1, nil
}

// Set操作
func (p *MockPipeline) SAdd(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if p.err != nil {
		return 0, p.err
	}
	return 1, nil
}

func (p *MockPipeline) SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if p.err != nil {
		return 0, p.err
	}
	return 1, nil
}

// ZSet操作
func (p *MockPipeline) ZAdd(ctx context.Context, key string, members ...*redis.Z) (int64, error) {
	if p.err != nil {
		return 0, p.err
	}
	return 1, nil
}

func (p *MockPipeline) ZRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	if p.err != nil {
		return 0, p.err
	}
	return 1, nil
}

// 管道控制
func (p *MockPipeline) Exec(ctx context.Context) ([]redis.Cmder, error) {
	if p.err != nil {
		return nil, p.err
	}
	return []redis.Cmder{}, nil
}

func (p *MockPipeline) Close() error {
	return p.err
}

// 过期时间操作
func (p *MockPipeline) ExpireAt(ctx context.Context, key string, tm time.Time) (bool, error) {
	if p.err != nil {
		return false, p.err
	}
	return true, nil
}
