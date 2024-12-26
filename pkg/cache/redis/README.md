# Redis 缓存包

这个包提供了一个简单且功能完整的 Redis 客户端封装,支持基础操作、哈希表、列表、集合、有序集合等 Redis 数据结构的操作。

## 特性

- 简单易用的 API 接口
- 完整的 Redis 数据类型支持
- 连接池管理
- 支持事务和管道操作
- 可配置的重试机制
- 支持 TLS 加密连接
- 支持集群模式
- 内置监控和链路追踪

## 快速开始

### 基础配置

```go
import (
"gobase/pkg/cache/redis/client"
"gobase/pkg/cache/redis/config/types"
)
cfg := &types.Config{
Addresses: []string{"localhost:6379"},
Password: "your-password",
Database: 0,
PoolSize: 10,
}
redisClient, err := client.NewClient(cfg)
if err != nil {
panic(err)
}
defer redisClient.Close()
```

### 基础操作示例

```go
ctx := context.Background()
// 设置键值对
err := redisClient.Set(ctx, "key", "value", 1time.Hour)
// 获取值
val, err := redisClient.Get(ctx, "key")
// 删除键
err = redisClient.Del(ctx, "key")
// 计数器操作
count, err := redisClient.Incr(ctx, "counter")
count, err = redisClient.IncrBy(ctx, "counter", 10)
```

### Hash 操作示例

```go
// 设置哈希表字段
err := redisClient.HSet(ctx, "hash-key", "field1", "value1", "field2", "value2")
// 获取哈希表字段
val, err := redisClient.HGet(ctx, "hash-key", "field1")
```

### List 操作示例

```go
// 向列表左端推入元素
err := redisClient.LPush(ctx, "list-key", "value1", "value2")
// 从列表左端弹出元素
val, err := redisClient.LPop(ctx, "list-key")
```

### Set 操作示例

```go
// 添加集合成员
err := redisClient.SAdd(ctx, "set-key", "member1", "member2")
// 移除集合成员
err = redisClient.SRem(ctx, "set-key", "member1")
```

### ZSet 操作示例

```go
// 添加有序集合成员
err := redisClient.ZAdd(ctx, "zset-key", 1.0, "member1", 2.0, "member2")
// 移除有序集合成员
err = redisClient.ZRem(ctx, "zset-key", "member1")
```

### 事务和管道操作

```go
// 开启事务
pipe := redisClient.TxPipeline()
// 执行命令
pipe.Set(ctx, "key1", "value1", 0)
pipe.Set(ctx, "key2", "value2", 0)
// 提交事务
err := pipe.Exec(ctx)
// 放弃事务
pipe.Discard()
```

## 配置说明

完整的配置选项:

```go
type Config struct {
// 基础配置
Addresses []string // Redis地址列表
Username string // 用户名
Password string // 密码
Database int // 数据库索引
// 连接池配置
PoolSize int // 连接池大小
MinIdleConns int // 最小空闲连接数
MaxRetries int // 最大重试次数
// 超时配置
DialTimeout time.Duration // 连接超时
ReadTimeout time.Duration // 读取超时
WriteTimeout time.Duration // 写入超时
// TLS配置
EnableTLS bool // 是否启用TLS
TLSCertFile string // TLS证书文件
TLSKeyFile string // TLS密钥文件
// 集群配置
EnableCluster bool // 是否启用集群
RouteRandomly bool // 是否随机路由
// 监控配置
EnableMetrics bool // 是否启用监控
EnableTracing bool // 是否启用链路追踪
}
```

## 错误处理

该包使用统一的错误处理机制,所有的操作都会返回标准的 error 接口。常见错误包括:

- 连接错误
- 超时错误
- 键不存在
- 类型不匹配
- 参数错误

建议在使用时总是检查返回的错误:

```go
val, err := redisClient.Get(ctx, "key")
if err != nil {
if err == redis.Nil {
// 键不存在
} else {
// 其他错误
}
}
```

## 测试

包含完整的单元测试和集成测试:

```bash
运行单元测试
go test ./pkg/cache/redis/...
运行集成测试 (需要 Docker)
go test ./pkg/cache/redis/... -tags=integration
```

## 性能建议

1. 合理设置连接池大小
2. 使用管道批量处理命令
3. 避免大键值对
4. 合理设置过期时间
5. 监控连接池状态

