# Config 配置管理包

## 简介
Config 包提供了一个统一的配置管理解决方案,支持多种配置源(本地文件、Nacos等),具有配置热更新、配置验证、配置监听等特性。该包设计灵活,易于扩展,可以满足从简单到复杂的各类配置管理需求。

## 主要特性
- 多配置源支持
  - 本地文件 (YAML、JSON等)
  - Nacos 配置中心
  - 可扩展的配置源接口
- 配置热更新
- 配置变更通知
- 配置验证
- 配置值类型转换
- 支持结构体映射
- 线程安全
- 优雅关闭
- 完整的错误处理
- 可观测性支持(日志、指标)

## 快速开始

### 1. 基础使用

```go
// 创建配置管理器
cfg := config.New()
// 添加本地配置源
cfg.AddSource(local.NewSource("config.yaml"))
// 启动配置管理
if err := cfg.Start(context.Background()); err != nil {
panic(err)
}
defer cfg.Stop(context.Background())
// 获取配置值
value := cfg.Get("app.name").String()
```

### 2. 结构体映射

```go
type AppConfig struct {
Name string yaml:"name"
Version string yaml:"version"
Port int yaml:"port"
}
var appCfg AppConfig
if err := cfg.Scan(&appCfg); err != nil {
panic(err)
}
```

### 3. 配置热更新

```go
// 监听配置变更
watcher := cfg.Watch(context.Background())
defer watcher.Stop()
go func() {
for {
select {
case event := <-watcher.Next():
// 处理配置变更
fmt.Printf("配置变更: %v\n", event)
}
}
}()
```

### 4. Nacos配置源

```go
// Nacos配置
nacosOpts := nacos.Options{
Address: "127.0.0.1:8848",
Username: "nacos",
Password: "nacos",
Group: "DEFAULT_GROUP",
DataID: "application.yaml",
Namespace: "public",
}
// 创建Nacos配置源
nacosSource := nacos.NewSource(nacosOpts)
// 添加到配置管理器
cfg.AddSource(nacosSource)
```

## 高级特性

### 1. 配置验证

```go
type DatabaseConfig struct {
Host string yaml:"host" validate:"required"
Port int yaml:"port" validate:"required,min=1,max=65535"
Username string yaml:"username" validate:"required"
Password string yaml:"password" validate:"required"
}
var dbCfg DatabaseConfig
if err := cfg.Scan(&dbCfg); err != nil {
panic(err)
}
```

### 2. 自定义配置源

```go
type CustomSource struct {
// 实现配置源接口
}
func (s CustomSource) Load(ctx context.Context) (provider.KeyValue, error) {
// 实现配置加载逻辑
}
func (s CustomSource) Watch(ctx context.Context) (provider.Watcher, error) {
// 实现配置监听逻辑
}
```

### 3. 配置优先级

```go
// 高优先级配置会覆盖低优先级配置
cfg.AddSource(local.NewSource("config.yaml"), provider.WithPriority(1))
cfg.AddSource(nacos.NewSource(nacosOpts), provider.WithPriority(2))
```

### 4. 配置加密

```go
// 创建加密配置源
encryptedSource := provider.NewEncryptedSource(
local.NewSource("config.yaml"),
crypto.NewAESCrypto(key),
)
cfg.AddSource(encryptedSource)
```

## 配置项说明

### 本地配置源选项

```go
type LocalOptions struct {
Path string // 配置文件路径
WatchInterval time.Duration // 文件监听间隔
Parser Parser // 自定义解析器
}

### Nacos配置源选项

go
type NacosOptions struct {
Address string // Nacos服务地址
Username string // 用户名
Password string // 密码
Namespace string // 命名空间
Group string // 配置分组
DataID string // 配置ID
Timeout time.Duration // 超时时间
RetryCount int // 重试次数
RetryBackoff time.Duration // 重试间隔
}

## 错误处理
包中定义了一系列错误类型:
- ConfigNotFoundError: 配置不存在
- ConfigInvalidError: 配置无效
- ConfigLoadError: 配置加载失败
- ConfigParseError: 配置解析失败
- ConfigWatchError: 配置监听失败
- ConfigValidateError: 配置验证失败
```

## 最佳实践

### 1. 配置分层

```yaml
#应用配置
app:
name: myapp
version: 1.0.0
#服务配置
server:
port: 8080
timeout: 30s
#数据库配置
database:
host: localhost
port: 3306
```

### 2. 环境配置分离

```go
// 根据环境加载不同配置
env := os.Getenv("ENV")
cfg.AddSource(local.NewSource(fmt.Sprintf("config.%s.yaml", env)))
```

### 3. 敏感信息处理

```go
// 使用环境变量覆盖敏感配置
cfg.AddSource(env.NewSource([]string{
"DATABASE_PASSWORD",
"API_KEY",
}))
```

### 4. 优雅关闭

```go
ctx, cancel := context.WithTimeout(context.Background(), 5time.Second)
defer cancel()
if err := cfg.Stop(ctx); err != nil {
log.Printf("配置管理器关闭失败: %v", err)
}
```

## 性能优化

### 1. 缓存优化
- 配置值会被缓存以提高读取性能
- 只有在配置发生变化时才会更新缓存
- 支持配置缓存大小限制

### 2. 并发控制
- 读操作无锁
- 写操作使用互斥锁
- 支持读写分离

### 3. 内存优化
- 使用对象池减少内存分配
- 支持配置压缩
- 可设置最大配置大小

## 监控指标

### 1. 基础指标
- 配置加载次数
- 配置更新次数
- 配置错误次数
- 配置大小

### 2. 性能指标
- 配置加载耗时
- 配置解析耗时
- 配置验证耗时
- 配置监听延迟

## 常见问题

### 1. 配置无法加载
- 检查配置文件路径是否正确
- 验证配置文件格式
- 确认配置源连接是否正常
- 查看详细错误日志

### 2. 配置更新不生效
- 确认配置监听是否启用
- 检查配置优先级设置
- 验证配置格式是否正确
- 确认配置变更通知是否正常

### 3. 性能问题
- 减少配置文件大小
- 优化配置更新频率
- 使用合适的缓存策略
- 控制配置监听器数量