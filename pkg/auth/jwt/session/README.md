# JWT Session 管理包

## 目录
- [概述](#概述)
- [特性](#特性)
- [组件](#组件)
- [使用方法](#使用方法)
- [存储实现](#存储实现)
- [测试覆盖](#测试覆盖)
- [性能报告](#性能报告)
- [注意事项](#注意事项)

## 概述
JWT Session 包提供了一个完整的 JWT 会话管理解决方案,支持会话的创建、存储、验证和清理等核心功能。该包设计用于处理分布式系统中的用户会话管理。

## 特性
- 完整的会话生命周期管理
- 支持 Redis 分布式存储
- 内置监控指标收集
- 分布式追踪支持(Jaeger)
- 完整的错误处理机制
- 高性能实现

## 组件

### 核心接口

```go
type Store interface {
    // 获取会话
    Get(ctx context.Context, key string) (*Session, error)
    
    // 设置会话
    Set(ctx context.Context, key string, session *Session) error
    
    // 删除会话
    Delete(ctx context.Context, key string) error
    
    // 关闭存储连接
    Close(ctx context.Context) error
}
```

### 会话数据结构

```go
type Session struct {
    // 用户ID
    UserID    string                 `json:"user_id"`
    
    // 令牌ID
    TokenID   string                 `json:"token_id"` 
    
    // 创建时间
    CreatedAt time.Time             `json:"created_at"`
    
    // 更新时间
    UpdatedAt time.Time             `json:"updated_at"`
    
    // 元数据
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
```

## 使用方法

### 初始化 Redis 存储

```go
// Redis配置选项
redisOpts := &session.RedisOptions{
    Addr:     "localhost:6379",
    Password: "password",
    DB:       0,
}

// 创建存储实例
store, err := session.NewRedisStore(redisClient)
if err != nil {
    return err
}
defer store.Close()
```

### 会话管理

```go
// 创建会话
sess := &Session{
    UserID:  "user-1",
    TokenID: "token-1",
}
err := store.Set(ctx, sess.TokenID, sess)

// 获取会话
sess, err := store.Get(ctx, tokenID)

// 删除会话
err := store.Delete(ctx, tokenID)
```

## 存储实现

### Redis 存储
- 支持会话数据的持久化
- 自动过期清理机制
- 原子操作保证
- 支持分布式部署

## 测试覆盖

### 单元测试
- [接口测试](tests/unit/interface_test.go)
- [管理器测试](tests/unit/manager_test.go)
- [Redis存储测试](tests/unit/redis_test.go)
- [Store测试](tests/unit/store_test.go)
- [类型测试](tests/unit/types_test.go)

### 集成测试
- [会话集成测试](tests/integration/session_test.go)
- [测试环境配置](tests/integration/setup_test.go)

### 性能测试
- [压力测试](tests/stress/session_stress_test.go)
- [基准测试](tests/benchmark/session_bench_test.go)
- [详细性能报告](tests/benchmark/README.md)

## 性能报告

详细的性能测试结果请参考:
- [基准测试报告](tests/benchmark/README.md)
- [压力测试报告](tests/stress/session_stress_test.go)

主要性能指标:
- 会话创建: ~40.791 µs/op
- 会话获取: ~28.298 µs/op
- 会话删除: ~32.408 µs/op

## 注意事项

1. Redis 配置
   - 建议配置适当的内存上限
   - 建议开启持久化机制
   - 建议配置合理的过期时间

2. 监控
   - 已集成 Prometheus 指标收集
   - 支持 Jaeger 分布式追踪
   - 建议配置适当的告警阈值

3. 错误处理
   - 所有错误都经过标准化处理
   - 支持错误代码和详细错误信息
   - 建议妥善处理所有返回的错误

4. 性能优化
   - 使用连接池管理 Redis 连接
   - 批量操作时使用 Pipeline
   - 合理使用上下文控制超时