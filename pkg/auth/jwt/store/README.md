# JWT Token Store

JWT Token 存储实现,支持内存和Redis两种存储方式。

## 目录
- [功能特性](#功能特性)
- [接口定义](#接口定义)
- [存储实现](#存储实现)
  - [内存存储](#内存存储)
  - [Redis存储](#redis存储)
- [配置选项](#配置选项)
- [使用示例](#使用示例)
- [性能测试](#性能测试)
- [测试覆盖](#测试覆盖)

## 功能特性

- 支持内存和Redis两种存储方式
- 完整的链路追踪支持(Jaeger)
- 完整的监控指标(Prometheus)
- 统一的日志处理(Logrus)
- 自动清理过期Token(内存模式)
- 用户Token映射关系维护
- 线程安全实现
- 高性能设计

## 接口定义

Store 接口定义了 JWT Token 存储的基本操作:

```go
// Store 定义JWT存储接口
type Store interface {
    // Set 存储JWT令牌
    Set(ctx context.Context, key string, value *jwt.TokenInfo, expiration time.Duration) error
    
    // Get 获取JWT令牌
    Get(ctx context.Context, key string) (*jwt.TokenInfo, error)
    
    // Delete 删除JWT令牌
    Delete(ctx context.Context, key string) error
    
    // Close 关闭存储连接
    Close() error
}
```

## 存储实现

### 内存存储

内存存储实现(`MemoryStore`)特点:

- 使用 map 存储 Token 信息
- 维护用户-Token映射关系
- 支持自动清理过期Token
- 使用读写锁保证并发安全
- 支持监控指标收集
- 支持链路追踪

### Redis存储 

Redis存储实现(`RedisTokenStore`)特点:

- 支持 Redis 单机/集群模式
- 自动序列化/反序列化 Token 信息
- 利用 Redis TTL 机制自动过期
- 支持监控指标收集
- 支持链路追踪

## 配置选项

```go
// StoreType 存储类型
type StoreType string

const (
    // TypeMemory 内存存储类型
    TypeMemory StoreType = "memory"
    // TypeRedis Redis存储类型 
    TypeRedis StoreType = "redis"
)

// Config 存储配置
type Config struct {
    // Type 存储类型
    Type StoreType `json:"type" yaml:"type"`
    // Host 主机地址
    Host string `json:"host" yaml:"host"`
    // Port 端口号
    Port int `json:"port" yaml:"port"`
    // Password 密码
    Password string `json:"password" yaml:"password"`
    // DB 数据库
    DB int `json:"db" yaml:"db"`
}

// Options 存储配置选项
type Options struct {
    // Redis配置
    Redis *RedisOptions
    
    // 监控指标
    EnableMetrics bool
    
    // 链路追踪
    EnableTracing bool
    
    // 前缀
    KeyPrefix string
    
    // CleanupInterval 清理过期会话的时间间隔
    // 如果设置为0或负值，则不进行自动清理
    // 仅用于内存存储模式
    CleanupInterval time.Duration
}

// RedisOptions Redis配置选项
type RedisOptions struct {
    Addr     string
    Password string
    DB       int
}
```
## 使用示例

```go
// 使用示例:
// 创建内存存储
memStore, err := store.NewMemoryStore(store.Options{
    CleanupInterval: time.Minute,
    EnableMetrics:   true,
    EnableTracing:   true,
})

// 创建Redis存储
redisStore := store.NewRedisTokenStore(redisClient, &store.Options{
    KeyPrefix:      "jwt:",
    EnableMetrics:  true,
    EnableTracing:  true,
}, logger)

// 存储Token
err := store.Set(ctx, token, tokenInfo, time.Hour)

// 获取Token
info, err := store.Get(ctx, token)

// 删除Token
err := store.Delete(ctx, token)
```

## 性能测试

详细的性能测试结果请参考: [性能测试报告](tests/benchmark/README.md)

主要性能指标:

- 内存存储:
  - Set: 183万 ops/sec
  - Get: 2400万 ops/sec
  
- Redis存储:
  - Set: 2.6万 ops/sec
  - Get: 3.2万 ops/sec

## 测试覆盖

- 单元测试
  - [内存存储测试](tests/unit/memory_test.go)
  - [Redis存储测试](tests/unit/redis_test.go)
  - [接口测试](tests/unit/interface_test.go)

- 集成测试
  - [存储集成测试](tests/integration/store_test.go)

- 压力测试
  - [存储压力测试](tests/stress/store_stress_test.go)

- 基准测试
  - [存储基准测试](tests/benchmark/store_bench_test.go)