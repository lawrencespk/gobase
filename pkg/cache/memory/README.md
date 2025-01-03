# Memory Cache

## 目录
- [简介](#简介)
- [特性](#特性)
- [安装](#安装)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [API文档](#api文档)
- [实现原理](#实现原理)
- [性能报告](#性能报告)
- [测试覆盖](#测试覆盖)
- [最佳实践](#最佳实践)
- [注意事项](#注意事项)

## 简介
Memory Cache是一个高性能的内存缓存实现，专门为Go语言设计。它提供了线程安全的缓存操作，支持TTL过期，自动清理，并发控制等特性。该实现采用分片设计来减少锁竞争，并集成了完整的监控和日志系统。

## 特性
- **高性能**
  - 采用256个分片的设计，显著减少锁竞争
  - 使用sync.Map实现无锁读取
  - 支持对象池复用，减少GC压力
  
- **线程安全**
  - 所有操作都是并发安全的
  - 支持高并发读写
  - 原子操作保证数据一致性

- **TTL支持**
  - 支持为每个缓存项设置过期时间
  - 自动清理过期数据
  - 惰性删除与主动清理结合

- **监控集成**
  - 集成Prometheus监控
  - 支持操作计数
  - 支持延迟统计
  - 支持错误统计

- **日志跟踪**
  - 集成统一的日志系统
  - 支持Debug级别的详细日志
  - 支持异步日志写入

## 安装

```go
import "gobase/pkg/cache/memory"
```

## 快速开始

```go
// 创建缓存实例
config := memory.DefaultConfig()
cache, err := memory.NewCache(config, logger)
if err != nil {
    return err
}
defer cache.Stop()

// 设置缓存
err = cache.Set(ctx, "key", "value", time.Minute)

// 获取缓存
value, err := cache.Get(ctx, "key")

// 删除缓存
err = cache.Delete(ctx, "key")

// 清空缓存
err = cache.Clear(ctx)
```

## 配置说明

```go
type Config struct {
    // 最大条目数
    MaxEntries int

    // 清理间隔
    CleanupInterval time.Duration

    // 默认过期时间
    DefaultTTL time.Duration
}
```

默认配置：
- MaxEntries: 10000
- CleanupInterval: 1分钟
- DefaultTTL: 1小时

## API文档

### 核心接口

```go
type Cache interface {
    // 设置缓存
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    
    // 获取缓存
    Get(ctx context.Context, key string) (interface{}, error)
    
    // 删除缓存
    Delete(ctx context.Context, key string) error
    
    // 清空缓存
    Clear(ctx context.Context) error
    
    // 获取缓存级别
    GetLevel() cache.Level
    
    // 停止缓存服务
    Stop()
}
```

## 实现原理

### 分片设计
- 使用256个分片减少锁竞争
- 基于FNV-64哈希算法进行分片
- 每个分片使用sync.Map存储

### 数据结构
```go
type Cache struct {
    shards    []*cacheShard   // 分片数组
    numShards int            // 分片数量
    config    *Config        // 配置信息
    logger    types.Logger   // 日志记录器
    metrics   *metric.Counter // 监控指标
    stopCh    chan struct{}  // 停止信号
    count     int64         // 有效缓存项数量
}

type cacheShard struct {
    data sync.Map
}

type cacheItem struct {
    value      interface{}
    expiration time.Time
}
```

### 清理机制
1. 惰性删除：获取时检查过期
2. 定时清理：后台协程定期清理
3. 内存触发：内存使用超过80%时触发清理

## 性能报告
详细的性能测试报告请参考：[性能测试报告](tests/benchmark/README.md)

## 测试覆盖

### 单元测试
- 基础操作测试
- 配置验证测试
- 并发操作测试
- TTL功能测试

### 集成测试
- 完整流程测试
- 过期清理测试
- 容量限制测试
- 并发安全测试

### 压力测试
- 高并发场景测试
- 内存使用监控
- 延迟统计分析
- 错误率统计

## 最佳实践

1. 合理设置容量
```go
config.MaxEntries = runtime.NumCPU() * 10000
```

2. 调整清理间隔
```go
config.CleanupInterval = time.Minute * 5
```

3. 使用合适的TTL
```go
// 热点数据
cache.Set(ctx, "hot-key", value, time.Hour)

// 临时数据
cache.Set(ctx, "temp-key", value, time.Minute)
```

## 注意事项

1. 内存使用
- 注意设置合理的MaxEntries避免OOM
- 监控内存使用情况
- 及时清理过期数据

2. 并发控制
- 避免过多goroutine并发写入
- 合理使用批量操作
- 注意错误处理

3. 监控告警
- 配置合理的监控指标
- 设置适当的告警阈值
- 关注错误率和延迟指标