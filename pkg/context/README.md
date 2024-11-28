# Context 模块

## 简介
Context 模块是一个基于 Go 标准库 `context.Context` 的扩展实现，提供了丰富的上下文信息管理功能。该模块设计用于在请求处理过程中传递和管理各种上下文信息，支持并发安全的元数据操作，并与 Gin 框架无缝集成。

## 目录结构

```
pkg/context/
├── types/ # 类型定义
│ ├── context.go # Context 接口定义
│ └── constants.go # 常量定义
├── metadata.go # 基础 Context 实现
├── values.go # 值类型转换工具
├── validator.go # 验证器实现
└── metadata_test.go # 单元测试
```

## 核心功能

### 1. Context 接口
继承自 `context.Context`，扩展了以下功能：
- 元数据管理
- 用户信息管理
- 请求信息管理
- 追踪信息管理
- 超时控制
- 错误处理
- 上下文克隆

### 2. 预定义常量

```
go

const (
KeyUserID = "user_id" // 用户ID
KeyUserName = "user_name" // 用户名
KeyRequestID = "request_id" // 请求ID
KeyClientIP = "client_ip" // 客户端IP
KeyTraceID = "trace_id" // 追踪ID
KeySpanID = "span_id" // SpanID
KeyError = "error" // 错误信息
)
```

## API 参考

### 创建上下文

```
go

// 创建新的上下文
ctx := context.NewContext(context.Background())
```

### 元数据操作

```
go

// 设置单个元数据
ctx.SetMetadata("key", value)
// 获取单个元数据
value, exists := ctx.GetMetadata("key")
// 获取所有元数据（返回副本）
metadataMap := ctx.Metadata()
// 删除元数据
ctx.DeleteMetadata("key")
// 批量设置元数据
data := map[string]interface{}{
"key1": "value1",
"key2": "value2",
}
ctx.SetMetadataMap(data)
```

### 用户信息管理
```
go

// 设置用户信息
ctx.SetUserID("user-123")
ctx.SetUserName("John Doe")
// 获取用户信息
userID := ctx.GetUserID()
userName := ctx.GetUserName()
```

### 请求信息管理
```
go

/ 设置请求信息
ctx.SetRequestID("req-123")
ctx.SetClientIP("192.168.1.1")
// 获取请求信息
requestID := ctx.GetRequestID()
clientIP := ctx.GetClientIP()
```

### 追踪信息管理
```
go

// 设置追踪信息
ctx.SetTraceID("trace-123")
ctx.SetSpanID("span-456")
// 获取追踪信息
traceID := ctx.GetTraceID()
spanID := ctx.GetSpanID()
```

### 错误处理
```
go

// 设置错误
ctx.SetError(errors.New("some error"))
// 获取错误
err := ctx.GetError()
// 检查是否有错误
hasError := ctx.HasError()
// 清除错误
ctx.SetError(nil)
```

### 上下文控制
```
go

// 设置超时
timeoutCtx, cancel := ctx.WithTimeout(5 time.Second)
defer cancel()
// 设置截止时间
deadline := time.Now().Add(10 time.Second)
deadlineCtx, cancel := ctx.WithDeadline(deadline)
defer cancel()
// 创建可取消的上下文
cancelCtx, cancel := ctx.WithCancel()
defer cancel()
// 克隆上下文
clonedCtx := ctx.Clone()
```

### 类型转换工具
```
go

// 获取字符串值
strVal, ok := GetStringValue(ctx, "string-key")
// 获取整数值
intVal, ok := GetIntValue(ctx, "int-key")
// 获取布尔值
boolVal, ok := GetBoolValue(ctx, "bool-key")
// 获取浮点数值
floatVal, ok := GetFloat64Value(ctx, "float-key")
// 获取时间值
timeVal, ok := GetTimeValue(ctx, "time-key")
```

### 验证器
```
go

// 预定义的验证规则集
var (
RequiredUserContext = []string{types.KeyUserID, types.KeyUserName}
RequiredTraceContext = []string{types.KeyRequestID, types.KeyTraceID}
RequiredBasicContext = []string{types.KeyRequestID, types.KeyClientIP}
)
// 自定义验证
err := ValidateContext(ctx, "key1", "key2")
// 验证用户上下文
err := ValidateUserContext(ctx)
// 验证追踪上下文
err := ValidateTraceContext(ctx)
```

### Gin 框架集成
```
go

// 注册中间件
r := gin.New()
r.Use(ctxMiddleware.Middleware())
// 在路由处理器中获取上下文
r.GET("/api", func(c gin.Context) {
ctx := ctxMiddleware.GetContextFromGin(c)
// 使用上下文...
})
```

## 特性

### 线程安全
- 所有操作都是并发安全的
- 使用 `sync.RWMutex` 保护元数据访问
- 支持并发读写操作

### 类型安全
- 所有类型转换操作都返回成功标志
- 预定义的类型转换函数避免运行时错误

### 内存安全
- 元数据操作返回深拷贝，避免外部修改
- Clone 操作创建完整的上下文副本

## 测试覆盖
- 完整的单元测试套件
- 并发安全性测试
- 边界条件测试
- 类型转换测试
- 超时和取消测试

## 最佳实践
1. 使用预定义常量而不是硬编码字符串
2. 使用 defer cancel() 确保资源正确释放
3. 总是检查类型转换的成功标志
4. 使用验证器确保必要字段存在
5. 通过 Clone() 创建上下文副本
6. 在中间件中统一注入基础信息

## 依赖
- Go 1.x
- github.com/gin-gonic/gin

## 限制
1. 元数据存储在内存中，不支持持久化
2. 类型转换仅支持基本数据类型
3. 验证器仅检查字段存在性，不验证值的有效性

## 错误处理
所有可能的错误都通过返回值显式处理，包括：
- 类型转换失败
- 验证失败
- 上下文取消或超时


## 待完成

### 日志集成
- 需要等待日志模块开发完成 (Logrus + ELK)
- 需要定义统一的日志接口和级别
- 需要实现日志上下文传递

### 上下文链路追踪增强
- 需要等待链路追踪系统选型和实现（Jaeger）
- 需要实现分布式追踪的完整功能
- 需要考虑跨服务调用的追踪信息传递

### 性能监控
- 需要等待监控系统的开发和集成 (Prometheus + Grafana)
- 需要定义统一的指标收集接口
- 需要实现指标的聚合和上报

### 上下文池化
- 需要等待性能测试结果
- 需要评估内存使用情况
- 需要在完整系统中测试池化效果

### 上下文验证器增强
- 需要等待业务逻辑的具体需求
- 需要根据实际使用场景扩展验证规则
- 需要考虑与其他验证系统的集成

### 中间件扩展
- 需要等待具体的业务需求
- 需要与其他中间件协调工作
- 需要考虑性能和资源消耗