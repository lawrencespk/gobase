package redis

import (
	"context"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/types"

	"github.com/go-redis/redis/v8"
)

// Pool Redis连接池接口
type Pool interface {
	// Stats 获取连接池统计信息
	Stats() *PoolStats
	// Close 关闭连接池
	Close() error
}

// PoolStats 连接池统计信息
type PoolStats struct {
	// Hits 命中次数
	Hits uint32
	// Misses 未命中次数
	Misses uint32
	// Timeouts 超时次数
	Timeouts uint32
	// TotalConns 总连接数
	TotalConns uint32
	// IdleConns 空闲连接数
	IdleConns uint32
}

// pool Redis连接池实现
type pool struct {
	client interface{} // 使用 interface{} 来支持不同类型的客户端
	logger types.Logger
}

// NewPool 创建连接池
func NewPool(client *redis.Client, logger types.Logger) Pool {
	return &pool{
		client: client,
		logger: logger,
	}
}

// Stats 获取连接池统计信息
func (p *pool) Stats() *PoolStats {
	switch c := p.client.(type) {
	case *redis.Client:
		stats := c.PoolStats()
		return &PoolStats{
			Hits:       uint32(stats.Hits),
			Misses:     uint32(stats.Misses),
			Timeouts:   uint32(stats.Timeouts),
			TotalConns: uint32(stats.TotalConns),
			IdleConns:  uint32(stats.IdleConns),
		}
	case *redis.ClusterClient:
		stats := c.PoolStats()
		return &PoolStats{
			Hits:       uint32(stats.Hits),
			Misses:     uint32(stats.Misses),
			Timeouts:   uint32(stats.Timeouts),
			TotalConns: uint32(stats.TotalConns),
			IdleConns:  uint32(stats.IdleConns),
		}
	default:
		p.logger.Error(context.Background(), "unknown redis client type")
		return &PoolStats{}
	}
}

// PoolStats 获取连接池统计信息
func (c *client) PoolStats() *PoolStats {
	stats := c.client.PoolStats()
	return &PoolStats{
		Hits:       uint32(stats.Hits),
		Misses:     uint32(stats.Misses),
		Timeouts:   uint32(stats.Timeouts),
		TotalConns: uint32(stats.TotalConns),
		IdleConns:  uint32(stats.IdleConns),
	}
}

// Pool 返回连接池实例
func (c *client) Pool() Pool {
	return &pool{
		client: c.client,
		logger: c.logger,
	}
}

// Close 关闭连接池
func (p *pool) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	switch c := p.client.(type) {
	case *redis.Client:
		err = c.Close()
	case *redis.ClusterClient:
		err = c.Close()
	default:
		p.logger.Error(ctx, "unknown redis client type")
		return errors.NewError(codes.CacheError, "unknown redis client type", nil)
	}

	if err != nil {
		p.logger.WithError(err).WithFields(types.Field{
			Key:   "timeout",
			Value: "5s",
		}).Error(ctx, "failed to close redis connection pool")
		return errors.NewError(codes.CacheError, "failed to close redis connection pool", err)
	}

	return nil
}
