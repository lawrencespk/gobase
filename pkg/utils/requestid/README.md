# Request ID Generator

请求ID生成器包提供了一个灵活且高性能的请求ID生成解决方案。支持多种ID生成策略,包括UUID、Snowflake以及自定义生成器。

## 功能特性

- 支持多种ID生成策略
  - UUID 格式 (默认)
  - Snowflake 算法
  - 自定义生成器
- 高性能设计
  - 对象池复用
  - 并发安全
  - 低延迟
- 可配置性
  - 自定义前缀
  - 工作节点ID
  - 数据中心ID
- 完整测试覆盖
  - 单元测试
  - 集成测试
  - 性能测试
  - 并发测试

## 快速开始

### 基础使用

```go
// 使用默认配置(UUID生成器)
generator := requestid.NewGenerator(nil)
id := generator.Generate()

// 使用Snowflake生成器
opts := &requestid.Options{
    Type:         "snowflake",
    Prefix:       "svc",
    WorkerID:     1,
    DatacenterID: 1,
}
generator := requestid.NewGenerator(opts)
id := generator.Generate()
```

### 自定义生成器

```go
counter := 0
generator := requestid.NewCustomGenerator("custom", func() string {
    counter++
    return fmt.Sprintf("id-%d", counter)
})
id := generator.Generate()
```

## 配置选项

```go
type Options struct {
    // 生成器类型: uuid, snowflake, custom
    Type string
    // 自定义前缀
    Prefix string
    // 是否启用对象池
    EnablePool bool
    // Snowflake相关配置
    WorkerID     int64
    DatacenterID int64
}
```

### 默认配置

```go
DefaultOptions() *Options {
    return &Options{
        Type:         "uuid",
        Prefix:       "",
        EnablePool:   true,
        WorkerID:     1,
        DatacenterID: 1,
    }
}
```

## 性能指标

基准测试结果(在标准硬件上):

- UUID生成器: >10,000 ops/s
- Snowflake生成器: >50,000 ops/s

## 最佳实践

### 1. 选择合适的生成器

- UUID: 适用于分布式系统,无需协调
- Snowflake: 适用于需要排序和趋势分析的场景
- 自定义: 适用于特殊格式要求

### 2. 性能优化

- 启用对象池以减少内存分配
- 适当配置worker ID和datacenter ID避免冲突
- 批量生成时复用生成器实例

### 3. 错误处理

- 处理时钟回拨情况
- 优雅处理生成失败的情况
- 监控生成性能和错误率

## 注意事项

1. Snowflake生成器
   - 需要确保workerId和datacenterId在分布式环境中唯一
   - 需要处理时钟回拨情况
   - 依赖系统时钟的准确性

2. UUID生成器
   - 启用对象��可以提升性能
   - UUID的随机性可能导致索引效率较低

3. 自定义生成器
   - 需要确保生成的ID唯一性
   - 需要考虑并发安全性

## 示例

### Web服务集成

```go
func main() {
    // 创建生成器
    generator := requestid.NewGenerator(&requestid.Options{
        Type:   "snowflake",
        Prefix: "api",
    })

    // Gin中间件
    r := gin.New()
    r.Use(func(c *gin.Context) {
        requestID := generator.Generate()
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        c.Next()
    })
}
```

### 分布式追踪

```go
func TracingMiddleware(generator requestid.Generator) gin.HandlerFunc {
    return func(c *gin.Context) {
        traceID := generator.Generate()
        span := opentracing.StartSpan(c.Request.URL.Path)
        span.SetTag("request_id", traceID)
        defer span.Finish()
        
        c.Next()
    }
}
```

## 测试

包含完整的测试套件:

```bash
# 运行所有测试
go test -v ./...

# 运行基准测试
go test -bench=. -benchmem

# 运行集成测试
go test -tags=integration ./...
```