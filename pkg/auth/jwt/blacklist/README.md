# JWT Blacklist 包

## 目录
- [概述](#概述)
- [特性](#特性)
- [接口](#接口)
- [实现](#实现)
  - [内存存储](#内存存储)
  - [Redis存储](#redis存储)
- [配置选项](#配置选项)
- [使用示例](#使用示例)
- [测试覆盖](#测试覆盖)
  - [单元测试](#单元测试)
  - [集成测试](#集成测试)
  - [压力测试](#压力测试)
  - [性能测试](#性能测试)

## 概述

JWT Blacklist 包提供了一个用于管理 JWT token 黑名单的解决方案。它支持将已注销或失效的 token 添加到黑名单中,并提供了验证 token 是否在黑名单中的功能。

## 特性

- 支持多种存储实现(内存、Redis)
- 自动过期清理机制
- 可配置的过期时间
- 完整的监控指标
- 分布式支持(通过 Redis 实现)
- 线程安全
- 高性能

## 接口

### TokenBlacklist 接口
```go
type TokenBlacklist interface {
    Add(ctx context.Context, token string, expiration time.Duration) error
    IsBlacklisted(ctx context.Context, token string) (bool, error)
    Remove(ctx context.Context, token string) error
    Clear(ctx context.Context) error
    Close() error
}
```

### Store 接口
```go
type Store interface {
    Add(ctx context.Context, tokenID, reason string, expiration time.Duration) error
    Get(ctx context.Context, tokenID string) (string, error)
    Remove(ctx context.Context, tokenID string) error
    Close() error
}
```

## 实现

### 内存存储

基于 sync.Map 的内存存储实现:

```go
type MemoryStore struct {
    tokens sync.Map
    opts   *Options
    
    // 监控指标
    tokenCount *metric.Gauge
    addTotal   *metric.Counter
    hitTotal   *metric.Counter 
    missTotal  *metric.Counter
}
```

特点:
- 高性能(纳秒级操作)
- 自动内存清理
- 线程安全
- 适合单机部署

### Redis存储

基于 Redis 的分布式存储实现:

```go
type RedisStore struct {
    client redis.Client
    opts   *Options
    
    // 监控指标
    tokenCount *metric.Gauge
    addTotal   *metric.Counter
    hitTotal   *metric.Counter
    missTotal  *metric.Counter
}
```

特点:
- 分布式支持
- 自动过期清理
- 高可用性
- 适合集群部署

## 配置选项

支持以下配置选项:

```go
type Options struct {
    DefaultExpiration time.Duration  // 默认过期时间
    CleanupInterval   time.Duration  // 清理间隔
    Logger           types.Logger    // 日志记录器
    KeyPrefix        string         // Redis key前缀
    EnableMetrics    bool           // 是否启用监控
}
```

配置方法:

```go
opts := blacklist.DefaultOptions()
blacklist.ApplyOptions(opts,
    blacklist.WithLogger(log),
    blacklist.WithDefaultExpiration(time.Hour*48),
    blacklist.WithCleanupInterval(time.Minute*30),
    blacklist.WithMetrics(true),
)
```

## 使用示例

### 内存存储示例

```go
store := blacklist.NewMemoryStore()
defer store.Close()

// 添加token到黑名单
err := store.Add(ctx, "token-123", "revoked", time.Hour*24)

// 检查token是否在黑名单中
reason, err := store.Get(ctx, "token-123")

// 从黑名单中移除token
err = store.Remove(ctx, "token-123")
```

### Redis存储示例

```go
client := redis.NewClient(&redis.Options{...})
store, err := blacklist.NewRedisStore(client, blacklist.DefaultOptions())
defer store.Close()

// 添加token到黑名单
err = store.Add(ctx, "token-123", "revoked", time.Hour*24)

// 检查token是否在黑名单中
reason, err := store.Get(ctx, "token-123")

// 从黑名单中移除token
err = store.Remove(ctx, "token-123")
```

## 测试覆盖

### 单元测试

- [接口测试](tests/unit/interface_test.go)
- [内存存储测试](tests/unit/memory_test.go)
- [Redis存储测试](tests/unit/redis_test.go)
- [配置选项测试](tests/unit/options_test.go)

### 集成测试

- [黑名单集成测试](tests/integration/blacklist_test.go)
  - 验证存储实现的完整功能
  - 测试过期和清理机制
  - 测试并发操作

### 压力测试

- [压力测试](tests/stress/blacklist_stress_test.go)
  - 并发操作测试
  - 大量数据处理测试
  - 稳定性验证

### 性能测试

详细的性能测试结果请参考 [性能测试报告](tests/benchmark/README.md)

主要性能指标:

#### MemoryStore 性能
- Add: 106.7 ns/op, 80 B/op, 3 allocs/op
- Get: 18.69 ns/op, 0 B/op, 0 allocs/op
- Remove: 16.85 ns/op, 0 B/op, 0 allocs/op

#### RedisStore 性能
- Add: 316.723 μs/op, 505 B/op, 17 allocs/op
- Get: 相关性能数据请参考完整报告
- Remove: 相关性能数据请参考完整报告