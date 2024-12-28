# 日志系统使用文档

## 目录
- [简介](#简介)
- [系统架构](#系统架构)
- [功能特性](#功能特性)
- [快速开始](#快速开始)
- [详细配置](#详细配置)
- [高级特性](#高级特性)
- [性能优化](#性能优化)
- [监控指标](#监控指标)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)
- [开发计划](#开发计划)

## 简介

本日志系统是一个高性能、可扩展的分布式日志解决方案。基于 logrus 构建,集成了 ELK Stack,提供完整的日志收集、存储、查询能力。

主要特点:
- 高性能异步处理
- 智能批量处理
- 自动重试机制
- 内存优化管理
- 完整监控指标
- 灵活的扩展性

## 系统架构

```
┌─────────────┐    ┌──────────┐    ┌─────────┐    ┌───────────────┐
│ Application │───>│ LogQueue │───>│ Writers │───>│ Storage Layer │
└─────────────┘    └──────────┘    └─────────┘    └───────────────┘
       │                                                   │
       │            ┌──────────┐    ┌─────────┐          │
       └────────────│  Hooks   │───>│  ELK    │──────────┘
                    └──────────┘    └─────────┘
```

## 功能特性

### 核心功能
- 多级别日志(Debug/Info/Warn/Error/Fatal)
- 结构化日志输出
- 自定义字段支持
- 调用者信息记录
- 错误堆栈跟踪
- 异步日志写入
- 日志文件轮转
- 日志压缩存储
- 过期日志清理

### 文件管理
- 自动创建日志目录
- 文件权限管理
- 文件句柄复用
- 自动轮转清理
- 压缩归档支持

### 内存管理
- 对象池复用
- 内存使用限制
- 缓冲区管理
- GC优化策略

### ELK集成
- 自动批量处理
- 索引管理
- 模板管理
- 文档管理
- 查询支持
- 错误重试
- 性能监控

## 快速开始

### 基础使用
```go
// 1. 创建日志配置
opts := logrus.DefaultOptions()
opts.OutputPaths = []string{"stdout", "./logs/app.log"}

// 2. 初始化日志器
logger, err := logrus.NewLogger(opts)
if err != nil {
    panic(err)
}
defer logger.Close()

// 3. 记录日志
logger.Info("应用启动成功")
logger.WithFields(logrus.Fields{
    "user_id": "12345",
    "action": "login",
}).Info("用户登录")
```

### ELK集成
```go
// 1. ELK配置
elkConfig := elk.Config{
    Addresses: []string{"http://elasticsearch:9200"},
    Username: "elastic",
    Password: "password",
    Index: "app-logs",
}

// 2. 创建Hook
hook, err := elk.NewHook(elkConfig)
if err != nil {
    panic(err)
}

// 3. 添加Hook
logger.AddHook(hook)
```

## 详细配置

### 日志器配置
```go
type Options struct {
    // 基础配置
    Level            string   // 日志级别
    Format           string   // 日志格式(text/json)
    OutputPaths      []string // 输出路径
    ErrorOutputPaths []string // 错误输出路径
    
    // 文件配置
    MaxSize        int  // 单个文件最大尺寸(MB)
    MaxAge         int  // 文件最大保留天数
    MaxBackups     int  // 最大备份数量
    Compress       bool // 是否压缩
    
    // 异步配置
    Async          bool          // 是否异步
    QueueSize      int          // 队列大小
    FlushInterval  time.Duration // 刷新间隔
    
    // 格式化配置
    TimeFormat     string // 时间格式
    CallerSkip     int    // 调用者跳过层级
    ReportCaller   bool   // 是否记录调用者
}
```

### ELK配置
```go
type ElkConfig struct {
    // 连接配置
    Addresses    []string      // ES地址
    Username     string        // 用户名
    Password     string        // 密码
    Timeout      time.Duration // 超时时间
    
    // 索引配置
    Index        string // 索引名称
    IndexPattern string // 索引模式
    Type         string // 文档类型
    
    // 批处理配置
    BatchSize    int           // 批量大小
    FlushBytes   int          // 刷新字节数
    FlushTimeout time.Duration // 刷新超时
    
    // 重试配置
    MaxRetries   int           // 最大重试次数
    RetryTimeout time.Duration // 重试超时
}
```

## 高级特性

### 异步处理
```go
// 配置异步处理
opts := &Options{
    Async: true,
    QueueSize: 10000,
    FlushInterval: time.Second,
}

// 优雅关闭
defer logger.Close()
```

### 批量处理
```go
// 配置批量处理
bulkConfig := &elk.BulkConfig{
    BatchSize: 1000,
    FlushBytes: 5 * 1024 * 1024,
    FlushTimeout: time.Second,
}

// 获取统计信息
stats := bulk.Stats()
fmt.Printf("处理文档数: %d\n", stats.ProcessedDocs)
```

### 错误重试
```go
// 配置重试策略
retryConfig := &elk.RetryConfig{
    MaxRetries: 3,
    RetryTimeout: time.Second * 30,
}

// 使用重试
err := elk.WithRetry(ctx, retryConfig, func() error {
    return client.Index(doc)
})
```

## 性能优化

### 内存优化
1. 对象池
```go
// 配置对象池
pool := &sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

// 使用对象池
buf := pool.Get().([]byte)
defer pool.Put(buf)
```

2. 批量处理
```go
// 配置批量参数
opts := &Options{
    BatchSize: 1000,      // 批量大小
    FlushInterval: time.Second, // 刷新间隔
}
```

3. 异步处理
```go
// 配置异步
opts := &Options{
    Async: true,
    QueueSize: 10000,
}
```

### 写入优化
1. 文件缓冲
```go
// 配置文件缓冲
opts := &Options{
    BufferSize: 256 * 1024, // 256KB
}
```

2. 批量写入
```go
// 配置批量写入
opts := &Options{
    BatchSize: 100,
    FlushInterval: time.Second,
}
```

## 监控指标

### 基础指标
- 总处理日志数
- 处理失败数
- 平均处理延迟
- 队列大小
- 内存使用量

### ELK指标
- 索引文档数
- 索引延迟
- 批量大小
- 重试次数
- 错误率

### 监控示例
```go
// 获取指标
metrics := logger.Metrics()
fmt.Printf("处理总数: %d\n", metrics.TotalLogs)
fmt.Printf("失败数: %d\n", metrics.FailedLogs)
fmt.Printf("平均延迟: %v\n", metrics.AvgLatency)
```

## 最佳实践

### 日志级别使用
- DEBUG: 调试信息
- INFO: 正常业务流程
- WARN: 需要注意的问题
- ERROR: 错误信息
- FATAL: 致命错误

### 性能优化建议
1. 合理配置
   - 根据业务量配置队列大小
   - 根据内存配置批量参数
   - 设置合适的刷新间隔

2. 内存管理
   - 使用对象池
   - 控制单条日志大小
   - 及时清理过期日志

3. 写入优化
   - 启用异步处理
   - 使用批量写入
   - 配置文件缓冲

### ELK最佳实践
1. 索引管理
   - 使用索引模板
   - 定期优化索引
   - 设置合理的分片

2. 批量处理
   - 合理的批量大小
   - 适当的刷新间隔
   - 监控处理延迟

3. 错误处理
   - 配置重试策略
   - 记录错误日志
   - 监控错误率

## 常见问题

### 1. 性能问题
问题: 日志处理延迟高
解决:
- 检查队列大小配置
- 优化批量处理参数
- 确认磁盘性能
- 监控系统资源

### 2. 内存问题
问题: 内存使用过高
解决:
- 使用对象池
- 控制队列大小
- 及时清理日志
- 监控内存使用

### 3. ELK问题
问题: 写入ES失败
解决:
- 检查网络连接
- 验证认证信息
- 确认集群状态
- 查看错误日志

## 开发计划

### 1. Kafka集成
目标: 解决直接写入ELK的压力
```
[Service] --> [Kafka] --> [Logstash] --> [Elasticsearch]
```

### 2. Redis集成
目标: 提供限流和缓存支持
```
[Service] --> [Redis] --> [Kafka] --> [Logstash] --> [Elasticsearch]
```

### 3. 监控增强
- Prometheus集成
- Grafana面板
- 告警支持

### 4. 功能增强
- 日志加密
- 字段脱敏
- 日志审计
- 实时分析