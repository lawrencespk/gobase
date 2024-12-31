package ratelimit

import (
	"context"

	"gobase/pkg/client/redis"
)

// Store Redis限流存储适配器
type Store struct {
	client redis.Client
}

// NewStore 创建Redis限流存储
func NewStore(client redis.Client) *Store {
	return &Store{client: client}
}

// Eval 执行Lua脚本
func (s *Store) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return s.client.Eval(ctx, script, keys, args...)
}

// Del 删除一个或多个键
func (s *Store) Del(ctx context.Context, keys ...string) error {
	_, err := s.client.Del(ctx, keys...)
	return err
}

// 实现其他必要的存储方法...