## 注意事项

1. 在生产环境中务必启用密码认证
2. 建议启用 TLS 加密传输
3. 合理配置超时时间
4. 注意键的命名规范
5. 建议启用监控和链路追踪

## 最佳实践

### 1. 键名设计规范
```go
// 推荐的命名方式
user:profile:1001    // 用户配置
order:detail:88888   // 订单详情
product:stock:66666  // 商品库存
```

### 2. 错误重试处理
```go
import "gobase/pkg/cache/redis/retry"

// 配置重试策略
retryOpts := &retry.Options{
    MaxRetries: 3,
    MinBackoff: 100 * time.Millisecond,
    MaxBackoff: 1 * time.Second,
}

// 使用重试机制执行操作
val, err := retry.Do(ctx, func() (interface{}, error) {
    return redisClient.Get(ctx, "key")
}, retryOpts)
```

### 3. 分布式锁实现
```go
import "gobase/pkg/cache/redis/lock"

// 获取分布式锁
lock := redisClient.NewLock("my-lock", &lock.Options{
    TTL: 10 * time.Second,
})

// 加锁
ok, err := lock.Lock(ctx)
if err != nil || !ok {
    return errors.New("获取锁失败")
}

// 解锁
defer lock.Unlock(ctx)
```

### 4. 监控集成
```go
import "gobase/pkg/cache/redis/metrics"

// 启用 Prometheus 监控
metrics.EnablePrometheus(redisClient, "my_service")

// 自定义监控指标
metrics.RegisterCounter("redis_custom_counter", "自定义计数器")
metrics.IncrCounter("redis_custom_counter")
```

### 5. 链路追踪
```go
import "gobase/pkg/cache/redis/trace"

// 启用 Jaeger 追踪
trace.EnableJaeger(redisClient, "my_service")

// 带追踪的上下文操作
ctx = trace.NewContext(ctx, "redis-operation")
val, err := redisClient.Get(ctx, "key")
```

## 高级特性

### 1. Lua 脚本支持
```go
script := redis.NewScript(`
    local key = KEYS[1]
    local value = ARGV[1]
    local ttl = tonumber(ARGV[2])
    
    local oldValue = redis.call('GET', key)
    if oldValue then
        return oldValue
    end
    
    redis.call('SET', key, value, 'EX', ttl)
    return value
`)

val, err := script.Run(ctx, redisClient,
    []string{"my-key"},   // KEYS
    "default-value", 300, // ARGV
).Result()
```

### 2. 发布订阅
```go
// 订阅频道
pubsub := redisClient.Subscribe(ctx, "news-channel")
defer pubsub.Close()

// 接收消息
for msg := range pubsub.Channel() {
    fmt.Printf("收到消息: %s\n", msg.Payload)
}

// 发布消息
err := redisClient.Publish(ctx, "news-channel", "Hello World").Err()
```

### 3. 集群支持
```go
import "gobase/pkg/cache/redis/cluster"

clusterClient, err := cluster.NewClient(&cluster.Options{
    Addrs: []string{
        "localhost:7000",
        "localhost:7001",
        "localhost:7002",
    },
    RouteRandomly: true,
})
```

## 故障排除

### 1. 常见问题

1. 连接超时
```go
// 增加连接超时时间
cfg.DialTimeout = 10 * time.Second
```

2. 内存占用过高
```go
// 减小连接池大小
cfg.PoolSize = 5
cfg.MinIdleConns = 2
```

3. 命令执行超时
```go
// 设置操作超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

### 2. 调试方法

1. 启用详细日志
```go
import "gobase/pkg/cache/redis/log"

log.SetLevel(log.LevelDebug)
```

2. 监控连接池状态
```go
stats := redisClient.PoolStats()
fmt.Printf("总连接数: %d\n", stats.TotalConns)
fmt.Printf("空闲连接数: %d\n", stats.IdleConns)
```

## 依赖

- go-redis/redis/v8: ^8.11.5
- prometheus/client_golang: ^1.11.0
- opentracing/opentracing-go: ^1.2.0
- uber/jaeger-client-go: ^2.30.0
- gobase 内部包

## 版本兼容性

- Go version >= 1.16
- Redis version >= 5.0