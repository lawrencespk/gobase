# 目录

1. [概述](#1-概述)
2. [配置详解](#2-配置详解)
3. [监控与指标](#3-监控与指标)
4. [高级特性](#4-高级特性)
5. [性能优化](#5-性能优化)
6. [最佳实践](#6-最佳实践)
7. [错误处理](#7-错误处理)
8. [常见问题和故障排除](#常见问题和故障排除)

# RateLimit 中间件

RateLimit 中间件提供了一个高性能、可扩展的限流解决方案，支持多种限流算法和分布式场景。

## 1. 概述

### 1.1 功能特性

- 支持多种限流算法（固定窗口、滑动窗口、令牌桶、漏桶）
- 分布式限流支持（基于 Redis）
- 灵活的限流策略配置
- 支持等待模式和直接拒绝模式
- 内置重试机制和指数退避
- 完整的监控指标
- 自定义响应格式
- 故障转移机制
- 高性能和低延迟

### 1.2 核心接口

```go
// Limiter 定义了限流器的核心接口
type Limiter interface {
    // Allow 检查是否允许请求通过
    Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error)
    
    // AllowN 检查是否允许N个请求通过
    AllowN(ctx context.Context, key string, n int64, limit int64, window time.Duration) (bool, error)
    
    // Wait 等待直到允许请求通过或超时
    Wait(ctx context.Context, key string, limit int64, window time.Duration) error
    
    // Reset 重置限流器状态
    Reset(ctx context.Context, key string) error
}
```

### 1.3 快速开始

```go
import (
    "github.com/gin-gonic/gin"
    "gobase/pkg/middleware/ratelimit"
    redisLimiter "gobase/pkg/ratelimit/redis"
)

func main() {
    // 创建 Redis 限流器
    limiter := redisLimiter.NewSlidingWindowLimiter(redisClient)
    
    // 创建路由
    r := gin.New()
    
    // 使用限流中间件
    r.Use(ratelimit.RateLimit(&ratelimit.Config{
        Limiter: limiter,
        Limit:   100,           // 每个时间窗口允许的请求数
        Window:  time.Minute,   // 时间窗口大小
    }))
}
```

## 2. 配置详解

### 2.1 基础配置

```go
type Config struct {
    // 必需配置
    Limiter Limiter           // 限流器实现
    Limit   int64            // 限流阈值
    Window  time.Duration    // 时间窗口
    
    // 可选配置
    KeyFunc     func(*gin.Context) string  // 限流键生成函数
    WaitMode    bool                       // 是否启用等待模式
    WaitTimeout time.Duration             // 等待超时时间
    Debug       bool                       // 是否启用调试模式
}
```

配置示例：

```go
config := &ratelimit.Config{
    Limiter: limiter,
    KeyFunc: func(c *gin.Context) string {
        return c.ClientIP()  // 基于 IP 的限流
    },
    Limit:       100,
    Window:      time.Minute,
    WaitMode:    true,
    WaitTimeout: 5 * time.Second,
    Debug:       true,
}
```

### 2.2 限流策略配置

#### 2.2.1 固定窗口限流

```go
limiter := redisLimiter.NewFixedWindowLimiter(redisClient, &redisLimiter.Config{
    KeyPrefix:    "ratelimit:fixed:",
    LockTimeout:  time.Second,
    RetryCount:   3,
})
```

#### 2.2.2 滑动窗口限流

```go
limiter := redisLimiter.NewSlidingWindowLimiter(redisClient, &redisLimiter.Config{
    KeyPrefix:    "ratelimit:sliding:",
    LockTimeout:  time.Second,
    RetryCount:   3,
    Precision:    10,  // 将窗口分割为10个小区间
})
```

#### 2.2.3 令牌桶限流

```go
limiter := redisLimiter.NewTokenBucketLimiter(redisClient, &redisLimiter.TokenBucketConfig{
    Capacity:      100,    // 桶容量
    Rate:          10,     // 令牌生成速率
    InitialTokens: 0,      // 初始令牌数
    KeyPrefix:     "ratelimit:token:",
})
```

#### 2.2.4 漏桶限流

```go
limiter := redisLimiter.NewLeakyBucketLimiter(redisClient, &redisLimiter.LeakyBucketConfig{
    Capacity:   100,    // 桶容量
    Rate:       10,     // 处理速率
    KeyPrefix:  "ratelimit:leaky:",
})
```

### 2.3 重试策略配置

```go
config := &ratelimit.Config{
    Retry: &ratelimit.RetryStrategy{
        MaxAttempts:           3,                    // 最大重试次数
        RetryInterval:         100 * time.Millisecond, // 重试间隔
        UseExponentialBackoff: true,                // 使用指数退避
        MaxRetryInterval:      2 * time.Second,     // 最大重试间隔
    },
}
```

### 2.4 错误处理配置

```go
config := &ratelimit.Config{
    Messages: &ratelimit.ErrorMessages{
        CheckFailedMessage:   "限流检查失败",
        LimitExceededMessage: "请求过于频繁",
    },
    Status: &ratelimit.StatusCodes{
        CheckFailed:   http.StatusInternalServerError,
        LimitExceeded: http.StatusTooManyRequests,
    },
    ResponseHandler: func(c *gin.Context, err error) {
        c.JSON(http.StatusTooManyRequests, gin.H{
            "code":    429,
            "message": "请求过于频繁，请稍后重试",
            "retry_after": 60,
        })
    },
}
```

## 3. 监控与指标

### 3.1 内置指标

```go
// 请求总数
RequestsTotal = metric.NewCounter(metric.CounterOpts{
    Namespace: "gobase",
    Subsystem: "ratelimit",
    Name:      "requests_total",
    Help:      "Total number of requests handled by rate limiter",
}).WithLabels("key", "result")

// 被拒绝的请求数
RejectedTotal = metric.NewCounter(metric.CounterOpts{
    Namespace: "gobase",
    Subsystem: "ratelimit",
    Name:      "rejected_total",
    Help:      "Total number of requests rejected by rate limiter",
}).WithLabels("key")

// 限流器操作延迟
LimiterLatency = metric.NewHistogram(metric.HistogramOpts{
    Namespace: "gobase",
    Subsystem: "ratelimit",
    Name:      "latency_seconds",
    Help:      "Latency of rate limiter operations",
}).WithLabels("key", "operation")
```

### 3.2 Prometheus 集成

```go
r.Use(
    middleware.Metrics(),
    ratelimit.RateLimit(&ratelimit.Config{
        Limiter: limiter,
        OnRejected: func(c *gin.Context) {
            metrics.IncCounter("ratelimit_rejected_total",
                "path", c.Request.URL.Path,
            )
        },
    }),
)
```

### 3.3 监控最佳实践

1. 监控关键指标
   - 请求通过率
   - 限流器延迟
   - 重试次数分布
   
2. 设置告警规则
   ```yaml
   groups:
   - name: ratelimit_alerts
     rules:
     - alert: HighRejectionRate
       expr: rate(gobase_ratelimit_rejected_total[5m]) > 0.1
       for: 5m
       labels:
         severity: warning
       annotations:
         summary: "High rate limit rejection rate"
   ```

## 4. 高级特性

### 4.1 分布式限流

```go
// Redis 集群配置
redisClient := redis.NewClusterClient(&redis.ClusterOptions{
    Addrs: []string{
        "redis-node1:6379",
        "redis-node2:6379",
        "redis-node3:6379",
    },
    RouteByLatency: true,
    ReadOnly:       true,
})

// 创建分布式限流器
limiter := redisLimiter.NewSlidingWindowLimiter(redisClient, &redisLimiter.Config{
    KeyPrefix:   "app:ratelimit:",
    LockTimeout: time.Second,
    RetryCount:  3,
})
```

### 4.2 多级限流

```go
// IP 级别限流
ipLimiter := ratelimit.RateLimit(&ratelimit.Config{
    Limiter: limiter,
    KeyFunc: func(c *gin.Context) string {
        return c.ClientIP()
    },
    Limit:  1000,
    Window: time.Minute,
})

// 用户级别限流
userLimiter := ratelimit.RateLimit(&ratelimit.Config{
    Limiter: limiter,
    KeyFunc: func(c *gin.Context) string {
        return c.GetHeader("X-User-ID")
    },
    Limit:  100,
    Window: time.Minute,
})

// API 级别限流
apiLimiter := ratelimit.RateLimit(&ratelimit.Config{
    Limiter: limiter,
    KeyFunc: func(c *gin.Context) string {
        return c.Request.URL.Path
    },
    Limit:  5000,
    Window: time.Minute,
})

// 组合使用
r.Use(ipLimiter, userLimiter, apiLimiter)
```

### 4.3 动态配置

```go
r.Use(ratelimit.RateLimit(&ratelimit.Config{
    Limiter: limiter,
    // 动态限流阈值
    LimitFunc: func(c *gin.Context) int64 {
        if isVipUser(c) {
            return 1000
        }
        return 100
    },
    // 动态时间窗口
    WindowFunc: func(c *gin.Context) time.Duration {
        if isVipUser(c) {
            return time.Minute
        }
        return 5 * time.Minute
    },
}))
```

### 4.4 故障转移

```go
config := &ratelimit.Config{
    Limiter: limiter,
    FailurePolicy: &ratelimit.FailurePolicy{
        FallbackLimit:  10,           // 故障时的降级限制
        FailureTimeout: time.Second,  // 故障检测超时
        RetryInterval:  time.Second,  // 重试间隔
        OnFailure: func(err error) {
            log.Printf("限流器故障: %v", err)
            alerting.Send("限流器故障告警")
        },
    },
}
```

## 5. 性能优化

### 5.1 基准测试数据

| 限流器类型 | QPS | 平均延迟 | P99延迟 |
|-----------|-----|---------|---------|
| 固定窗口   | 50k | 0.2ms   | 1ms     |
| 滑动窗口   | 30k | 0.5ms   | 2ms     |
| 令牌桶     | 40k | 0.3ms   | 1.5ms   |
| 漏桶       | 35k | 0.4ms   | 1.8ms   |

### 5.2 性能优化建议

1. Redis 连接池优化
```go
redisClient := redis.NewClient(&redis.Options{
    PoolSize:     10,
    MinIdleConns: 5,
    MaxConnAge:   time.Hour,
})
```

2. 本地缓存优化
```go
config := &ratelimit.Config{
    Limiter: redisLimiter.NewSlidingWindowLimiter(redisClient),
    UseLocalCache:  true,
    LocalCacheSize: 1000,
    LocalCacheTTL:  time.Second,
}
```

### 5.3 资源使用优化

1. 合理设置限流键过期时间
2. 避免过大的滑动窗口精度
3. 使用合适的重试策略

## 6. 最佳实践

### 6.1 限流算法选择

- 一般场景：滑动窗口
- 允许突发：令牌桶
- 严格控制：漏桶
- 简单需求：固定窗口

### 6.2 配置推荐

1. 基础服务配置
```go
config := &ratelimit.Config{
    Limiter:     limiter,
    Limit:       1000,
    Window:      time.Minute,
    WaitMode:    false,
    RetryCount:  3,
}
```

2. 关键服务配置
```go
config := &ratelimit.Config{
    Limiter:     limiter,
    Limit:       100,
    Window:      time.Minute,
    WaitMode:    true,
    WaitTimeout: 5 * time.Second,
    RetryCount:  5,
}
```

### 6.3 常见问题解决

1. Redis 连接问题
```go
config := &ratelimit.Config{
    Limiter: redisLimiter.NewSlidingWindowLimiter(redisClient),
    FailurePolicy: &ratelimit.FailurePolicy{
        FallbackLimit:  10,
        FailureTimeout: time.Second,
        RetryInterval:  time.Second,
    },
}
```

2. 限流键冲突
```go
config := &ratelimit.Config{
    KeyFunc: func(c *gin.Context) string {
        return fmt.Sprintf("%s:%s:%s",
            c.Request.URL.Path,
            c.ClientIP(),
            c.GetHeader("X-User-ID"),
        )
    },
    KeyPrefix: "app:ratelimit:",
}
```

### 6.4 生产环境部署

1. Redis 集群配置
2. 监控告警配置
3. 日志配置
4. 故障转移策略

## 7. 错误处理

中间件会在以下情况返回错误：

- 限流检查失败：500 状态码
- 超过限流阈值：429 状态码
- 等待超时：500 状态码

自定义错误处理：

```go
config := &ratelimit.Config{
    ResponseHandler: func(c *gin.Context, err error) {
        switch {
        case errors.Is(err, ratelimit.ErrLimitExceeded):
            c.JSON(429, gin.H{
                "code":    429,
                "message": "请求过于频繁",
                "retry_after": 60,
            })
        default:
            c.JSON(500, gin.H{
                "code":    500,
                "message": "系统繁忙",
            })
        }
    },
}
```

## 限流算法对比

### 固定窗口限流
- **适用场景**: 简单的QPS限制
- **优点**: 实现简单，内存占用小
- **缺点**: 临界时间点可能突发流量
- **示例**:
```go
limiter := redisLimiter.NewFixedWindowLimiter(redisClient)
```

### 滑动窗口限流
- **适用场景**: 需要平滑限流的场景
- **优点**: 流量更平滑，避免突发
- **缺点**: 实现相对复杂，内存占用较大
- **示例**:
```go
limiter := redisLimiter.NewSlidingWindowLimiter(redisClient)
```

### 令牌桶限流
- **适用场景**: 允许突发流量，但需要限制平均速率
- **优点**: 支持突发流量，具有缓冲能力
- **缺点**: 令牌生成和消费的计算开销
- **示例**:
```go
limiter := redisLimiter.NewTokenBucketLimiter(redisClient, &redisLimiter.TokenBucketConfig{
    Capacity: 100,    // 桶容量
    Rate:     10,     // 令牌生成速率
    InitialTokens: 0, // 初始令牌数
})
```

### 漏桶限流
- **适用场景**: 需要严格控制处理速率
- **优点**: 流量最平滑，严格控制处理速率
- **缺点**: 无法处理突发流量
- **示例**:
```go
limiter := redisLimiter.NewLeakyBucketLimiter(redisClient, &redisLimiter.LeakyBucketConfig{
    Capacity: 100, // 桶容量
    Rate:     10,  // 处理速率
})
```

## 与其他中间件集成

### 与 Logger 中间件集成
```go
r.Use(
    middleware.Logger(),
    ratelimit.RateLimit(&ratelimit.Config{
        Limiter: limiter,
        OnRejected: func(c *gin.Context) {
            logger.WithContext(c).Warn("请求被限流",
                "path", c.Request.URL.Path,
                "ip", c.ClientIP(),
            )
        },
    }),
)
```

### 与 Trace 中间件集成
```go
r.Use(
    middleware.Trace(),
    ratelimit.RateLimit(&ratelimit.Config{
        Limiter: limiter,
        KeyFunc: func(c *gin.Context) string {
            // 使用 TraceID 作为限流键的一部分
            return fmt.Sprintf("%s:%s", 
                c.GetString("trace_id"),
                c.ClientIP(),
            )
        },
    }),
)
```

### 与 Metrics 中间件集成
```go
r.Use(
    middleware.Metrics(),
    ratelimit.RateLimit(&ratelimit.Config{
        Limiter: limiter,
        OnRejected: func(c *gin.Context) {
            metrics.IncCounter("ratelimit_rejected_total", 
                "path", c.Request.URL.Path,
            )
        },
    }),
)
```

## 常见问题和故障排除

### 1. Redis 连接问题

**症状**: 限流器频繁返回错误
**解决方案**:
```go
// 配置 Redis 重试和容错
redisClient := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
    MaxRetries: 3,
    MaxRetryBackoff: time.Second,
})

config := &ratelimit.Config{
    Limiter: redisLimiter.NewSlidingWindowLimiter(redisClient),
    FailurePolicy: &ratelimit.FailurePolicy{
        FallbackLimit: 10,           // 故障时的降级限制
        FailureTimeout: time.Second, // 故障检测超时
        RetryInterval: time.Second,  // 重试间隔
    },
}
```

### 2. 性能问题

**症状**: 限流操作延迟高
**解决方案**:
```go
// 优化 Redis 连接池
redisClient := redis.NewClient(&redis.Options{
    PoolSize:     10,
    MinIdleConns: 5,
    MaxConnAge:   time.Hour,
})

// 使用本地缓存优化
config := &ratelimit.Config{
    Limiter: redisLimiter.NewSlidingWindowLimiter(redisClient),
    UseLocalCache: true,
    LocalCacheSize: 1000,
    LocalCacheTTL: time.Second,
}
```

### 3. 限流键冲突

**症状**: 不同请求被错误地限流
**解决方案**:
```go
config := &ratelimit.Config{
    Limiter: limiter,
    KeyFunc: func(c *gin.Context) string {
        // 使用多个维度组合作为限流键
        return fmt.Sprintf("%s:%s:%s",
            c.Request.URL.Path,
            c.ClientIP(),
            c.GetHeader("X-User-ID"),
        )
    },
    KeyPrefix: "app:ratelimit:", // 添加业务前缀
}
```

## 性能测试基准

在标准硬件配置下（8核CPU，16GB内存），各种限流器的性能表现：

| 限流器类型 | QPS | 平均延迟 | P99延迟 |
|-----------|-----|---------|---------|
| 固定窗口   | 50k | 0.2ms   | 1ms     |
| 滑动窗口   | 30k | 0.5ms   | 2ms     |
| 令牌桶     | 40k | 0.3ms   | 1.5ms   |
| 漏桶       | 35k | 0.4ms   | 1.8ms   |

> 注：以上数据基于本地 Redis 实例测试，实际性能会受网络环境影响。

## 最佳实践建议

1. **选择合适的限流算法**
   - 一般场景：滑动窗口
   - 允许突发：令牌桶
   - 严格控制：漏桶
   - 简单需求：固定窗口

2. **优化限流键设计**
   - 避免键空间过大
   - 合理设置过期时间
   - 使用前缀区分业务

3. **配置合理的重试策略**
   - 设置最大重试次数
   - 使用指数退避
   - 设置重试超时

4. **监控和告警**
   - 监控限流指标
   - 设置合理的告警阈值
   - 记录详细的限流日志