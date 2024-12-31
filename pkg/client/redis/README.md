# Redis Client Package

## 目录
- [简介](#简介)
- [特性](#特性)
- [安装](#安装)
- [快速开始](#快速开始)
- [配置选项](#配置选项)
- [基础操作](#基础操作)
- [高级特性](#高级特性)
- [监控与追踪](#监控与追踪)
- [错误处理](#错误处理)
- [测试](#测试)

## 简介

这是一个基于 go-redis/redis 的 Redis 客户端封装包,提供了更加易用和功能完善的 Redis 操作接口。该包支持单机模式和集群模式,并集成了监控、链路追踪等企业级特性。

## 特性

- 支持单机和集群模式
- 完整的 Redis 数据类型操作(String, Hash, List, Set, ZSet)
- 支持 Pipeline 和事务
- 自动重试机制
- 连接池管理
- TLS 加密支持
- Prometheus 指标监控
- OpenTracing 分布式追踪
- 统一的错误处理
- 完善的日志记录
- 丰富的测试用例

## 安装

```bash
go get gobase/pkg/client/redis
```

## 快速开始

```go
import "gobase/pkg/client/redis"

// 创建客户端
client, err := redis.NewClient(
    redis.WithAddress("localhost:6379"),
    redis.WithPassword("password"),
)
if err != nil {
    panic(err)
}
defer client.Close()

// 基本操作
ctx := context.Background()
err = client.Set(ctx, "key", "value", time.Minute)
if err != nil {
    panic(err)
}

value, err := client.Get(ctx, "key")
if err != nil {
    panic(err)
}
```

## 配置选项

支持以下配置选项:

```go
type Config struct {
    // 基础配置
    Addresses []string  // Redis地址列表
    Username  string    // 用户名
    Password  string    // 密码
    Database  int      // 数据库编号

    // 连接池配置
    PoolSize     int   // 连接池大小
    MinIdleConns int   // 最小空闲连接数

    // TLS配置
    EnableTLS   bool   // 是否启用TLS
    TLSCertFile string // 证书文件路径
    TLSKeyFile  string // 密钥文件路径

    // 集群配置
    EnableCluster bool // 是否启用集群模式
    RouteRandomly bool // 是否随机路由

    // 监控配置
    EnableMetrics    bool   // 是否启用指标收集
    MetricsNamespace string // 指标命名空间
    EnableTracing    bool   // 是否启用链路追踪
}
```

## 基础操作

### String 操作
```go
Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
Get(ctx context.Context, key string) (string, error)
Del(ctx context.Context, keys ...string) (int64, error)
```

### Hash 操作
```go
HSet(ctx context.Context, key string, values ...interface{}) (int64, error)
HGet(ctx context.Context, key, field string) (string, error)
HDel(ctx context.Context, key string, fields ...string) (int64, error)
```

### List 操作
```go
LPush(ctx context.Context, key string, values ...interface{}) (int64, error)
LPop(ctx context.Context, key string) (string, error)
```

### Set 操作
```go
SAdd(ctx context.Context, key string, members ...interface{}) (int64, error)
SRem(ctx context.Context, key string, members ...interface{}) (int64, error)
```

### ZSet 操作
```go
ZAdd(ctx context.Context, key string, members ...*Z) (int64, error)
ZRem(ctx context.Context, key string, members ...interface{}) (int64, error)
```

## 高级特性

### Pipeline
```go
pipe := client.TxPipeline()
pipe.Set(ctx, "key1", "value1", time.Minute)
pipe.Set(ctx, "key2", "value2", time.Minute)
cmds, err := pipe.Exec(ctx)
```

### 连接池管理
```go
stats := client.Pool().Stats()
fmt.Printf("Total connections: %d\n", stats.TotalConns)
fmt.Printf("Idle connections: %d\n", stats.IdleConns)
```

## 监控与追踪

### Prometheus 指标
```go
client, err := redis.NewClient(
    redis.WithAddress("localhost:6379"),
    redis.WithMetrics(true),
    redis.WithMetricsNamespace("myapp_redis"),
)
```

### OpenTracing 追踪
```go
client, err := redis.NewClient(
    redis.WithAddress("localhost:6379"),
    redis.WithTracing(true),
)
```

## 错误处理

包内置了统一的错误处理机制:

```go
var (
    ErrInvalidConfig     = errors.NewError(codes.CacheError, "invalid redis config", nil)
    ErrConnectionFailed  = errors.NewError(codes.CacheError, "failed to connect", nil)
    ErrOperationFailed   = errors.NewError(codes.CacheError, "operation failed", nil)
    ErrKeyNotFound       = errors.NewError(codes.NotFound, "key not found", nil)
    ErrPoolTimeout       = errors.NewError(codes.CacheError, "pool timeout", nil)
)
```

## 测试

包含完整的测试套件:

- 单元测试
- 集成测试
- 基准测试
- 压力测试
- 恢复测试

运行测试:

```bash
# 运行所有测试
go test ./...

# 运行基准测试
go test -bench=. ./...

# 运行压力测试
go test -tags=stress ./...
```

