# Redis 包

提供高性能、可扩展的Redis缓存和分布式限流实现。

## 目录
1. [功能特性](#1-功能特性)
2. [快速开始](#2-快速开始) 
3. [配置说明](#3-配置说明)
4. [监控指标](#4-监控指标)
5. [错误处理](#5-错误处理)
6. [性能测试](#6-性能测试)
7. [最佳实践](#7-最佳实践)

## 1. 功能特性

### 1.1 缓存功能
- 支持所有Redis基础数据类型操作
  - String操作
  - Hash操作 
  - List操作
  - Set操作
  - ZSet操作
- 支持Pipeline批量操作
- 支持Cluster集群模式
- 内置连接池管理
- 支持TLS加密连接

### 1.2 限流功能
- 支持多种限流算法
  - 滑动窗口 (Sliding Window)
  - 令牌桶 (Token Bucket) 
  - 漏桶 (Leaky Bucket)
  - 固定窗口 (Fixed Window)
- 分布式限流支持
- 支持等待模式
- 支持自定义限流规则

### 1.3 通用特性
- 完整的监控指标
- 支持链路追踪
- 优雅的错误处理
- 高性能实现

## 2. 快速开始

### 2.1 缓存基础用法

```go
import "gobase/pkg/client/redis"

// 创建Redis客户端
client := redis.NewClient(
    redis.WithAddresses("localhost:6379"),
    redis.WithPassword(""),
    redis.WithDatabase(0),
)

// String操作
err := client.Set(ctx, "key", "value", 0)
val, err := client.Get(ctx, "key")

// Hash操作
err := client.HSet(ctx, "hash", "field", "value")
val, err := client.HGet(ctx, "hash", "field")
```

### 2.2 限流基础用法

```go
import (
    "gobase/pkg/ratelimit/redis"
    "gobase/pkg/ratelimit/core"
)

// 创建限流器
limiter := redis.NewSlidingWindowLimiter(redisClient, 
    core.WithName("api_limiter"),
    core.WithMetrics(true),
)

// 检查是否允许请求
allowed, err := limiter.Allow(ctx, "user:123", 100, time.Minute)
if err != nil {
    return err
}
```

## 3. 配置说明

### 3.1 Redis基础配置

```go
config := redis.Config{
    Addresses:     []string{"localhost:6379"},
    Username:      "",
    Password:      "",
    Database:      0,
    PoolSize:      10,
    MaxRetries:    3,
    DialTimeout:   5 * time.Second,
    ReadTimeout:   3 * time.Second,
    WriteTimeout:  3 * time.Second,
}
```

### 3.2 限流器配置

```go
limiter := redis.NewSlidingWindowLimiter(redisClient,
    core.WithName("api_limiter"),
    core.WithAlgorithm("sliding_window"),
    core.WithMetrics(true),
    core.WithMetricsPrefix("myapp"),
    core.WithTracing(true),
)
```

## 4. 监控指标

### 4.1 缓存指标
```go
// 操作延迟
CacheLatency.WithLabelValues("get").Observe(duration)

// 错误计数
CacheErrors.WithLabelValues("set").Inc()

// 连接池指标
PoolStats.WithLabelValues("idle").Set(float64(stats.IdleConns))
```

### 4.2 限流指标
```go
// 请求计数器
RequestsTotal.WithLabelValues(key, "allowed").Inc()

// 被拒绝计数器
RejectedTotal.WithLabelValues(key).Inc()

// 延迟直方图
LimiterLatency.WithLabelValues(key, "allow").Observe(duration)
```

## 5. 错误处理

```go
// 缓存错误处理
if err := client.Set(ctx, key, value, ttl); err != nil {
    switch {
    case redis.IsConnError(err):
        // 处理连接错误
    case redis.IsTimeoutError(err):
        // 处理超时错误
    default:
        // 处理其他错误
    }
}

// 限流错误处理
if err := limiter.Wait(ctx, key, limit, window); err != nil {
    switch {
    case errors.Is(err, context.DeadlineExceeded):
        // 处理超时错误
    case errors.Is(err, context.Canceled):
        // 处理取消错误
    default:
        // 处理其他错误
    }
}
```

## 6. 性能测试

### 6.1 缓存性能
关于缓存的详细性能测试结果，请参考 [缓存性能测试报告](tests/benchmark/README.md)

### 6.2 限流性能
关于限流器的详细性能测试结果，请参考 [限流性能测试报告](ratelimit/tests/benchmark/README.md)

## 7. 最佳实践

### 7.1 缓存使用建议
- 合理设置TTL避免内存泄漏
- 使用Pipeline批量操作提升性能
- 合理配置连接池大小
- 实现优雅降级机制

### 7.2 限流使用建议
- 根据场景选择合适的限流算法
  - 一般场景：滑动窗口
  - 允许突发：令牌桶
  - 严格控制：漏桶
  - 简单需求：固定窗口
- 设置合理的限流阈值
- 监控限流指标
- 配置告警阈值

### 7.3 Redis配置建议
- 使用Redis集群保证高可用
- 合理配置连接池大小
- 设置适当的超时时间
- 启用重试机制

### 7.4 监控建议
- 监控关键性能指标
- 设置合理的告警阈值
- 记录详细的操作日志
- 定期检查错误日志