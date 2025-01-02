# RateLimit 限流包

提供高性能、可扩展的分布式限流实现。

## 1. 功能特性

- 支持多种限流算法
  - 滑动窗口 (Sliding Window)
  - 令牌桶 (Token Bucket)
  - 漏桶 (Leaky Bucket)
  - 固定窗口 (Fixed Window)
- 分布式限流支持 (基于 Redis)
- 完整的监控指标
- 支持链路追踪
- 优雅的错误处理
- 高性能实现

## 2. 快速开始

### 2.1 基础用法

```go
import (
    "gobase/pkg/ratelimit/redis"
    "gobase/pkg/ratelimit/core"
)

// 创建Redis限流器
limiter := redis.NewSlidingWindowLimiter(redisClient, 
    core.WithName("api_limiter"),
    core.WithMetrics(true),
)

// 检查是否允许请求
allowed, err := limiter.Allow(ctx, "user:123", 100, time.Minute)
if err != nil {
    return err
}
if !allowed {
    return errors.New("rate limit exceeded")
}
```

### 2.2 等待模式

```go
// 等待直到允许请求或超时
err := limiter.Wait(ctx, "user:123", 100, time.Minute)
if err != nil {
    return err
}
```

## 3. 限流器配置

### 3.1 Redis配置

```go
redisConfig := &core.RedisConfig{
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

### 3.2 限流器选项

```go
limiter := redis.NewSlidingWindowLimiter(redisClient,
    core.WithName("api_limiter"),
    core.WithAlgorithm("sliding_window"),
    core.WithRedisConfig(redisConfig),
    core.WithMetrics(true),
    core.WithMetricsPrefix("myapp"),
    core.WithTracing(true),
)
```

## 4. 监控指标

### 4.1 基础指标

```go
// 请求计数器
RequestsTotal.WithLabelValues(key, "allowed").Inc()

// 被拒绝计数器
RejectedTotal.WithLabelValues(key).Inc()

// 延迟直方图
LimiterLatency.WithLabelValues(key, "allow").Observe(duration)
```

### 4.2 Prometheus集成

```go
// 注册限流器收集器
collector := metrics.NewRateLimitCollector()
prometheus.MustRegister(collector)
```

## 5. 错误处理

```go
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

最新的基准测试结果显示([查看完整测试报告](../cache/redis/ratelimit/tests/benchmark/README.md)):

| 操作类型 | 操作延迟 (ns/op) | 内存分配 (B/op) | 分配次数 (allocs/op) | 并发度 |
|---------|----------------|----------------|-------------------|--------|
| Eval    | 618,258       | 486           | 13               | 32     |
| Del     | 733,306       | 413           | 13               | 32     |
| Parallel| 59,888        | 542           | 17               | 32     |

### 性能特点

- 并发性能优秀，Parallel 测试比单线程快约 10 倍
- 内存分配稳定，各测试间差异不大
- 延迟在可接受范围内
- 基本操作的内存分配保持在 400-550 字节范围内
- 分配次数保持稳定，在 13-17 次之间

## 7. 性能优化

### 7.1 Redis优化

```go
redisConfig := &core.RedisConfig{
    PoolSize:     100,    // 增加连接池大小
    MaxRetries:   3,      // 设置重试次数
    DialTimeout:  100 * time.Millisecond,
    ReadTimeout:  100 * time.Millisecond,
    WriteTimeout: 100 * time.Millisecond,
}
```

### 7.2 本地缓存

建议在高并发场景下使用本地缓存来减少 Redis 访问：

```go
import "gobase/pkg/cache/memory"

cache := memory.NewCache(
    memory.WithExpiration(time.Second),
    memory.WithCleanupInterval(time.Minute),
)
```

## 8. 测试

### 8.1 单元测试

```go
func TestSlidingWindowLimiter_Allow(t *testing.T) {
    mockClient := new(MockRedisClient)
    limiter := redis.NewSlidingWindowLimiter(mockClient)
    
    allowed, err := limiter.Allow(context.Background(), "test_key", 10, time.Second)
    assert.NoError(t, err)
    assert.True(t, allowed)
}
```

### 8.2 集成测试

```go
func TestSlidingWindowLimiter_Integration(t *testing.T) {
    redisClient := setupRedisClient(t)
    limiter := redis.NewSlidingWindowLimiter(redisClient)
    
    const (
        key     = "test_integration"
        limit   = int64(10)
        window  = time.Second
    )
    
    // 并发测试
    var wg sync.WaitGroup
    for i := 0; i < 20; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            allowed, err := limiter.Allow(context.Background(), key, limit, window)
            require.NoError(t, err)
            t.Logf("Request allowed: %v", allowed)
        }()
    }
    wg.Wait()
}
```

## 9. 最佳实践

1. 选择合适的限流算法
   - 一般场景：滑动窗口
   - 允许突发：令牌桶
   - 严格控制：漏桶
   - 简单需求：固定窗口

2. Redis配置建议
   - 使用Redis集群保证高可用
   - 合理配置连接池大小
   - 设置适当的超时时间
   - 启用重试机制

3. 监控告警配置
   - 监控限流指标
   - 设置合理的告警阈值
   - 记录详细的限流日志

4. 性能优化建议
   - 使用本地缓存
   - 批量处理请求
   - 异步处理指标收集
   - 合理设置过期时间
