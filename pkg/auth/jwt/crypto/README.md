# JWT Crypto Package

## 目录
- [简介](#简介)
- [功能特性](#功能特性)
- [接口设计](#接口设计)
- [支持的算法](#支持的算法)
- [使用示例](#使用示例)
- [密钥管理](#密钥管理)
- [性能测试](#性能测试)
- [测试覆盖](#测试覆盖)

## 简介

crypto 包提供了 JWT (JSON Web Token) 签名和验证的核心加密功能实现。该包支持多种签名算法,并提供了安全的密钥管理机制。

## 功能特性

- 支持多种签名算法 (HMAC-SHA系列, RSA系列)
- 线程安全的密钥管理
- 支持密钥轮换(RSA)
- 完整的日志记录
- 统一的错误处理
- 高性能实现

## 接口设计

### KeyProvider 接口
```go
type KeyProvider interface {
    GetSigningKey() (interface{}, error)
    GetVerificationKey() (interface{}, error)
    RotateKeys() error
}
```

### Algorithm 接口
```go
type Algorithm interface {
    Sign(data []byte, key interface{}) ([]byte, error)
    Verify(data []byte, signature []byte, key interface{}) error
    Name() jwt.SigningMethod
}
```

## 支持的算法

### HMAC-SHA 系列
- HS256 (HMAC using SHA-256)
- HS384 (HMAC using SHA-384)
- HS512 (HMAC using SHA-512)

### RSA 系列
- RS256 (RSA using SHA-256)
- RS384 (RSA using SHA-384)
- RS512 (RSA using SHA-512)

## 使用示例

### 创建密钥管理器
```go
logger := yourLogger // 实现 types.Logger 接口
km, err := crypto.NewKeyManager(jwt.RS256, logger)
if err != nil {
    // 处理错误
}
```

### 初始化密钥
```go
config := &jwt.KeyConfig{
    PrivateKey: "-----BEGIN PRIVATE KEY-----\n...",
}
err = km.InitializeKeys(ctx, config)
if err != nil {
    // 处理错误
}
```

### 签名和验证
```go
// 获取签名密钥
signingKey, err := km.GetSigningKey()
if err != nil {
    // 处理错误
}

// 签名数据
signature, err := algorithm.Sign(data, signingKey)
if err != nil {
    // 处理错误
}

// 获取验证密钥
verificationKey, err := km.GetVerificationKey()
if err != nil {
    // 处理错误
}

// 验证签名
err = algorithm.Verify(data, signature, verificationKey)
if err != nil {
    // 处理错误
}
```

## 密钥管理

### RSA 密钥轮换
```go
// 仅支持 RSA 密钥轮换
if err := km.RotateKeys(ctx); err != nil {
    // 处理错误
}
```

### 线程安全
KeyManager 实现了完整的并发安全机制:
- 使用 sync.RWMutex 保护密钥访问
- 支持并发的密钥获取操作
- 安全的密钥轮换机制

## 性能测试

详细的性能测试结果请参考: [性能测试报告](tests/benchmark/README.md)

主要性能指标:
- HMAC-SHA256: ~217万次/秒 (签名), ~208万次/秒 (验证)
  - 内存分配: 536B/操作
  - 分配次数: 7次/操作
- RSA-SHA256: 
  - 签名: ~1,576次/秒 (634.64 µs/操作)
    - 内存分配: 1056B/操作
    - 分配次数: 7次/操作
  - 验证: ~50,403次/秒 (19.84 µs/操作)
    - 内存分配: 1264B/操作
    - 分配次数: 9次/操作

性能测试环境及完整分析请参考:
- [详细性能测试报告](tests/benchmark/README.md)
## 测试覆盖

### 单元测试
- [密钥管理测试](tests/unit/keys_test.go)
- [算法实现测试](tests/unit/algorithm_test.go)

### 集成测试
- [端到端加密测试](tests/integration/crypto_test.go)

### 压力测试
- [并发性能测试](tests/stress/crypto_stress_test.go)

### 基准测试
- [算法性能测试](tests/benchmark/crypto_bench_test.go)
