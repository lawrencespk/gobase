# JWT Config

## 目录
- [简介](#简介)
- [配置结构](#配置结构)
- [默认配置](#默认配置)
- [配置选项](#配置选项)
- [使用示例](#使用示例)
- [测试覆盖](#测试覆盖)

## 简介
JWT Config 包提供了一个灵活的配置系统，用于管理 JWT (JSON Web Token) 的各项设置。支持包括签名方法、密钥管理、令牌过期时间、黑名单、会话管理等多个方面的配置。

## 配置结构
主要配置结构包含以下核心部分：

```go
type Config struct {
    // 签名方法配置
    SigningMethod jwt.SigningMethod 

    // 密钥配置
    SecretKey  string
    PublicKey  string 
    PrivateKey string

    // Token配置
    AccessTokenExpiration  time.Duration
    RefreshTokenExpiration time.Duration

    // 黑名单配置
    BlacklistEnabled bool
    BlacklistType    string // memory or redis
    Redis           *RedisConfig

    // 安全配置
    EnableRotation   bool
    RotationInterval time.Duration

    // 会话配置
    EnableSession     bool
    MaxActiveSessions int

    // 绑定配置
    EnableIPBinding     bool
    EnableDeviceBinding bool

    // 可观测性配置
    EnableMetrics bool
    EnableTracing bool
}
```

## 默认配置
通过 `DefaultConfig()` 函数获取默认配置：

```go
cfg := config.DefaultConfig()
// 默认值:
// - SigningMethod: HS256
// - AccessTokenExpiration: 2h
// - RefreshTokenExpiration: 24h
// - BlacklistEnabled: true
// - BlacklistType: "memory"
// - EnableRotation: false
// - RotationInterval: 24h
// - EnableSession: false
// - MaxActiveSessions: 1
// - EnableMetrics: true
// - EnableTracing: true
```

## 配置选项
提供了一系列配置选项函数用于自定义配置：

```go
// 设置签名方法
WithSigningMethod(method jwt.SigningMethod)

// 设置密钥
WithSecretKey(key string)
WithKeyPair(publicKey, privateKey string)

// 设置令牌过期时间
WithAccessTokenExpiration(d time.Duration)
WithRefreshTokenExpiration(d time.Duration)

// 配置黑名单
WithBlacklist(enabled bool, typ string)
WithRedis(addr, password string, db int)

// 配置Token轮换
WithRotation(enabled bool, interval time.Duration)

// 配置会话管理
WithSession(enabled bool, maxActive int)

// 配置绑定选项
WithBinding(enableIP, enableDevice bool)

// 配置可观测性
WithObservability(enableMetrics, enableTracing bool)
```

## 使用示例

### 基础配置
```go
cfg := config.DefaultConfig()
config.WithSigningMethod(jwt.RS256)(cfg)
config.WithKeyPair("public-key", "private-key")(cfg)
```

### Redis黑名单配置
```go
cfg := config.DefaultConfig()
config.WithBlacklist(true, "redis")(cfg)
config.WithRedis("localhost:6379", "", 0)(cfg)
```

## 测试覆盖

### 单元测试
- 测试文件：`tests/unit/config_test.go`
- 覆盖内容：
  - 默认配置验证
  - 配置选项功能验证
  - 各配置项独立测试

### 集成测试
- 测试文件：`tests/integration/config_test.go`
- 覆盖内容：
  - Redis黑名单配置集成测试
  - 配置选项组合测试
