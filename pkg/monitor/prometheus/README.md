# Prometheus 监控模块

## 目录
- [简介](#简介)
- [架构设计](#架构设计)
- [目录结构](#目录结构)
- [依赖项](#依赖项)
- [安装](#安装)
- [配置](#配置)
- [使用指南](#使用指南)
- [API文档](#api文档)
- [指标类型](#指标类型)
- [内置收集器](#内置收集器)
- [测试](#测试)
- [性能优化](#性能优化)
- [错误处理](#错误处理)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)

## 简介
这是一个基于 Prometheus 的监控模块，提供了完整的指标收集、导出和监控功能。该模块设计简洁、易用、易扩展，适用于分布式系统的监控需求。

### 主要特性
- 支持三种基本指标类型：Counter、Gauge、Histogram
- 内置六种核心收集器：HTTP、Redis、System、Runtime、Business、Resource
- 支持指标采样和自定义标签
- 支持优雅启动和关闭
- 完整的单元测试和集成测试覆盖
- 支持容器化部署

## 架构设计

### 核心组件
1. Exporter: 指标导出器
2. Collector: 指标收集器
3. Metric: 指标类型实现
4. Sampler: 采样器
5. Config: 配置管理

### 数据流
```
[业务代码] -> [Collector] -> [Metric] -> [Exporter] -> [Prometheus Server]
```

## 目录结构
```
pkg/monitor/prometheus/
├── collector/    # 指标收集器
├── config/       # 配置管理
├── exporter/     # 指标导出器
├── metric/       # 指标类型
├── sampler/      # 采样器
└── tests/        # 测试用例
    ├── integration/  # 集成测试
    ├── unit/        # 单元测试
    └── testutils/   # 测试工具
```

## 依赖项
- Go 1.18+
- prometheus/client_golang v1.x
- shirou/gopsutil v3.x
- testcontainers/testcontainers-go (用于测试)

## 安装
```bash
go get github.com/lawrencespk/gobase/pkg/monitor/prometheus
```

## 配置
```yaml
prometheus:
  enabled: true
  port: 9090
  path: "/metrics"
  namespace: "myapp"
  labels:
    env: "prod"
    region: "us-east"
  collectors:
    - "http"
    - "redis"
    - "system"
    - "runtime"
    - "business"
    - "resource"
  sampling:
    enabled: true
    rate: 0.1
```

## 使用指南

### 基础使用场景

#### 1. HTTP 服务监控
```go
import (
    "gobase/pkg/monitor/prometheus/collector"
    "gobase/pkg/monitor/prometheus/exporter"
)

func main() {
    // 创建 HTTP 收集器
    httpCollector := collector.NewHTTPCollector("my_service")
    
    // 记录请求
    httpCollector.ObserveRequest("GET", "/api/v1/users", 200, 100, 1024, 2048)
    
    // 监控慢请求
    httpCollector.ObserveSlowRequest("GET", "/api/v1/users", 2.5)
}
```

#### 2. Redis 性能监控
```go
// 监控 Redis 连接池和操作
redisCollector := collector.NewRedisCollector("my_service")

// 更新连接池统计
stats := &PoolStats{
    ActiveCount: 10,
    IdleCount: 5,
    TotalCount: 15,
}
redisCollector.UpdatePoolStats(stats)

// 记录命令执行
redisCollector.ObserveCommand("GET", 0.05, nil) // 成功的 GET 操作
redisCollector.ObserveCommand("SET", 0.08, errors.New("timeout")) // 失败的 SET 操作
```

#### 3. 业务指标监控
```go
// 创建业务收集器
businessCollector := collector.NewBusinessCollector("my_service")

// 记录业务操作
businessCollector.ObserveOperation("create_order", 0.5, nil) // 成功创建订单
businessCollector.ObserveOperation("payment", 1.2, errors.New("insufficient_balance")) // 支付失败

// 监控队列
businessCollector.SetQueueSize("order_queue", 100)
businessCollector.SetProcessRate("order_processing", 50.0)
```

#### 4. 系统资源监控
```go
// 创建系统收集器
systemCollector := collector.NewSystemCollector("my_service")

// 自动收集系统指标
go func() {
    for {
        systemCollector.Collect()
        time.Sleep(time.Second * 15)
    }
}()
```

### 高级使用场景

#### 1. 带标签的指标监控
```go
// 创建带标签的计数器
counter := metric.NewCounter(prometheus.CounterOpts{
    Namespace: "my_service",
    Name: "requests_total",
    Help: "Total number of requests",
}).WithLabels("method", "status", "path")

// 记录不同标签组合的指标
counter.WithLabelValues("GET", "200", "/api/users").Inc()
counter.WithLabelValues("POST", "500", "/api/orders").Inc()
```

#### 2. 自定义采样监控
```go
// 创建采样器
sampler := sampler.NewSampler(0.1) // 10% 采样率

// 在高频操作中使用采样
func processRequest(req *Request) {
    if sampler.ShouldSample() {
        // 记录指标
        histogram.Observe(req.Duration)
    }
}
```

#### 3. 分布式系统监控
```go
// 创建导出器
exporter, err := exporter.NewExporter(&configTypes.Config{
    Enabled:   true,
    Port:      9090,
    Path:      "/metrics",
    Namespace: "distributed_system",
    Labels: map[string]string{
        "region": "us-west",
        "dc":     "dc1",
    },
})

// 注册多个收集器
exporter.RegisterCollector(httpCollector)
exporter.RegisterCollector(redisCollector)
exporter.RegisterCollector(businessCollector)

// 启动指标导出服务
if err := exporter.Start(context.Background()); err != nil {
    log.Fatal(err)
}
```

#### 4. 批量指标处理
```go
// 创建批量处理器
batchProcessor := NewBatchProcessor(100) // 批量大小

// 批量记录指标
for i := 0; i < 1000; i++ {
    batchProcessor.Add(func() {
        counter.WithLabelValues("batch_op").Inc()
    })
}

// 执行批量处理
batchProcessor.Process()
```

### 监控告警集成

#### 1. Grafana 告警规则示例
```yaml
# HTTP 服务告警
- alert: HighErrorRate
  expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: High HTTP error rate

# Redis 连接池告警
- alert: RedisPoolExhausted
  expr: redis_pool_active_connections > redis_pool_total_connections * 0.9
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: Redis connection pool near capacity
```

## API文档

### Exporter
```go
type Exporter interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    GetHTTPCollector() collector.HTTPCollector
    GetBusinessCollector() collector.BusinessCollector
    GetSystemCollector() collector.SystemCollector
}
```

### Collector
```go
type Collector interface {
    prometheus.Collector
    Register() error
}
```

### Metric
```go
type Metric interface {
    prometheus.Collector
    Register() error
    GetCollector() prometheus.Collector
}
```

## 指标类型

### 1. Counter
- 单调递增的计数器
- 适用场景：请求总数、错误总数
- 主要方法：
  - Inc()
  - Add(float64)
  - WithLabelValues(...string)

### 2. Gauge
- 可增可减的数值
- 适用场景：内存使用量、队列长度
- 主要方法：
  - Set(float64)
  - Inc()
  - Dec()
  - Add(float64)
  - Sub(float64)
  - WithLabelValues(...string)

### 3. Histogram
- 数值分布统计
- 适用场景：请求延迟、响应大小
- 主要方法：
  - Observe(float64)
  - WithLabelValues(...string)

## 内置收集器

### 1. BusinessCollector
```go
type BusinessCollector struct {
    operationTotal     metric.Counter   // 业务操作计数器
    operationDuration  metric.Histogram // 业务操作延迟
    operationErrors    metric.Counter   // 业务操作错误计数
    queueSize         metric.Gauge     // 业务队列大小
    processRate       metric.Gauge     // 业务处理速率
}
```

### 2. HTTPCollector
```go
type HTTPCollector struct {
    requestTotal     metric.Counter   // 请求总数
    requestDuration  metric.Histogram // 请求延迟
    activeRequests   metric.Gauge     // 活跃请求数
    requestSize      metric.Histogram // 请求大小
    responseSize     metric.Histogram // 响应大小
    slowRequests     metric.Counter   // 慢请求计数器
}
```

### 3. RedisCollector
```go
type RedisCollector struct {
    poolActive   metric.Gauge     // 活跃连接数
    poolIdle     metric.Gauge     // 空闲连接数
    poolTotal    metric.Gauge     // 总连接数
    cmdTotal     metric.Counter   // 命令执行总数
    cmdErrors    metric.Counter   // 命令执行错误数
    cmdDuration  metric.Histogram // 命令执行耗时
}
```

### 4. SystemCollector
监控系统级指标：
- CPU 使用率
- 内存使用情况
- 文件描述符
- 网络连接数

### 5. RuntimeCollector
监控 Go 运行时指标：
- Goroutine 数量
- 内存分配统计
- GC 统计

### 6. ResourceCollector
系统资源指标：
- CPU 使用率
- 内存使用情况
- 磁盘使用情况

## 性能优化

### 采样策略
- 合理设置采样率
- 针对高频指标启用采样
- 动态调整采样率

### 标签优化
- 避免高基数标签
- 合理使用标签组合
- 定期清理过期标签

### 批量处理
- 使用批量接口
- 合理设置缓冲区大小
- 异步处理非关键指标

## 错误处理

### 错误类型
1. 配置错误
2. 注册错误
3. 采集错误
4. 导出错误

### 错误处理示例
```go
if err := collector.Register(); err != nil {
    switch {
    case errors.Is(err, prometheus.AlreadyRegisteredError{}):
        // 处理重复注册
    case errors.Is(err, InvalidConfigError):
        // 处理配置错误
    default:
        // 处理其他错误
    }
}
```

## 最佳实践

### 1. 指标命名
- 使用有意义的名称
- 遵循命名规范
- 添加清晰的帮助文本

### 2. 标签使用
- 避免高基数标签
- 使用有意义的标签键
- 控制标签数量

### 3. 性能考虑
- 合理使用采样
- 异步处理非关键指标
- 批量处理指标数据

### 4. 运维建议
- 监控指标数量
- 定期清理过期指标
- 设置合理的告警阈值

## 测试指南

### 指标验证最佳实践

#### 1. Counter 验证
```go
// 不要这样做
counter.Get()  // ❌ 错误：Counter 没有 Get 方法

// 正确的做法
assert.NotPanics(t, func() {
    counter.WithLabelValues("label1").Inc()
}, "计数器应该能正常递增")
```

#### 2. Gauge 验证
```go
// 不要这样做
gauge.Get()  // ❌ 错误：Gauge 没有 Get 方法

// 正确的做法
assert.NotPanics(t, func() {
    gauge.Set(123.45)
}, "仪表盘应该能正常设置值")
```

#### 3. Histogram 验证
```go
// 不要这样做
histogram.Get()  // ❌ 错误：Histogram 没有 Get 方法

// 正确的做法
assert.NotPanics(t, func() {
    histogram.Observe(0.123)
}, "直方图应该能正常观察值")
```

### 测试注意事项
1. 不要尝试直接读取指标值，这不是推荐的测试方法
2. 使用 NotPanics 断言来验证指标操作的正确性
3. 关注指标记录的可用性，而不是具体的值
4. 在实际业务场景中验证指标的记录

## 常见问题

### 1. 指标重复注册
```go
// 正确处理
if err := collector.Register(); err != nil {
    if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
        return err
    }
}
```

### 2. 内存占用过高
- 减少标签组合
- 启用采样
- 及时清理过期指标

### 3. 性能问题
- 使用 pprof 分析
- 优化采集频率
- 合理使用缓存


