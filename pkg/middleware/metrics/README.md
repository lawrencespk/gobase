# Metrics Middleware

HTTP 请求指标收集中间件，用于收集请求相关的 Prometheus 指标。

## 特性

- 支持请求采样
- 记录活跃请求数
- 统计请求延迟
- 统计请求/响应大小
- 识别慢请求

## 快速开始

### 基础用法

```go
import (
    "github.com/gin-gonic/gin"
    "gobase/pkg/middleware/metrics"
    "gobase/pkg/monitor/prometheus/collector"
)

func main() {
    r := gin.New()
    
    // 创建 HTTP 指标收集器
    httpCollector := collector.NewHTTPCollector()
    
    // 使用默认配置
    metricsMiddleware := metrics.Middleware(httpCollector, &metrics.Config{})
    
    r.Use(metricsMiddleware)
    r.Run(":8080")
}
```

### 自定义配置

```go
metricsMiddleware := metrics.Middleware(httpCollector, &metrics.Config{
    EnableSampling:       true,    // 启用采样
    SamplingRate:        0.1,     // 采样率 10%
    EnableRequestSize:    true,    // 收集请求大小
    EnableResponseSize:   true,    // 收集响应大小
    SlowRequestThreshold: 200,     // 慢请求阈值 200ms
})
```

## 配置说明

| 配置项 | 说明 | 默��值 |
|--------|------|--------|
| EnableSampling | 是否启用采样 | false |
| SamplingRate | 采样率 (0.0-1.0) | 1.0 |
| EnableRequestSize | 是否收集请求体大小 | false |
| EnableResponseSize | 是否收集响应体大小 | false |
| SlowRequestThreshold | 慢请求阈值(毫秒) | 0 |

## 指标说明

中间件通过 HTTPCollector 收集以下指标：

1. 活跃请求数
```go
httpCollector.IncActiveRequests(method)  // 增加活跃请求计数
httpCollector.DecActiveRequests(method)  // 减少活跃请求计数
```

2. 请求指标
```go
httpCollector.ObserveRequest(
    method,          // 请求方法
    path,           // 请求路径
    status,         // 响应状态码
    duration,       // 请求耗时
    requestSize,    // 请求大小
    responseSize,   // 响应大小
)
```

3. 慢请求指标
```go
httpCollector.ObserveSlowRequest(
    method,    // 请求方法
    path,     // 请求路径
    duration, // 请求耗时
)
```

## 最佳实践

1. 采样配置
- 低流量场景可以不开启采样
- 高流量场景建议开启采样并设置合适的采样率
- 采样率建议范围：0.1-1.0

2. 请求大小统计
- 按需开启请求/响应大小统计
- 注意大小统计可能带来的性能影响

3. 慢请求监控
- 根据实际场景设置合适的慢请求阈��
- 建议范围：100ms-500ms
