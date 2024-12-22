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
这是一个基于Prometheus的监控模块，提供了完整的指标收集、导出和监控功能。该模块设计简洁、易用、易扩展，适用于分布式系统的监控需求。

### 主要特性
- 支持多种指标类型：Counter、Gauge、Histogram
- 内置多个指标收集器
- 支持指标采样
- 支持自定义标签
- 支持优雅关闭
- 完整的测试覆盖
- 支持容器化部署

## 架构设计

### 核心组件
1. Exporter: 指标导出器
2. Collector: 指标收集器
3. Metric: 指标类型实现
4. Sampler: 采样器
5. Config: 配置管理

### 数据流

[业务代码] -> [Collector] -> [Metric] -> [Exporter] -> [Prometheus Server]

## 目录结构

pkg/monitor/prometheus/
├── collector/ # 指标收集器
├── config/ # 配置管理
├── exporter/ # 指标导出器
├── metric/ # 指标类型
├── sampler/ # 采样器
└── tests/ # 测试用例
├── integration/ # 集成测试
├── unit/ # 单元测试
└── testutils/ # 测试工具

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

### 配置文件示例 (config.yaml)



```yaml
prometheus:
enabled: true
port: 9090
path: "/metrics"
namespace: "myapp"
labels:
app: "myapp"
env: "prod"
collectors:
"business"
"http"
"system"
"runtime"
"resource"
sampling:
enabled: true
rate: 0.1
```

### 配置项说明
| 配置项 | 类型 | 必填 | 默认值 | 说明 |
|-------|------|------|--------|------|
| enabled | bool | 是 | false | 是否启用Prometheus监控 |
| port | int | 是 | 9090 | metrics接口端口 |
| path | string | 是 | "/metrics" | metrics接口路径 |
| namespace | string | 是 | - | 指标命名空间 |
| labels | map | 否 | {} | 全局标签 |
| collectors | []string | 否 | [] | 启用的收集器列表 |
| sampling.enabled | bool | 否 | false | 是否启用采样 |
| sampling.rate | float | 否 | 1.0 | 采样率(0.0-1.0) |

## 使用指南

### 1. 初始化导出器

```go
import (
"gobase/pkg/monitor/prometheus/exporter"
"gobase/pkg/monitor/prometheus/config/types"
)
func initPrometheus(ctx context.Context) error {
cfg := &types.Config{
Enabled: true,
Port: 9090,
Path: "/metrics",
Namespace: "myapp",
Labels: map[string]string{
"app": "myapp",
"env": "prod",
},
Collectors: []string{
"business",
"http",
"system",
},
}
exp, err := exporter.New(cfg, logger)
if err != nil {
return fmt.Errorf("failed to create exporter: %w", err)
}
return exp.Start(ctx)
}
```

### 2. 使用HTTP收集器

```go
func handleRequest(w http.ResponseWriter, r http.Request) {
start := time.Now()
httpCollector := exp.GetHTTPCollector()
defer func() {
duration := time.Since(start)
httpCollector.ObserveRequest(
r.Method,
r.URL.Path,
http.StatusOK,
duration,
calculateRequestSize(r),
calculateResponseSize(w),
)
}()
// 处理请求...
}
```

### 3. 使用业务收集器

```go
func processOrder(order Order) error {
start := time.Now()
businessCollector := exp.GetBusinessCollector()
defer func() {
duration := time.Since(start)
businessCollector.ObserveOperation(
"process_order",
duration,
map[string]string{
"type": order.Type,
"status": "success",
},
)
}()
// 处理订单...
}
```


### 4. 自定义指标

```go
// 创建计数器
counter := metric.NewCounter(prometheus.CounterOpts{
Namespace: "myapp",
Name: "custom_operations_total",
Help: "Total number of custom operations",
}).WithLabels("operation", "status")
// 注册指标
if err := counter.Register(); err != nil {
return err
}
// 使用指标
counter.WithLabelValues("create", "success").Inc()
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
operationTotal metric.Counter // 业务操作计数器
operationDuration metric.Histogram // 业务操作延迟
operationErrors metric.Counter // 业务操作错误计数
queueSize metric.Gauge // 业务队列大小
processRate metric.Gauge // 业务处理速率
}
```

### 2. HTTPCollector

```go
type HTTPCollector struct {
requestTotal metric.Counter // 请求总数
requestDuration metric.Histogram // 请求延迟
activeRequests metric.Gauge // 活跃请求数
requestSize metric.Histogram // 请求大小
responseSize metric.Histogram // 响应大小
slowRequests metric.Counter // 慢请求计数器
}
```

### 3. ResourceCollector

```go
type ResourceCollector struct {
cpuUsage prometheus.Gauge // CPU使用率
memUsage prometheus.Gauge // 内存使用
memTotal prometheus.Gauge // 总内存
memFree prometheus.Gauge // 空闲内存
diskUsage prometheus.GaugeVec // 磁盘使用
}
```

### 4. RuntimeCollector

```go
type RuntimeCollector struct {
goroutines metric.Gauge // goroutine数量
memAlloc metric.Gauge // 已分配内存
memTotal metric.Gauge // 总内存
memSys metric.Gauge // 系统内存
gcPause metric.Histogram // GC暂停时间
gcCount metric.Counter // GC次数
}
```

### 5. SystemCollector

```go
type SystemCollector struct {
cpuUsage metric.Gauge // CPU使用率
memUsage metric.Gauge // 内存使用
loadAverage metric.Gauge // 系统负载
fdUsage metric.Gauge // 文件描述符使用
netConnections metric.Gauge // 网络连接数
}
```

## 测试

### 运行测试

```bash
运行所有测试
go test ./... -v
运行集成测试
go test ./tests/integration -v
运行单元测试
go test ./tests/unit -v
运行基准测试
go test -bench=. ./...
```

### 测试覆盖率

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Docker测试

```bash
启动测试环境
docker-compose -f tests/docker-compose.yml up -d
运行集成测试
go test ./tests/integration -v
清理测试环境
docker-compose -f tests/docker-compose.yml down
```

## 性能优化

### 采样策略

```go
sampler := sampler.NewSampler(0.1) // 10%采样率
if sampler.ShouldSample() {
// 收集指标
}
```

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

## 常见问题

### 1. 指标重复注册

```go
// 正确处理
if err := collector.Register(); err != nil {
if , ok := err.(prometheus.AlreadyRegisteredError); !ok {
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