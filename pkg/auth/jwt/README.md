# JWT Package

## 目录
- [简介](#简介)
- [功能特性](#功能特性)
- [包结构](#包结构)
- [相关文档](#相关文档)

## 简介
JWT 包提供了一个完整的 JSON Web Token 解决方案，包含令牌的生成、验证、黑名单、绑定验证等功能。该包设计用于分布式系统中的身份认证和授权管理。

## 功能特性

### 1. 令牌绑定 (binding)
- 设备绑定：支持设备 ID 和设备指纹的双重验证
- IP 绑定：支持 IP 地址绑定验证
- [详细文档](binding/README.md)

### 2. 令牌黑名单 (blacklist)
- 支持内存和 Redis 存储
- 自动清理过期令牌
- 完整的监控指标
- [详细文档](blacklist/README.md)

### 3. 配置管理 (config)
- 灵活的配置选项
- 支持动态配置更新
- [详细文档](config/README.md)

### 4. 加密实现 (crypto)
- 支持多种签名算法
- 安全的密钥管理
- [详细文档](crypto/README.md)

### 5. 事件管理 (events)
- 基于 Redis 的事件发布/订阅
- 支持令牌相关事件通知
- [详细文档](events/README.md)

### 6. 安全增强 (security)
- 令牌重用检测
- 密钥轮换机制
- 完整的令牌验证
- [详细文档](security/README.md)

### 7. 会话管理 (session)
- 分布式会话支持
- 会话生命周期管理
- [详细文档](session/README.md)

### 8. 令牌存储 (store)
- 支持多种存储后端
- 高性能的缓存机制
- [详细文档](store/README.md)

### 9. 同步机制 (sync)
- 分布式锁实现
- 并发操作保护
- [详细文档](sync/README.md)


### 10. 指标 (metrics)
- 支持Prometheus监控
- 支持OpenTelemetry追踪
- [详细文档](metrics/README.md)

## 包结构

```
jwt/
├── binding/ # 令牌绑定验证
├── blacklist/ # 令牌黑名单管理
├── config/ # JWT 配置管理
├── crypto/ # 加密和签名实现
├── events/ # 事件通知系统
├── security/ # 安全增强功能
├── session/ # 会话管理
├── store/ # 令牌存储实现
└── sync/ # 同步机制
```

## 相关文档
每个子包都包含详细的文档，包括功能说明、使用示例和性能报告。请参考各子包中的 README.md 文件获取更多信息。