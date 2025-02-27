# GoBase 底层包

GoBase 是一个模块化的 Go 语言基础框架包,提供了构建企业级分布式应用所需的核心基础组件。每个组件都是独立的、可插拔的,并提供了完整的文档和测试。

## 目录
- [核心功能](#核心功能)
- [中间件](#中间件)
- [监控与追踪](#监控与追踪)
- [存储与缓存](#存储与缓存)
- [认证与安全](#认证与安全)
- [工具组件](#工具组件)

## 核心功能

### [错误处理](/pkg/errors/README.md)
统一的错误处理机制,支持错误链追踪、错误码管理、错误分类等特性。

### [日志系统](/pkg/logger/README.md)
结构化日志系统,支持多级别日志、ELK集成、异步写入、日志轮转等功能。

### [配置系统](/pkg/config/README.md)
灵活的配置管理系统,支持多格式配置文件、配置热更新、分布式配置中心等特性。

### [上下文](/pkg/context/README.md)
统一的上下文管理,支持元数据传递、超时控制、链路追踪等功能。

## 中间件

### [Recovery 中间件](/pkg/middleware/recovery/README.md)
用于捕获并处理 HTTP 请求处理过程中的 panic,确保服务稳定性。

### [Context 中间件](/pkg/middleware/context/README.md)
提供请求上下文管理,支持请求ID注入、元数据传递等功能。

### [JWT 中间件](/pkg/middleware/jwt/README.md)
JWT认证中间件,提供:
- 令牌验证和解析
- Context操作API
- 自定义Claims扩展
- 用户身份与权限管理
- 性能优异(0.156ms/请求)

### [Ratelimit 中间件](/pkg/middleware/ratelimit/README.md)
HTTP 请求限流中间件,支持多种限流算法和分布式限流。

### [Metrics 中间件](/pkg/middleware/metrics/README.md)
HTTP 请求指标收集中间件,用于收集请求相关的 Prometheus 指标。

### [Logger 中间件](/pkg/middleware/logger/README.md)
HTTP 请求日志中间件,支持结构化日志、采样策略、异步写入等特性。

## 监控与追踪

### [Prometheus 监控](/pkg/monitor/prometheus/README.md)
完整的监控指标收集方案,支持系统指标、业务指标、自定义指标等。

### [Grafana 监控](/pkg/monitor/grafana/README.md)
可视化监控方案,提供开箱即用的仪表盘模板和告警规则配置。

### [Jaeger 链路追踪](/pkg/trace/jaeger/README.md)
分布式链路追踪解决方案,支持请求链路追踪、性能分析、问题定位等。

## 存储与缓存

### [Redis 客户端](/pkg/client/redis/README.md)
Redis 客户端封装,支持连接池管理、集群模式、监控集成等特性。

### [Cache 缓存](/pkg/cache/README.md)
多级缓存解决方案,支持本地缓存、分布式缓存、缓存同步等功能。

### [Ratelimit 限流](/pkg/ratelimit/README.md)
通用限流组件,支持多种限流算法、分布式限流、监控集成等特性。

## 认证与安全

### [JWT 认证](/pkg/auth/jwt/README.md)
JWT 认证解决方案,支持令牌管理、黑名单机制、安全增强等功能。

## 工具组件

### [请求ID生成器](/pkg/utils/requestid/README.md)
全局唯一请求ID生成器,支持多种生成策略,适用于分布式场景。

## 使用要求
- Go 1.23.4+
- Redis 6.0+
- Jaeger 1.30+
- Prometheus 2.30+
- Grafana 8.0+

## 许可证
MIT License

## 开发计划

### 即将实现
1. JWT 中间件
   - JWT Token 验证
   - 权限控制
   - 刷新Token
   - 黑名单管理

2. Kafka 消息队列
   - 生产者/消费者封装
   - 消息持久化
   - 消息重试
   - 死信队列
   - 监控集成

3. Consul 服务发现
   - 服务注册与发现
   - 健康检查
   - 配置中心
   - KV存储
   - 监控集成

## 参与贡献

我非常欢迎社区贡献，无论是修复错误还是添加新功能！

### 如何贡献

1. Fork 本仓库
2. 创建您的特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交您的更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开一个 Pull Request

### 贡献指南

- 确保通过所有测试
- 更新相关文档
- 遵循现有的代码风格
- 添加必要的测试用例
- 在PR中详细描述您的更改

### 报告问题

如果您发现任何问题或有新功能建议，欢迎创建 Issue。请尽可能详细地描述问题或建议。
