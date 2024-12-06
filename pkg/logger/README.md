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

3. ELK 集成建议
   - 合理设置批处理参数
   - 配置错误重试机制
   - 监控日志队列大小
   - 定期检查 ES 连接状态

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