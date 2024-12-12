# Jaeger 链路追踪模块

## 简介

Jaeger 链路追踪模块提供了分布式追踪功能，帮助开发者追踪和监控分布式系统中的请求流程。本模块基于 OpenTracing 规范实现，支持：

- 分布式上下文传播
- 灵活的采样策略
- HTTP/gRPC 协议支持
- 自定义标签和日志
- 异步缓冲上报

## 快速开始

### 1. 配置文件

在 `config.yaml` 中添加 Jaeger 配置：

```yaml
jaeger:
  enable: true
  service_name: "your-service-name"
  agent:
    host: "localhost"
    port: "6831"
  collector:
    endpoint: "http://localhost:14268/api/traces"
    timeout: 5s
  sampler:
    type: "const"    # 可选: const, probabilistic, rateLimiting, remote
    param: 1         # 采样参数
  buffer:
    enable: true
    size: 1000
    flush_interval: 1s
  tags:              # 全局标签
    env: "dev"
    version: "1.0.0"
```

### 2. 创建 Tracer Provider

```go
import "gobase/pkg/trace/jaeger"

provider, err := jaeger.NewProvider()
if err != nil {
    // 处理错误
}
defer provider.Close()
```

### 3. 创建和使用 Span

#### 基础用法

```go
// 创建根 Span
span := provider.Tracer().StartSpan("operation-name")
defer span.Finish()

// 添加标签
span.SetTag("key", "value")

// 记录日志
span.LogKV("event", "test-event")
```

#### HTTP 请求追踪

```go
// 服务端
func HandleHTTP(w http.ResponseWriter, r *http.Request) {
    span, err := jaeger.StartSpanFromHTTP(r, "http-operation")
    if err != nil {
        // 处理错误
    }
    defer span.Finish()
    
    // 处理请求
}

// 客户端
func MakeHTTPRequest(ctx context.Context, url string) {
    req, _ := http.NewRequest("GET", url, nil)
    
    span, ctx := jaeger.StartSpanFromContext(ctx, "http-client")
    defer span.Finish()
    
    // 注入追踪信息
    carrier := jaeger.HTTPCarrier(req.Header)
    err := jaeger.Inject(span.Context(), carrier)
    if err != nil {
        // 处理错误
    }
    
    // 发送请求
}
```

#### gRPC 请求追踪

```go
// 服务端
func (s *Server) HandleGRPC(ctx context.Context, req *pb.Request) (*pb.Response, error) {
    span, err := jaeger.StartSpanFromGRPC(ctx, "grpc-operation")
    if err != nil {
        // 处理错误
    }
    defer span.Finish()
    
    // 处理请求
    return &pb.Response{}, nil
}

// 客户端
func MakeGRPCRequest(ctx context.Context) {
    span, ctx := jaeger.StartSpanFromContext(ctx, "grpc-client")
    defer span.Finish()
    
    md := metadata.New(nil)
    carrier := jaeger.MetadataCarrier(md)
    err := jaeger.Inject(span.Context(), carrier)
    if err != nil {
        // 处理错误
    }
    
    // 创建带追踪信息的上下文
    ctx = metadata.NewOutgoingContext(ctx, md)
    
    // 发送 gRPC 请求
}
```

### 4. 采样策略

支持以下采样策略：

- `const`: 常量采样 (0 或 1)
- `probabilistic`: 概率采样 (0.0 - 1.0)
- `rateLimiting`: 速率限制采样
- `remote`: 远程动态采样

```go
// 配置示例
sampler:
  type: "probabilistic"  # 概率采样
  param: 0.1            # 采样 10% 的请求
```

### 5. 高级特性

#### 自定义标签

```go
span.SetTag("custom.tag", "value")
span.SetTag("error", true)
span.SetTag("http.status_code", 200)
```

#### 带选项的 Span 创建

```go
span := jaeger.NewSpan(
    "operation-name",
    jaeger.WithTag("key", "value"),
    jaeger.WithParent(ctx),
    jaeger.WithStartTime(time.Now()),
    jaeger.WithSample(true),
)
defer span.Finish()
```

#### 异步缓冲上报

```yaml
buffer:
  enable: true         # 启用缓冲
  size: 1000          # 缓冲区大小
  flush_interval: 1s   # 刷新间隔
```

## 本地开发

### 启动 Jaeger

使用 Docker Compose 启动本地 Jaeger：

```bash
docker-compose -f docker/local/docker-compose.yml up -d
```

访问 Jaeger UI: http://localhost:16686

### 运行测试

```bash
# 运行单元测试
go test ./pkg/trace/jaeger/tests/unit/...

# 运行集成测试
go test ./pkg/trace/jaeger/tests/integration/...
```

## 注意事项

1. 在生产环境中，建议根据实际流量调整采样策略
2. 使用 `defer span.Finish()` 确保 Span 被正确关闭
3. 错误处理时记得设置错误标签：`span.SetTag("error", true)`
4. 大量并发请求时，建议启用缓冲功能以提高性能
5. 定期检查 Jaeger UI 监控追踪数据

## 错误码

- `JaegerInitError`: Jaeger 初始化错误
- `JaegerShutdownError`: Jaeger 关闭错误
- `JaegerInjectError`: 追踪信息注入错误
- `JaegerExtractError`: 追踪信息提取错误