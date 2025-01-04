# JWT Events

JWT Events 包提供了基于 Redis 的事件发布/订阅机制,用于处理 JWT 相关的事件通知,如令牌吊销、过期、密钥轮换等场景。

## 目录结构

```
pkg/auth/jwt/events/
├── options.go        # 配置选项定义
├── publisher.go      # 事件发布器实现
├── subscriber.go     # 事件订阅器实现  
├── types.go          # 事件类型与结构定义
├── types/
│   └── event.go      # 事件类型常量定义
└── tests/
    ├── unit/         # 单元测试
    ├── integration/  # 集成测试  
    ├── stress/       # 压力测试
    ├── benchmark/    # 基准测试
    └── testutils/    # 测试工具
```

## 功能特性

### 事件类型支持

- TokenRevoked: 令牌吊销事件
- TokenExpired: 令牌过期事件  
- KeyRotated: 密钥轮换事件
- SessionCreated: 会话创建事件
- SessionDestroyed: 会话销毁事件

### 发布器 (Publisher)

```go
// 创建发布器
publisher, err := events.NewPublisher(redisClient, logger, 
    events.WithChannel("jwt:events"),
    events.WithPublisherMetrics(metrics),
)

// 发布事件
err = publisher.Publish(ctx, events.TokenRevoked, payload)
```

主要特性:
- 支持自定义 Redis 通道
- 集成 Prometheus 指标监控
- 事件 ID 自动生成
- 结构化日志记录
- 分布式追踪支持

### 订阅器 (Subscriber)

```go
// 创建订阅器
subscriber := events.NewSubscriber(redisClient, logger,
    events.WithSubscriberMetrics(metrics),
)

// 注册事件处理器
subscriber.RegisterHandler(events.TokenRevoked, func(event *Event) error {
    // 处理事件
    return nil
})

// 启动订阅
err = subscriber.Subscribe(ctx)
```

主要特性:
- 支持多个事件处理器注册
- 异步事件处理
- 优雅关闭支持
- 结构化日志记录
- 指标监控集成

## 性能表现

完整的性能测试报告请参考: [性能测试报告](tests/benchmark/README.md)

主要性能指标:
- Publisher: ~2,367 ops/s
- Subscriber: 近零延迟的事件处理
- 内存分配: 发布端 ~7KB/op

## 可观测性

### 日志

- 使用结构化日志
- 支持多个日志级别
- 包含详细的上下文信息
- 异常情况完整记录

### 指标

支持以下 Prometheus 指标:
- 事件发布计数
- 事件处理计数
- 处理延迟统计
- 错误计数

### 追踪

集成 Jaeger 分布式追踪:
- 事件发布链路
- 事件处理链路
- 关键操作耗时

## 测试覆盖

- 单元测试: 核心逻辑验证
- 集成测试: Redis 交互验证
- 压力测试: 高并发场景验证
- 基准测试: 性能指标验证

## 使用示例

```go
// 初始化发布器
publisher, err := events.NewPublisher(redisClient, logger)
if err != nil {
    return err
}

// 发布令牌吊销事件
err = publisher.Publish(ctx, events.TokenRevoked, map[string]interface{}{
    "token_id": "xxx",
    "user_id": "xxx",
})

// 初始化订阅器
subscriber := events.NewSubscriber(redisClient, logger)

// 注册处理器
subscriber.RegisterHandler(events.TokenRevoked, func(event *Event) error {
    // 处理令牌吊销逻辑
    return nil
})

// 启动订阅
err = subscriber.Subscribe(ctx)
```

## 最佳实践

1. 错误处理
   - 始终检查返回的错误
   - 使用统一的错误处理机制
   - 合理设置重试策略

2. 性能优化
   - 合理设置 Redis 连接池
   - 避免过大的事件负载
   - 注意处理器的执行效率

3. 可观测性
   - 合理配置日志级别
   - 监控关键指标
   - 追踪重要操作

4. 优雅关闭
   - 使用 context 控制生命周期
   - 等待处理中的事件完成
   - 清理资源

## 依赖

- Redis: 用于事件发布订阅
- Prometheus: 指标收集
- Jaeger: 分布式追踪
- Logrus: 结构化日志
