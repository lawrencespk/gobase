# 多级缓存 (Multilevel Cache)

## 目录
- [简介](#简介)
- [特性](#特性)
- [架构设计](#架构设计)
- [使用方法](#使用方法)
- [配置说明](#配置说明)
- [性能指标](#性能指标)
- [最佳实践](#最佳实践)
- [注意事项](#注意事项)

## 简介
多级缓存是一个高性能的分布式缓存解决方案，结合了本地内存缓存(L1)和Redis分布式缓存(L2)的优势。它提供了一致性的缓存访问接口，自动的缓存同步，以及灵活的配置选项。

## 特性
- 两级缓存架构
  - L1: 本地内存缓存，提供快速访问
  - L2: Redis分布式缓存，提供持久化存储
- 自动缓存同步
  - L2 -> L1 自动回填
  - 写入时同时更新 L1 和 L2
- 智能缓存预热
  - 支持手动预热
  - 支持自动预热
  - 可配置预热并发数
- 完整的监控指标
  - 操作计数
  - 命中率统计
  - 错误监控
- 分布式友好
  - 支持集群部署
  - 保证缓存一致性
- 完善的错误处理
  - 统一的错误码
  - 详细的错误信息

## 架构设计
```
┌─────────────────┐
│    应用程序      │
└────────┬────────┘
         │
┌────────┴────────┐
│  缓存管理器      │
├─────────────────┤
│ L1 内存缓存     │
├─────────────────┤
│ L2 Redis 缓存   │
└─────────────────┘
```

## 使用方法
```go
// 创建配置
config := &multilevel.Config{
    L1Config: &multilevel.L1Config{
        MaxEntries:      1000,
        CleanupInterval: time.Minute,
    },
    L2Config: &multilevel.L2Config{
        RedisAddr:     "localhost:6379",
        RedisPassword: "",
        RedisDB:       0,
    },
    L1TTL:             time.Hour,
    EnableAutoWarmup:  true,
    WarmupInterval:    time.Hour,
    WarmupConcurrency: 10,
}

// 创建缓存管理器
manager, err := multilevel.NewManager(config, redisClient, logger)
if err != nil {
    return err
}

// 基本操作
err = manager.Set(ctx, "key", "value", time.Hour)
value, err := manager.Get(ctx, "key")
err = manager.Delete(ctx, "key")

// 缓存预热
err = manager.Warmup(ctx, []string{"key1", "key2"})
```

## 配置说明
### L1配置
```go
type L1Config struct {
    MaxEntries      int           // 最大条目数
    CleanupInterval time.Duration // 清理间隔
}
```

### L2配置
```go
type L2Config struct {
    RedisAddr     string // Redis地址
    RedisPassword string // Redis密码
    RedisDB       int    // Redis数据库
}
```

### 全局配置
```go
type Config struct {
    L1Config          *L1Config     // L1缓存配置
    L2Config          *L2Config     // L2缓存配置
    L1TTL             time.Duration // L1缓存TTL
    EnableAutoWarmup  bool          // 是否启用自动预热
    WarmupInterval    time.Duration // 预热间隔
    WarmupConcurrency int          // 预热并发数
}
```

## 性能指标
详细的性能测试报告请参考: [性能测试报告](/pkg/cache/multilevel/tests/benchmark/README.md)

基准测试包括：
- 基本操作性能测试
- 并发访问测试
- 压力测试
- 内存使用分析

## 最佳实践
1. 合理配置缓存大小
   - L1缓存建议配置为热点数据大小
   - 避免过大的L1缓存占用过多内存

2. 预热策略
   - 系统启动时预热核心数据
   - 根据业务访问模式配置自动预热

3. 监控指标
   - 监控缓存命中率
   - 关注错误率变化
   - 观察内存使用情况

## 注意事项
1. L1缓存过期时间
   - L1缓存TTL应小于等于L2缓存TTL
   - 建议L1缓存TTL设置为1小时以内

2. 预热并发控制
   - 合理设置预热并发数
   - 避免过高并发影响系统性能

3. 错误处理
   - L1缓存错误不会影响L2缓存
   - 建议做好错误监控和告警

4. 内存管理
   - 定期监控L1缓存内存使用
   - 适时调整MaxEntries参数 