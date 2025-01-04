# JWT Binding Package

## 简介
JWT Binding 包提供了一个用于 JWT token 绑定验证的解决方案，支持设备绑定和 IP 绑定两种验证方式。该包主要用于增强 JWT 认证的安全性，防止 token 被盗用。

## 主要特性

### 1. 设备绑定验证
- 支持设备 ID 和设备指纹的双重验证
- 完整的设备信息记录(ID、类型、名称、操作系统、浏览器等)
- 自动的设备信息合法性校验

### 2. IP 绑定验证
- 支持 IP 地址绑定验证
- IP 变化时的自动警告日志
- 灵活的 IP 匹配规则

### 3. 存储实现
- 基于 Redis 的高性能存储实现
- 支持自定义过期时间
- 自动的数据清理机制

### 4. 监控指标
- 完整的 Prometheus 指标收集
- 支持绑定成功/失败统计
- 支持操作延迟统计
- 支持错误计数统计

### 5. 日志追踪
- 集成 Jaeger 分布式追踪
- 详细的操作日志记录
- 异常情况自动警告

## 快速开始

### 1. 创建存储实例

```go
redisClient := redis.NewClient(options)
store, err := binding.NewRedisStore(redisClient)
if err != nil {
    // 处理错误
}
```

### 2. 创建验证器

```go
// 创建设备验证器
deviceValidator, err := binding.NewDeviceValidator(store)
if err != nil {
    // 处理错误
}
// 创建 IP 验证器
ipValidator, err := binding.NewIPValidator(store)
if err != nil {
    // 处理错误
}
```

### 3. 验证绑定

```go
// 验证设备绑定
err = deviceValidator.ValidateDevice(ctx, claims, deviceInfo)
if err != nil {
    // 处理验证失败
}
// 验证 IP 绑定
err = ipValidator.ValidateIP(ctx, claims, currentIP)
if err != nil {
    // 处理验证失败
}
```

## 性能表现
该包经过严格的性能测试，包括基准测试、压力测试和并发测试。详细的性能测试报告请查看 [性能测试报告](tests/benchmark/README.md)。

主要性能指标：
- 设备验证: ~32,833 ops/s
- IP 验证: ~35,338 ops/s
- 设备绑定保存: ~24,515 ops/s
- IP 绑定保存: ~30,857 ops/s

## 错误处理
包内使用统一的错误处理机制，主要错误类型包括：
- `BindingInvalidError`: 绑定信息无效
- `BindingMismatchError`: 绑定信息不匹配
- `StoreError`: 存储操作错误

## 测试覆盖
该包包含完整的测试套件：
- 单元测试：覆盖所有核心功能
- 集成测试：验证与 Redis 的交互
- 压力测试：验证高并发场景下的稳定性
- 性能测试：验证性能表现

## 注意事项
1. Redis 存储默认使用 24 小时过期时间
2. 建议在生产环境中配置适当的监控告警
3. 设备指纹和设备 ID 都是必需的验证字段
4. IP 绑定验证支持 IPv4 和 IPv6