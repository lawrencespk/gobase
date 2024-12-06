# 日志系统使用文档

## 目录
- [简介](#简介)
- [功能特性](#功能特性)
- [快速开始](#快速开始)
- [基础使用](#基础使用)
- [高级特性](#高级特性)
- [配置说明](#配置说明)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)

## 简介
本日志系统基于 logrus 构建，并集成了 ELK Stack，提供了完整的日志收集、存储、查询解决方案。

## 功能特性
### 基础功能
- 多级别日志记录(Debug, Info, Warn, Error, Fatal)
- 结构化日志输出
- 自定义字段支持
- 调用者信息记录
- 异步日志写入
- 日志文件轮转
- 日志压缩存储
- 自动清理过期日志

### 高级特性
- ELK 集成
- 批量日志处理
- 错误重试机制
- 优雅关闭
- 性能优化
- 自定义格式化

## 快速开始
```go
package main
import (
"gobase/pkg/logger/logrus"
"gobase/pkg/logger/elk"
)
func main() {
// 1. 创建基础配置
opts := logrus.DefaultOptions()
opts.OutputPaths = []string{"stdout", "./logs/app.log"}
// 2. 创建文件管理器
fm, err := logrus.NewFileManager(logrus.FileOptions{
BufferSize: 32 1024,
MaxOpenFiles: 100,
DefaultPath: "./logs/app.log",
})
if err != nil {
panic(err)
}
// 3. 初始化日志器
logger, err := logrus.NewLogger(fm, logrus.QueueConfig{
MaxSize: 1000,
BatchSize: 100,
Workers: 1,
}, opts)
if err != nil {
panic(err)
}
defer logger.Close()
// 4. 记录日志
logger.Info("应用启动成功")
}
```
## 基础使用

### 日志级别
```go
logger.Debug("调试信息")
logger.Info("普通信息")
logger.Warn("警告信息")
logger.Error("错误信息")
logger.Fatal("致命错误") // 会导致程序退出
```
### 添加字段
```go
logger.WithFields([]types.Field{
{Key: "user_id", Value: "12345"},
{Key: "action", Value: "login"},
}).Info("用户登录")
```
### 异步写入
```go
opts := &logrus.Options{
AsyncConfig: logrus.AsyncConfig{
Enable: true,
BufferSize: 8192,
FlushInterval: time.Second,
},
}
```
## 高级特性

### ELK 集成
```go
elkConfig := elk.DefaultElkConfig()
elkConfig.Addresses = []string{"http://elasticsearch:9200"}
hookOpts := elk.ElkHookOptions{
Config: elkConfig,
Index: "app-logs",
BatchConfig: &elk.BulkProcessorConfig{
BatchSize: 100,
FlushInterval: time.Second,
},
}
hook, err := elk.NewElkHook(hookOpts)
if err != nil {
panic(err)
}
logger.AddHook(hook)
```
### ELK 集成详细配置
```go
// 1. 创建 ELK 配置
elkConfig := elk.DefaultElkConfig()
elkConfig.Addresses = []string{"http://elasticsearch:9200"}
elkConfig.Username = "elastic"
elkConfig.Password = "password"
elkConfig.Index = "my-app-logs"
elkConfig.Timeout = 30 * time.Second

// 2. 配置 Hook 选项
hookOpts := elk.ElkHookOptions{
    Config: elkConfig,
    Levels: []logrus.Level{
        logrus.InfoLevel,
        logrus.WarnLevel,
        logrus.ErrorLevel,
    },
    Index: "my-app-logs",
    BatchConfig: &elk.BulkProcessorConfig{
        BatchSize: 100,
        FlushBytes: 5 * 1024 * 1024, // 5MB
        Interval: time.Second,
        RetryCount: 3,
        RetryWait: time.Second,
        CloseTimeout: 10 * time.Second,
    },
    MaxDocSize: 5 * 1024 * 1024, // 5MB
    IndexPrefix: "logs",
    IndexSuffix: time.Now().Format("2006.01.02"),
}

// 3. 创建并添加 Hook
hook, err := elk.NewElkHook(hookOpts)
if err != nil {
    panic(err)
}
logger.AddHook(hook)
```

### 自定义格式化
```go
// 1. 创建自定义格式化器
formatter := &logrus.Formatter{
    TimestampFormat: time.RFC3339,
    PrettyPrint: true,
}

// 2. 设置格式化器
logger.SetFormatter(formatter)
```

### 错误处理与重试机制
```go
// 1. 配置重试
retryConfig := elk.RetryConfig{
    MaxRetries: 3,
    InitialWait: time.Second,
    MaxWait: 5 * time.Second,
}

// 2. 使用重试机制
err := elk.WithRetry(ctx, retryConfig, func() error {
    return logger.Info("重要消息")
}, logger)
```

### 批量处理配置
```go
// 1. 创建批处理配置
bulkConfig := elk.BulkProcessorConfig{
    BatchSize: 1000,
    FlushBytes: 5 * 1024 * 1024,
    Interval: time.Second,
    DefaultIndex: "my-app-logs",
    RetryCount: 3,
    RetryWait: time.Second,
    CloseTimeout: 10 * time.Second,
}

// 2. 创建批处理器
processor, err := elk.NewBulkProcessor(client, &bulkConfig)
if err != nil {
    panic(err)
}
defer processor.Close()
```

### 索引管理
```go
// 1. 创建索引映射
mapping := elk.DefaultIndexMapping()
mapping.Settings["number_of_shards"] = 3
mapping.Settings["number_of_replicas"] = 2

// 2. 创建索引
err := client.CreateIndex(ctx, "my-app-logs", mapping)

// 3. 检查索引是否存在
exists, err := client.IndexExists(ctx, "my-app-logs")

// 4. 删除索引
err := client.DeleteIndex(ctx, "my-app-logs")
```

### 缓冲池管理
```go
// 1. 创建缓冲池
bufferPool := logrus.NewBufferPool()

// 2. 创建写入池
writePool := logrus.NewWritePool(32 * 1024) // 32KB buffer size
```

### 压缩配置
```go
compressConfig := logrus.CompressConfig{
    Enable: true,
    Algorithm: "gzip",
    Level: gzip.BestCompression,
    DeleteSource: true,
    Interval: time.Hour,
    LogPaths: []string{"./logs"},
}
```

### 清理配置
```go
cleanupConfig := logrus.CleanupConfig{
    Enable: true,
    MaxBackups: 7,
    MaxAge: 30,
    Interval: 24 * time.Hour,
    LogPaths: []string{"./logs"},
}
```

## 配置说明

### 文件管理配置
```go
type FileOptions struct {
BufferSize int // 写入缓冲区大小
FlushInterval time.Duration // 刷新间隔
MaxOpenFiles int // 最大打开文件数
DefaultPath string // 默认日志文件路径
}
```
### 异步配置
```go
type AsyncConfig struct {
Enable bool // 是否启用异步写入
BufferSize int // 缓冲区大小
FlushInterval time.Duration // 定期刷新间隔
BlockOnFull bool // 缓冲区满时是否阻塞
DropOnFull bool // 缓冲区满时是否丢弃
FlushOnExit bool // 退出时是否刷新缓冲区
}
```
### ELK 配置
```go
type ElkConfig struct {
Addresses []string // ES 服务器地址
Username string // 用户名
Password string // 密码
Index string // 索引名称
Timeout time.Duration // 超时时间
}
```




## 最佳实践

### 性能优化建议
1. 日志级别使用建议
   - DEBUG: 开发环境使用，用于调试
   - INFO: 记录正常的业务流程
   - WARN: 记录需要注意但不影响系统运行的问题
   - ERROR: 记录影响当前操作的错误
   - FATAL: 记录需要立即处理的严重问题

2. 性能优化
   - 生产环境建议启用异步写入
   - 合理配置批处理大小和刷新间隔
   - 定期清理过期日志
   - 使用结构化的日志格式

3. 合理配置批处理参数
   - BatchSize: 根据日志量设置,通常 100-1000
   - FlushBytes: 根据内存使用情况设置,通常 1-5MB
   - Interval: 根据实时性要求设置,通常 1-5秒

4. 内存管理
   - 使用缓冲池减少内存分配
   - 合理设置单个文档大小限制
   - 定期清理过期日志

5. 错误处理
   - 实现自定义错误处理器
   - 合理配置重试机制
   - 记录错误统计信息

6. ELK 集成优化
   - 使用批量写入
   - 合理设置索引分片
   - 定期优化索引

7. ELK 集成建议
   - 合理设置批处理参数
   - 配置错误重试机制
   - 监控日志队列大小
   - 定期检查 ES 连接状态

### 监控指标
1. 日志统计
   - 总处理文档数
   - 总处理字节数
   - 刷新次数
   - 错误次数
   - 最后错误信息
   - 最后刷新时间

2. 性能指标
   - 写入队列大小
   - 批处理延迟
   - 写入成功率
   - 重试次数

## 常见问题

1. 日志丢失问题
   - 检查异步配置
   - 验证队列大小设置
   - 确认刷新间隔合理性
   - 查看错误处理配置

2. 性能问题
   - 调整批处理参数
   - 优化索引配置
   - 使用缓冲池
   - 合理设置压缩级别

3. ELK 连接问题
   - 验证网络连接
   - 检查认证信息
   - 确认集群状态
   - 查看重试配置


## 常见问题

1. 日志写入失败
   - 检查文件权限
   - 确认磁盘空间
   - 查看文件描述符限制

2. ELK 连接问题
   - 验证 ES 服务可用性
   - 检查网络连接
   - 确认认证信息正确

3. 性能问题
   - 调整批处理参数
   - 优化日志级别
   - 检查磁盘 IO
   - 监控内存使用

# 待完成

1. Kafka 集成, 消除直接写入ELK, 降低写入压力, 同时支持数据持久化, 避免丢失数据
2. 集成Redis, 提升性能\
计划流程

[Service] --> [Redis Rate Limiter] --> [Kafka] --> [Logstash] --> [Elasticsearch]
 ---------------------↑ 限流控制------------------↑ 消息队列--↑ 批量处理

