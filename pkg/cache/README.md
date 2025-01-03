# Cache 缓存包

## 目录结构

pkg/cache/
├── README.md # 本文档
├── interface.go # 缓存接口定义
├── memory/ # 内存缓存实现
│ └── README.md # 内存缓存文档
├── multilevel/ # 多级缓存实现
│ └── README.md # 多级缓存文档
├── redis/ # Redis缓存实现
│ ├── README.md # Redis缓存文档
│ └── ratelimit/ # Redis限流实现
│ └── README.md # Redis限流文档

## 文档索引

### 1. [内存缓存(Memory Cache)](memory/README.md)
- 高性能的本地内存缓存实现
- 特性:
  - 线程安全
  - TTL过期机制
  - 自动清理
  - 分片设计减少锁竞争
  - 完整的监控和日志系统

### 2. [多级缓存(Multilevel Cache)](multilevel/README.md)
- 结合本地内存和Redis的分布式缓存解决方案
- 特性:
  - 两级缓存架构(L1本地+L2分布式)
  - 自动缓存同步
  - 缓存预热
  - 监控指标
  - 错误处理
  - 性能优化

### 3. [Redis缓存(Redis Cache)](redis/README.md)
- 基于Redis的分布式缓存实现
- 特性:
  - 分布式支持
  - 集群支持
  - 连接池管理
  - 错误重试
  - 监控指标

### 4. [Redis限流(Redis RateLimit)](redis/ratelimit/README.md)
- 基于Redis的分布式限流实现
- 支持算法:
  - 滑动窗口
  - 令牌桶
  - 漏桶
  - 固定窗口
- 特性:
  - 分布式限流
  - 监控指标
  - 链路追踪
  - 高性能实现

## 性能报告

各个实现的性能测试报告可以在对应的 tests/benchmark 目录下找到:

- [内存缓存性能测试](memory/tests/benchmark/README.md)
- [多级缓存性能测试](multilevel/tests/benchmark/README.md)

## 最佳实践

1. 缓存选型:
- 单机场景: 使用内存缓存
- 分布式场景: 使用多级缓存或Redis缓存
- 限流需求: 使用Redis限流

2. 监控建议:
- 配置缓存监控指标
- 设置合理的告警阈值
- 关注错误率和延迟指标
- 监控内存使用情况

3. 性能优化:
- 合理设置缓存容量
- 优化热点数据访问
- 使用批量操作
- 配置合适的过期时间
- 启用预热机制

4. 错误处理:
- 做好降级处理
- 使用重试机制
- 记录详细日志
- 设置超时控制

## 注意事项

1. 内存管理:
- 合理设置缓存大小
- 监控内存使用
- 及时清理过期数据

2. 并发控制:
- 注意锁的粒度
- 合理使用批量操作
- 控制并发数量

3. 分布式场景:
- 注意缓存一致性
- 合理设置过期时间
- 做好容错处理

4. 监控告警:
- 配置核心指标监控
- 设置合理的告警阈值
- 关注性能变化趋势