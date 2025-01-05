# JWT Security Package

## 目录
- [概述](#概述)
- [主要功能](#主要功能)
  - [令牌策略管理](#令牌策略管理)
  - [密钥轮换](#密钥轮换)
  - [令牌验证](#令牌验证)
- [使用示例](#使用示例)
- [性能表现](#性能表现)
- [测试覆盖](#测试覆盖)

## 概述

JWT Security 包提供了一套完整的 JWT 令牌安全管理解决方案，包括令牌重用检测、密钥轮换和令牌验证等核心功能。该包集成了缓存、链路追踪和监控等基础设施，以提供企业级的安全保障。

## 主要功能

### 令牌策略管理

Policy 组件提供令牌重用检测和生命周期管理：

```go
type Policy struct {
    Cache              cache.Cache
    MaxTokenAge        time.Duration
    TokenReuseInterval time.Duration
}
```

主要功能：
- 令牌重用检测
- 可配置的令牌最大生命周期
- 可配置的令牌重用间隔
- 基于 Redis 的分布式令牌追踪

### 密钥轮换

Rotation 组件提供自动密钥轮换机制：

```go
type KeyRotator struct {
    keyManager crypto.KeyManager
    interval   time.Duration
    logger     types.Logger
    metrics    *metric.Counter
}
```

主要特性：
- 自动定期轮换密钥
- 支持优雅启动和停止
- 完整的日志记录
- Prometheus 指标监控

### 令牌验证

Validator 组件提供全面的令牌验证：

```go
type TokenValidator struct {
    maxAge time.Duration
}
```

验证内容：
- 令牌过期检查
- IP 地址验证
- 设备信息验证
- 令牌类型检查

## 使用示例

### 策略验证示例

```go
policy := security.NewPolicy(
    WithCache(redisCache),
    WithMaxAge(24 * time.Hour),
    WithReuseInterval(5 * time.Minute),
)

err := policy.ValidateTokenReuse(ctx, tokenID)
if err != nil {
    // 处理令牌重用错误
}
```

### 密钥轮换示例

```go
rotator := security.NewKeyRotator(
    keyManager,
    time.Hour,
    logger,
    metrics,
)

err := rotator.Start(ctx)
if err != nil {
    // 处理启动错误
}
defer rotator.Stop()
```

## 性能表现

完整的性能测试报告请参考：[性能测试报告](tests/benchmark/README.md)

主要性能指标：
- 令牌验证: 每秒可处理超过10亿请求
- 令牌重用检测: 平均延迟 < 1ms (Redis 存储)
- 密钥轮换: 无性能影响

## 测试覆盖

### 单元测试
- [Policy 测试](tests/unit/policy_test.go)
- [Rotation 测试](tests/unit/rotation_test.go)
- [Validator 测试](tests/unit/validator_test.go)

### 集成测试
- [Security 集成测试](tests/integration/security_test.go)

### 压力测试
- [Security 压力测试](tests/stress/security_stress_test.go)

### 基准测试
- [Security 基准测试](tests/benchmark/security_bench_test.go)

所有测试用例覆盖了核心功能、边界条件和错误场景，确保了包的稳定性和可靠性。
