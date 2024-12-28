# Logger Middleware

高性能、可配置的 HTTP 请求日志中间件，支持 ELK 集成、Prometheus 指标收集、分布式追踪等特性。

## 目录
- [特性](#特性)
- [快速开始](#快速开始)
  - [基础用法](#基础用法)
  - [自定义配置](#自定义配置)
  - [自定义字段](#自定义字段)
  - [自定义格式化器](#自定义格式化器)
  - [自定义采样策略](#自定义采样策略)
  - [异步写入配置](#异步写入配置)
  - [ELK 集成](#elk-集成)
  - [Jaeger 集成](#jaeger-集成)
  - [错误处理](#错误处理)
  - [优雅关闭](#优雅关闭)
- [日志输出示例](#日志输出示例)
  - [JSON 格式](#json-格式)
  - [Text 格式](#text-格式)
- [配置说明](#配置说明)
  - [基础配置](#基础配置)
  - [指标配置](#指标配置-metrics)
  - [追踪配置](#追踪配置-trace)
  - [日志轮转配置](#日志轮转配置-rotate)
  - [缓冲配置](#缓冲配置-buffer)
- [性能优化](#性能优化)
- [监控指标](#监控指标)
  - [自定义指标](#自定义指标)
- [故障排查](#故障排查)
- [最佳实践](#最佳实践)
  - [生产环境配置示例](#生产环境配置示例)
  - [开发环境配置示例](#开发环境配置示例)

## 特性

- 自动记录 HTTP 请求/响应信息
- 支持 JSON/Text 格式输出
- 支持日志轮转
- 支持采样控制
- 支持慢请求识别
- 支持请求体/响应体记录
- 支持 ELK 集成
- 支持 Prometheus 指标
- 支持 Jaeger 追踪
- 支持异步缓冲写入
- 支持优雅关闭
- 支持自定义日志字段
- 支持自定义格式化器
- 支持自定义采样策略
- 支持自定义错误处理

## 快速开始

### 基础用法

```go
import (
"github.com/gin-gonic/gin"
"gobase/pkg/middleware/logger"
)
func main() {
r := gin.New()
// 使用默认配置
logMiddleware := logger.Middleware()
r.Use(logMiddleware)
r.Run(":8080")
}
```

### 自定义配置

```go
logMiddleware := logger.Middleware(
logger.WithConfig(&logger.Config{
Enable: true,
Level: "info",
Format: "json",
SampleRate: 1.0,
SlowThreshold: 200, // 200ms
RequestBodyLimit: 1024, // 1KB
}),
)
```

### 自定义字段

```go
logMiddleware := logger.Middleware(
logger.WithCustomFields(map[string]interface{}{
"app_name": "my-service",
"env":      "production",
"version":  "v1.0.0",
}),
)
```

### 自定义格式化器

```go
type CustomFormatter struct{}

func (f *CustomFormatter) Format(param *logger.LogFormatterParam) map[string]interface{} {
return map[string]interface{}{
"custom_time":   param.TimeStamp.Format(time.RFC3339),
"custom_method": param.Method,
// ... 其他自定义字段
}
}

logMiddleware := logger.Middleware(
logger.WithFormatter(&CustomFormatter{}),
)
```

### 自定义采样策略

```go
type CustomSampler struct{}

func (s *CustomSampler) Sample(c *gin.Context) bool {
// 自定义采样逻辑
return true
}

logMiddleware := logger.Middleware(
logger.WithSampler(&CustomSampler{}),
)
```

### 异步写入配置

```go
logMiddleware := logger.Middleware(
logger.WithAsyncConfig(&logger.AsyncConfig{
Enable:        true,
BufferSize:    1000,
FlushInterval: time.Second,
FlushOnError:  true,
}),
)
```

### ELK 集成

```go
import (
"gobase/pkg/logger/elk"
)
// 配置 ELK
elkConfig := &elk.Config{
Addresses: []string{"http://elasticsearch:9200"},
Index: "app-logs",
BatchConfig: &elk.BulkProcessorConfig{
BatchSize: 1000,
Interval: time.Second,
},
}
// 创建 ELK Hook
elkHook, _ := elk.NewHook(elkConfig)
// 添加到 logger
logMiddleware := logger.Middleware(
logger.WithHook(elkHook),
)
```

### Jaeger 集成

```go
logMiddleware := logger.Middleware(
logger.WithTracer(jaegerTracer),
logger.WithTracingConfig(&logger.TracingConfig{
Enable:      true,
HeaderName:  "X-Trace-ID",
ContextKey:  "trace_id",
}),
)
```

### 错误处理

```go
logMiddleware := logger.Middleware(
logger.WithErrorHandler(func(c *gin.Context, err error) {
// 自定义错误处理逻辑
logger.Error("Request error", err)
}),
)
```

### 优雅关闭

```go
func main() {
r := gin.New()
logMiddleware := logger.Middleware()
r.Use(logMiddleware)

// 注册关闭钩子
shutdown := make(chan os.Signal, 1)
signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

go func() {
<-shutdown
// 等待日志写入完成
logger.WaitForAsyncFlush(5 * time.Second)
os.Exit(0)
}()

r.Run(":8080")
}
```

## 日志输出示例

### JSON 格式

```json
{
"level": "info",
"time": "2024-03-12T10:00:00Z",
"method": "POST",
"path": "/api/users",
"status": 200,
"latency_ms": 50,
"client_ip": "127.0.0.1",
"request_id": "req-123456",
"user_agent": "curl/7.64.1",
"request_size": 1024,
"response_size": 2048,
"trace_id": "trace-123456",
"error": null,
"app_name": "my-service",
"env": "production"
}
```

### Text 格式

```text
[2024-03-12T10:00:00Z] INFO POST /api/users 200 50ms 127.0.0.1 req-123456
```

## 配置说明

### 基础配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| Enable | 是否启用中间件 | true |
| Level | 日志级别 | "info" |
| Format | 日志格式 (json/text) | "json" |
| SampleRate | 采样率 (0.0-1.0) | 1.0 |
| SlowThreshold | 慢请求阈值(毫秒) | 200 |
| RequestBodyLimit | 请求体记录阈值(字节) | 1024 |
| ResponseBodyLimit | 响应体记录阈值(字节) | 1024 |
| SkipPaths | 跳过记录的路径 | ["/health", "/metrics"] |

### 指标配置 (Metrics)

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| Enable | 是否启用指标收集 | true |
| Prefix | 指标前缀 | "http_request" |
| EnableLatencyHistogram | 是否收集延迟直方图 | true |
| EnableSizeHistogram | 是否收集大小直方图 | true |

### 追踪配置 (Trace)

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| Enable | 是否启用追踪 | true |
| SamplerType | 采样类型 | "const" |
| SamplerParam | 采样参数 | 1 |

### 日志轮转配置 (Rotate)

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| Enable | 是否启用轮转 | true |
| MaxSize | 单个文件最大尺寸(MB) | 100 |
| MaxAge | 文件保留天数 | 7 |
| MaxBackups | 最大备份数 | 5 |
| Compress | 是否压缩 | true |
| FilePath | 日志文件路径 | "./logs/app.log" |

### 缓冲配置 (Buffer)

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| Enable | 是否启用缓冲 | true |
| Size | 缓冲区大小(字节) | 4096 |
| FlushInterval | 刷新间隔(毫秒) | 1000 |
| FlushOnError | 错误时是���立即刷新 | true |

## 性能优化

1. 合理设置采样率，高流量场景建议降低采样率
2. 适当调整请求体/响应体记录阈值
3. 启用缓冲写入减少 I/O 压力
4. 配置合适的日志轮转策略
5. 使用 JSON 格式便于后续处理
6. 使用异步写入提高性能
7. 合理配置缓冲区大小

## 监控指标

中间件自动收集以下 Prometheus 指标：

- `http_request_total`: 请求总数
- `http_request_duration_seconds`: 请求延迟分布
- `http_request_size_bytes`: 请求大小分布
- `http_response_size_bytes`: 响应大小分布
- `http_request_errors_total`: 错误请求总数

### 自定义指标

```go
logMiddleware := logger.Middleware(
logger.WithMetricsConfig(&logger.MetricsConfig{
Enable:                true,
Prefix:               "custom_http",
EnableLatencyBuckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1},
}),
)
```

## 故障排查

1. 日志未写入
   - 检查 Enable 配置是否为 true
   - 检查日志级别设置
   - 检查文件权限

2. ELK 集成问题
   - 检查 Elasticsearch 连接配置
   - 查看 bulk processor 错误日志
   - 检查网络连接

3. 性能问题
   - 检查采样率设置
   - 调整缓冲区配置
   - 检查磁盘 I/O

## 最佳实践

### 生产环境配置示例

```go
logMiddleware := logger.Middleware(
logger.WithConfig(&logger.Config{
Enable:     true,
Level:      "info",
Format:     "json",
SampleRate: 0.1,
}),
logger.WithAsyncConfig(&logger.AsyncConfig{
Enable:        true,
BufferSize:    5000,
FlushInterval: time.Second,
}),
logger.WithCustomFields(map[string]interface{}{
"env": "production",
}),
logger.WithHook(elkHook),
)
```

### 开发环境配置示例

```go
logMiddleware := logger.Middleware(
logger.WithConfig(&logger.Config{
Enable:     true,
Level:      "debug",
Format:     "text",
SampleRate: 1.0,
}),
logger.WithCustomFields(map[string]interface{}{
"env": "development",
}),
)
```