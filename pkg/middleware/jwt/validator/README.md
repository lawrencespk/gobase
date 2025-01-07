# JWT Token Validator

JWT Token验证器，提供了Claims验证和黑名单验证等功能。

## 目录
- [功能特性](#功能特性)
- [接口定义](#接口定义)
- [验证器实现](#验证器实现)
- [使用示例](#使用示例)
- [性能报告](#性能报告)
- [测试覆盖](#测试覆盖)

## 功能特性

- Claims 验证
  - Token 过期验证（可配置）
  - Token 类型验证（可配置）
  - 支持自定义验证规则
- 黑名单验证
  - Redis 存储实现
  - 支持 Token 吊销
  - 自动过期清理
- 链式验证
  - 支持多个验证器串联
  - 验证器按序执行
  - 任一验证失败即终止

## 接口定义

```go
// TokenValidator 定义了Token验证器接口
type TokenValidator interface {
    Validate(c *gin.Context, claims jwt.Claims) error
}

// ValidatorFunc 函数类型实现了TokenValidator接口
type ValidatorFunc func(c *gin.Context, claims jwt.Claims) error
```

## 验证器实现

### Claims 验证器

```go
// 创建Claims验证器
validator := validator.NewClaimsValidator(
    validator.WithExpiryValidation(true),
    validator.WithTokenTypeValidation(true),
    validator.WithAllowedTokenTypes(jwt.AccessToken),
)
```

### 黑名单验证器

```go
// 创建黑名单验证器
store, _ := blacklist.NewRedisStore(redisClient, blacklist.DefaultOptions())
bl := blacklist.NewStoreAdapter(store)
validator := validator.NewBlacklistValidator(bl)
```

### 链式验证器

```go
// 创建链式验证器
chain := validator.ChainValidator([]TokenValidator{
    claimsValidator,
    blacklistValidator,
})
```

## 使用示例

```go
func JWTMiddleware() gin.HandlerFunc {
    // 创建验证器
    validator := validator.ChainValidator([]TokenValidator{
        validator.NewClaimsValidator(),
        validator.NewBlacklistValidator(blacklist),
    })

    return func(c *gin.Context) {
        // 解析token获取claims
        claims := // ... 解析token获取claims ...

        // 验证token
        if err := validator.Validate(c, claims); err != nil {
            c.AbortWithError(http.StatusUnauthorized, err)
            return
        }
        
        c.Next()
    }
}
```

## 性能报告

详细的性能测试报告请参考：
- [基准测试](tests/benchmark/README.md)
- [压力测试](tests/stress/blacklist_stress_test.go)

### 关键指标

| 场景 | 操作延迟 | 内存分配 | 内存分配次数 | 每秒处理能力 |
|------|----------|----------|--------------|--------------|
| 有效Token验证 | 0.31ms | 12.3KB | 205次 | ~3,300次 |
| 黑名单Token验证 | 0.29ms | 3.4KB | 49次 | ~4,100次 |

## 测试覆盖

- [单元测试](tests/unit/)
  - [Claims验证测试](tests/unit/claims_test.go)
  - [接口实现测试](tests/unit/interface_test.go)
- [集成测试](tests/integration/)
  - [黑名单验证测试](tests/integration/blacklist_test.go)
- [基准测试](tests/benchmark/)
  - [性能基准测试](tests/benchmark/blacklist_bench_test.go)
- [压力测试](tests/stress/)
  - [并发压力测试](tests/stress/blacklist_stress_test.go)
