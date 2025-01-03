# Jaeger 分布式追踪模块

## 目录
- [简介](#简介)
- [特性](#特性)
- [架构设计](#架构设计)
- [快速开始](#快速开始)
- [配置说明](#配置说明)
- [使用指南](#使用指南)
- [采样策略](#采样策略)
- [高级特性](#高级特性)
- [本地开发](#本地开发)
- [错误处理](#错误处理)
- [测试指南](#测试指南)

## 简介

Jaeger 分布式追踪模块提供了完整的链路追踪功能,支持:
- 多种采样策略
- 上下文传播
- HTTP/gRPC 集成
- 异步缓冲上报
- 优雅启动关闭

## 特性

- 支持多种采样策略:
  - 常量采样 (const) - 已实现
  - 概率采样 (probabilistic) - 已实现
  - 速率限制采样 (rateLimiting) - 已实现
  - 远程采样 (remote) - 部分实现, 刷新机制待开发
  - 自适应采样 (adaptive) - 预留接口, 计划未来实现

- 上下文传播:
  - HTTP Header 传播
  - gRPC Metadata 传播
  - Context 传播

- 可观测性:
  - 详细的 Span 标签
  - 错误记录
  - 采样决策记录
  - 性能指标收集

## 架构设计

### 核心组件

1. Provider: 全局 Tracer 提供者
2. Span: 追踪单元封装
3. Sampler: 采样决策器
4. Context: 上下文传播
5. Config: 配置管理

### 工作流程

```
请求 -> Context提取 -> 采样决策 -> Span创建 -> 标签/日志记录 -> 异步上报
```

## 配置说明

```yaml
jaeger:
  enable: true                     # 启用开关
  service_name: "my-service"       # 服务名称
  
  agent:                          # 代理配置
    host: "localhost"
    port: "6831"
    
  collector:                      # 收集器配置
    endpoint: "http://localhost:14268/api/traces"
    username: ""                  # 可选
    password: ""                  # 可选
    timeout: 5s
    
  sampler:                        # 采样配置
    type: "probabilistic"         # 采样类型
    param: 0.1                    # 采样参数
    server_url: ""               # 远程采样服务地址
    max_operations: 2000         # 最大操作数
    refresh_interval: 60         # 刷新间隔(秒)
    
  buffer:                        # 缓冲配置
    enable: true
    size: 1000                  # 缓冲区大小
    flush_interval: 1s          # 刷新间隔
    
  tags:                         # 全局标签
    environment: "production"
```

## 使用指南

### 基础使用

1. 创建 Provider:
```go
provider, err := jaeger.NewProvider()
if err != nil {
    return err
}
defer provider.Close()
```

2. 创建 Span:
```go
span := jaeger.NewSpan(
    "operation-name",
    jaeger.WithTag("key", "value"),
    jaeger.WithParent(ctx),
)
defer span.Finish()
```

### HTTP 集成

```go
// 服务端
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        span, err := jaeger.StartSpanFromHTTP(r, "http-handler")
        if err != nil {
            // 错误处理
            return
        }
        defer span.Finish()
        
        ctx := opentracing.ContextWithSpan(r.Context(), span)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// 客户端
func makeRequest(ctx context.Context, url string) error {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return err
    }
    
    carrier := jaeger.HTTPCarrier(req.Header)
    if err := jaeger.Inject(carrier); err != nil {
        return err
    }
    
    // 发送请求
    ...
}
```

### gRPC 集成

```go
// 服务端
func unaryInterceptor(ctx context.Context, req interface{}, 
    info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
    
    span, err := jaeger.StartSpanFromGRPC(ctx, info.FullMethod)
    if err != nil {
        return nil, err
    }
    defer span.Finish()
    
    return handler(opentracing.ContextWithSpan(ctx, span), req)
}

// 客户端
func makeGRPCRequest(ctx context.Context) error {
    md, ok := metadata.FromOutgoingContext(ctx)
    if !ok {
        md = metadata.New(nil)
    }
    
    carrier := jaeger.MetadataCarrier(md)
    if err := jaeger.Inject(carrier); err != nil {
        return err
    }
    
    ctx = metadata.NewOutgoingContext(ctx, metadata.MD(carrier))
    // 发送 gRPC 请求
    ...
}
```

## 采样策略

### 已实现的采样策略

1. 常量采样 (const)
   - 全采样(1)或不采样(0)
   - 适用于开发环境或低流量场景

2. 概率采样 (probabilistic)
   - 按比例采样
   - 适用于稳定的生产环境

3. 速率限制采样 (rateLimiting)
   - 按每秒采样数量限制
   - 适用于高流量系统的资源控制

### 待完善的采样策略

1. 远程采样 (remote)
   - 基础功能已实现
   - 动态刷新机制待开发 (TODO)
   - 计划支持从中心服务动态更新采样策略

2. 自适应采样 (adaptive)
   - 接口已预留
   - 计划未来实现以下功能:
     * 基于系统负载自动调整采样率
     * 基于请求特征智能采样
     * 支持自定义采样规则

## 高级特性

### 异步缓冲上报

```yaml
buffer:
  enable: true
  size: 1000          # 缓冲区大小
  flush_interval: 1s  # 刷新间隔
```

### 自定义标签
```go
span.SetTag("custom.tag", "value")
span.SetTag("error", true)
span.SetTag("http.status_code", 200)
```

### 错误处理
```go
if err != nil {
    span.SetError(err)  // 自动记录错误详情
}
```

## 错误码

- `JaegerInitError`: Jaeger 初始化错误
- `JaegerShutdownError`: Jaeger 关闭错误
- `JaegerInjectError`: 追踪信息注入错误
- `JaegerExtractError`: 追踪信息提取错误
- `JaegerSpanError`: Span 操作错误
- `JaegerSamplerError`: 采样器错误
- `JaegerPropagateError`: 上下文传播错误

## 本地开发

### 启动 Jaeger

```bash
docker-compose -f docker/local/docker-compose.yml up -d
```

访问 UI: http://localhost:16686

### 运行测试

```bash
# 单元测试
go test ./pkg/trace/jaeger/tests/unit/...

# 集成测试
go test ./pkg/trace/jaeger/tests/integration/...
```

## 注意事项

1. 在生产环境中,建议:
   - 根据实际流量调整采样策略
   - 启用异步缓冲上报
   - 合理设置缓冲区大小和刷新间隔
   
2. 开发最佳实践:
   - 总是使用 defer span.Finish()
   - 正确处理错误并记录到 span
   - 使用有意义的操作名称
   - 合理使用标签,避免过多标签