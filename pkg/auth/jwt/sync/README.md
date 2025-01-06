# 同步JWT

## 目录
- [简介](#简介)
- [功能特性](#功能特性)
- [接口定义](#接口定义)
- [使用示例](#使用示例)
- [性能测试](#性能测试)
- [测试覆盖](#测试覆盖)
- [最佳实践](#最佳实践)

## 简介
同步JWT包提供了基于Redis的分布式锁实现，用于确保JWT操作的并发安全性。该实现支持基本的锁操作，包括获取锁、尝试获取锁和释放锁等功能。

## 功能特性
- 基于Redis的分布式锁实现
- 支持超时自动释放(30秒)
- 支持错误重试
- 支持链路追踪
- 支持性能监控
- 支持并发操作

## 接口定义
```go
type Locker interface {
    // Lock 获取锁
    Lock(ctx context.Context) error

    // TryLock 尝试获取锁
    TryLock(ctx context.Context) (bool, error)

    // Unlock 释放锁
    Unlock(ctx context.Context) error
}
```

## 使用示例
```go
import (
    "context"
    "gobase/pkg/auth/jwt/sync"
    "gobase/pkg/client/redis"
)

func Example() {
    // 创建Redis客户端
    client, _ := redis.NewClient(
        redis.WithAddress("localhost:6379"),
    )
    
    // 创建分布式锁
    locker := sync.NewLocker(client, "my-key")
    
    // 获取锁
    if err := locker.Lock(context.Background()); err != nil {
        // 处理错误
        return
    }
    
    // 使用完后释放锁
    defer locker.Unlock(context.Background())
    
    // 执行需要同步的操作
    // ...
}
```

## 性能测试
详细的性能测试[基准测试文档](pkg/auth/jwt/sync/tests/benchmark/README.md)

主要性能指标：
- 单个锁操作延迟: < 25μs
- 吞吐量: 40,000+ QPS
- 内存占用: ~400 bytes/op

### 压力测试结果
在100个并发工作线程下，每个线程执行1000次操作的测试中，系统表现稳定。详细测试用例可以查看[压力测试代码](pkg/auth/jwt/sync/tests/stress/interface_stress_test.go)。

## 测试覆盖
包含以下类型的测试：
1. 单元测试
   - 基本锁操作测试
   - 错误处理测试
   - Mock测试
   
2. 集成测试
   - Redis连接测试
   - 并发操作测试
   - 超时处理测试
   
3. 压力测试
   - 高并发测试
   - 持续运行测试
   
4. 基准测试
   - 操作延迟测试
   - 内存分配测试
   - 并发性能测试

## 最佳实践
1. 锁的使用
   - 总是在defer中调用Unlock
   - 设置合适的上下文超时
   - 使用TryLock处理竞争场景

2. 错误处理
   - 检查Lock的返回错误
   - 处理Unlock可能的失败
   - 使用context控制超时

3. 性能优化
   - 最小化锁持有时间
   - 合理设置超时时间
   - 避免长时间持有锁