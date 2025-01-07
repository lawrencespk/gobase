# JWT Context

JWT Context 包提供了在 Gin 框架中处理 JWT 上下文信息的功能。

## 目录
- [功能特性](#功能特性)
- [使用方法](#使用方法)
- [性能报告](#性能报告)
- [测试覆盖](#测试覆盖)
- [最佳实践](#最佳实践)

## 功能特性

### 上下文操作
- Claims 管理
  - 设置/获取 JWT Claims
  - 支持自定义 Claims 结构
- Token 管理
  - 设置/获取 Token 字符串
  - 设置/获取 Token 类型
- 用户信息
  - 设置/获取用户 ID
  - 设置/获取用户名
  - 设置/获取用户角色
  - 设置/获取用户权限
- 设备信息
  - 设置/获取设备 ID
  - 设置/获取 IP 地址

### 中间件集成
- 完整的 Gin 中间件支持
- 中间件链上下文传递
- 错误处理机制

## 使用方法

### 基础用法
```go
// 设置 JWT 上下文
ctx = jwtContext.WithJWTContext(ctx, claims, token)

// 获取用户信息
userID, err := jwtContext.GetUserID(ctx)
userName, err := jwtContext.GetUserName(ctx)
```

### 中间件使用
```go
func JWTMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx := jwtContext.WithJWTContext(c.Request.Context(), claims, token)
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}
```

### 错误处理
```go
userID, err := jwtContext.GetUserID(ctx)
if err != nil {
    // 处理错误情况
}
```

## 性能报告

详细的性能测试报告请参考：[性能测试报告](tests/benchmark/README.md)

### 关键指标
- WithJWTContext: 426.4 ns/op, 576 B/op
- GetAllValues: 136.9 ns/op, 0 B/op
- MiddlewareChain: 592.8 ns/op, 897 B/op

## 测试覆盖

### 单元测试
- [context_test.go](tests/unit/context_test.go)
  - 所有上下文操作的测试
  - 错误处理测试
  - 边界条件测试

### 集成测试
- [context_test.go](tests/integration/context_test.go)
  - 中间件链测试
  - 完整请求流程测试
  - 上下文传递测试

### 压力测试
- [context_stress_test.go](tests/stress/context_stress_test.go)
  - 高并发读写测试
  - 上下文值并发修改测试
  - 超时和错误处理测试

## 最佳实践

1. 上下文使用
   - 避免存储大量数据
   - 及时处理错误返回
   - 使用类型安全的 getter/setter

2. 中间件链
   - 合理安排中间件顺序
   - 避免重复设置上下文值
   - 正确处理上下文传递

3. 错误处理
   - 始终检查错误返回
   - 使用统一的错误处理方式
   - 避免吞掉错误
