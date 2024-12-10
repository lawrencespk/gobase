# Recovery 中间件

## 简介
Recovery 中间件用于捕获并处理 HTTP 请求处理过程中发生的 panic,确保服务的稳定性。该中间件提供了灵活的配置选项和自定义处理能力。

## 特性
- 自动捕获并恢复 panic
- 支持自定义错误处理
- 支持自定义日志记录
- 支持自定义堆栈跟踪
- 支持自定义响应格式
- 支持上下文传递
- 支持错误码映射

## 快速开始

### 基础用法

```go
import "gobase/pkg/middleware/recovery"
// 使用默认配置
r := gin.New()
r.Use(recovery.Recovery())
```

### 自定义配置

```go
import "gobase/pkg/middleware/recovery"
// 使用自定义配置
import (
"gobase/pkg/middleware/recovery"
"gobase/pkg/logger"
)
// 创建自定义选项
opts := &recovery.Options{
// 自定义日志记录器
Logger: logger.NewLogger(),
// 自定义错误处理函数
ErrorHandler: func(ctx context.Context, err interface{}) error {
return errors.NewSystemError("系统内部错误", err)
},
// 自定义响应处理函数
ResponseHandler: func(c gin.Context, err error) {
c.JSON(http.StatusInternalServerError, gin.H{
"code": "500",
"message": err.Error(),
})
},
// 是否打印堆栈信息
PrintStack: true,
// 堆栈信息深度
StackDepth: 64,
// 是否记录请求信息
LogRequest: true,
// 是否记录请求头
LogHeaders: true,
}
// 使用自定义配置
r.Use(recovery.RecoveryWithOptions(opts))
```

### 与日志系统集成

```go
import (
"gobase/pkg/logger"
"gobase/pkg/logger/logrus"
)
// 创建日志记录器
logger := logrus.NewLogger(...)
// 配置Recovery中间件
opts := &recovery.Options{
Logger: logger,
LogRequest: true,
LogHeaders: true,
}
r.Use(recovery.RecoveryWithOptions(opts))
```

### 与错误处理系统集成

```go
import "gobase/pkg/errors"
opts := &recovery.Options{
ErrorHandler: func(ctx context.Context, err interface{}) error {
// 转换为系统错误
if sysErr, ok := err.(error); ok {
return errors.NewSystemError("系统内部错误", sysErr)
}
// 处理panic值
return errors.NewSystemError("系统内部错误", fmt.Errorf("%v", err))
},
}
```

## 配置选项说明

### Options 结构体

```go
type Options struct {
Logger types.Logger // 日志记录器
ErrorHandler ErrorHandlerFunc // 错误处理函数
ResponseHandler ResponseHandlerFunc // 响应处理函数
PrintStack bool // 是否打印堆栈
StackDepth int // 堆栈深度
LogRequest bool // 是否记录请求信息
LogHeaders bool // 是否记录请求头
}
```

### 默认值

```go
DefaultOptions = Options{
Logger: nil, // 需要手动设置
PrintStack: true,
StackDepth: 64,
LogRequest: true,
LogHeaders: false,
```

## 最佳实践

### 1. 错误处理

```go
opts := &recovery.Options{
ErrorHandler: func(ctx context.Context, err interface{}) error {
// 1. 区分不同类型的panic
switch e := err.(type) {
case error:
return errors.NewSystemError("系统错误", e)
case string:
return errors.NewSystemError(e, nil)
default:
return errors.NewSystemError("未知错误", fmt.Errorf("%v", e))
}
},
}
```

### 2. 日志记录

```go
opts := &recovery.Options{
Logger: logger,
LogRequest: true,
LogHeaders: true,
PrintStack: true,
}
```
### 3. 响应处理

```go
opts := &recovery.Options{
ResponseHandler: func(c gin.Context, err error) {
// 1. 获取错误码
code := errors.GetErrorCode(err)
// 2. 获取HTTP状态码
status := errors.GetHTTPStatus(err)
// 3. 返回响应
c.JSON(status, gin.H{
"code": code,
"message": err.Error(),
})
},
}
```

## 注意事项

1. 中间件顺序
   - Recovery 中间件应该注册在所有其他中间件之前
   - 确保能够捕获所有可能的 panic

2. 错误处理
   - 建议在 ErrorHandler 中统一处理所有类型的 panic
   - 将 panic 转换为适当的错误类型
   - 避免在 ErrorHandler 中再次触发 panic

3. 日志记录
   - 建议开启 LogRequest 以便问题排查
   - 生产环境建议关闭 LogHeaders 以保护敏感信息
   - 根据环境配置适当的堆栈深度

4. 性能考虑
   - PrintStack 会带来一定的性能开销
   - LogHeaders 会增加日志存储量
   - 根据实际需求平衡功能和性能

## 常见问题

1. Q: 如何处理自定义panic值?
   A: 在 ErrorHandler 中进行类型断言和转换

2. Q: 如何避免敏感信息泄露?
   A: 配置 LogHeaders=false 并在 ResponseHandler 中过滤敏感信息

3. Q: 如何调试panic问题?
   A: 开启 PrintStack 和 LogRequest,设置合适的 StackDepth

4. Q: 如何自定义错误响应格式?
   A: 实现自定义 ResponseHandler 函数

## 版本历史

- v1.0.0
  - 初始版本
  - 基础panic恢复功能
  - 可配置的日志记录
  - 自定义错误处理

## 后续规划

1. 支持更多的日志格式
2. 增加性能统计功能
3. 支持错误报警功能
4. 支持自定义过滤规则