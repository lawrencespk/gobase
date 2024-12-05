# 错误处理系统使用文档

## 目录
- [1. 基础错误创建](#1-基础错误创建)
- [2. 错误包装与链式处理](#2-错误包装与链式处理)
- [3. 错误组处理](#3-错误组处理)
- [4. 重试机制](#4-重试机制)
- [5. Web 集成](#5-web-集成)
- [6. 错误信息获取](#6-错误信息获取)
- [7. 特定场景错误处理](#7-特定场景错误处理)
- [8. 最佳实践](#8-最佳实践)

## 1. 基础错误创建

### 1.1 系统错误
```go
// 系统内部错误
err := NewSystemError("系统内部错误", cause)
// 配置错误
err := NewConfigError("配置文件无效", cause)
// 网络错误
err := NewNetworkError("网络连接失败", cause)
// 数据库错误
err := NewDatabaseError("数据库连接失败", cause)
```

### 1.2 业务错误
```go
// 参数验证错误
err := NewInvalidParamsError("用户ID不能为空", nil)
// 未授权错误
err := NewUnauthorizedError("用户未登录", nil)
// 资源不存在
err := NewNotFoundError("用户不存在", nil)
// 权限不足
err := NewPermissionDeniedError("没有删除权限", nil)
```

### 1.3 错误接口说明
```go
// 所有自定义错误都实现了以下接口
type Error interface {
    error           // 标准错误接口
    Code() string   // 获取错误码
    Message() string // 获取错误消息
    Details() []interface{} // 获取错误详情
    Unwrap() error // 获取原始错误（用于错误链）
    Stack() []string // 获取错误堆栈
}

// 示例：获取错误信息
err := NewSystemError("系统错误", cause)
fmt.Println(err.Error())          // 完整错误信息
fmt.Println(err.(Error).Code())   // 错误码
fmt.Println(err.(Error).Message()) // 错误消息
```
## 2. 错误包装与链式处理

### 2.1 基础包装
```go
// 简单包装
err = Wrap(err, "处理用户请求失败")
// 使用特定错误码包装
err = WrapWithCode(err, codes.InvalidParams, "验证用户信息失败")
// 格式化包装
err = Wrapf(err, "处理用户 %s 的请求失败", userID)
```
### 2.2 错误链处理
```go
// 获取完整错误链
chain := GetErrorChain(err)
// 获取第一个错误
firstErr := FirstError(err)
// 获取最后一个错误
lastErr := LastError(err)
// 获取根因
rootErr := RootCause(err)
```
### 2.3 错误链格式化
```go
// 格式化完整错误链
chainStr := FormatErrorChain(err)

// 判断错误类型范围
if IsErrorType(err, "2000", "3000") {
    // 处理业务错误
}
```
## 3. 错误组处理

### 3.1 基础使用
```go
// 创建错误组
group := NewErrorGroup()
// 添加错误
group.Add(err1)
group.Add(err2)
// 检查是否有错误
if group.HasErrors() {
// 处理错误
}
// 获取所有错误
errors := group.GetErrors()
```
### 3.2 并发错误处理
```go
group := NewErrorGroup()
// 并发执行多个任务
group.Go(ctx,
func(ctx context.Context) error {
return task1()
},
func(ctx context.Context) error {
return task2()
},
)
```
### 3.3 错误组高级用法
```go
group := NewErrorGroup()

// 清除所有错误
group.Clear()

// 获取错误数量
count := group.Len()

// 获取第一个错误
firstErr := group.First()
```
## 4. 重试机制

### 4.1 基础重试
```go
// 使用默认配置
err := WithRetry(ctx, func(ctx context.Context) error {
return someOperation()
})
```
### 4.2 自定义重试配置
```go
config := RetryConfig{
MaxAttempts: 5,
InitialInterval: 100 time.Millisecond,
MaxInterval: 2 time.Second,
Multiplier: 2.0,
}
err := WithRetry(ctx, func(ctx context.Context) error {
return someOperation()
}, config)
```
### 4.3 重试条件控制
```go
// 判断错误是否可重试
if IsRetryableError(err) {
    // 进行重试
}

// 自定义重试配置
config := RetryConfig{
    MaxAttempts:     5,
    InitialInterval: 100 * time.Millisecond,
    MaxInterval:     2 * time.Second,
    Multiplier:      2.0,
}
```
## 5. Web 集成

### 5.1 错误处理中间件
```go
// 使用默认配置
router.Use(ErrorHandler())
// 使用自定义配置
router.Use(ErrorHandler(&ErrorHandlerOptions{
IncludeStack: true,
IncludeDetails: true,
EnableLogging: true,
}))
```
### 5.2 手动错误处理
```go
// 在 Gin 处理函数中
func Handler(c gin.Context) {
if err := someOperation(); err != nil {
HandleError(c, err)
return
}
}
```
### 5.3 错误转换
```go
// 获取HTTP状态码
status := GetHTTPStatus(err)
// 转换为HTTP响应
status, response := ToHTTPResponse(err)
// 自定义响应格式
type ErrorResponse struct {
Code string json:"code"
Message string json:"message"
Details interface{} json:"details,omitempty"
}
```
### 5.4 中间件配置
```go
// 自定义中间件选项
options := &ErrorHandlerOptions{
    IncludeStack:   true,
    IncludeDetails: true,
    EnableLogging:  true,
    ResponseFormatter: func(c *gin.Context, err error) interface{} {
        // 自定义响应格式
        return customFormat(err)
    },
}

// 使用自定义选项
router.Use(ErrorHandler(options))
```
## 6. 错误信息获取

### 6.1 错误详情获取
```go
// 获取错误码
code := GetErrorCode(err)
// 获取错误消息
message := GetErrorMessage(err)
// 获取错误详情
details := GetErrorDetails(err)
// 获取错误堆栈
stack := GetErrorStack(err)
```
### 6.2 错误类型判断
```go
// 判断是否为业务错误
if IsBusinessError(err) {
// 处理业务错误
}
// 判断是否为系统错误
if IsSystemError(err) {
// 处理系统错误
}
// 判断是否包含特定错误码
if HasErrorCode(err, codes.InvalidParams) {
// 处理特定错误
}
```
### 6.3 错误类型判断进阶
```go
// 错误类型范围判断
if IsErrorType(err, "2100", "2200") {
    // 处理用户相关错误
}

// 判断是否包含特定错误
if Is(err, targetErr) {
    // 处理特定错误
}

// 错误类型转换
var customErr *CustomError
if As(err, &customErr) {
    // 使用自定义错误类型
}
```
### 6.4 错误类型断言与接口
```go
// 错误类型断言方式一：使用类型断言接口
if customErr, ok := err.(interface{ Code() string }); ok {
    code := customErr.Code()
    // 使用错误码
}

// 错误类型断言方式二：使用 errors 包提供的工具函数
code := errors.GetErrorCode(err)       // 获取错误码
message := errors.GetErrorMessage(err) // 获取错误信息

// 错误类型断言方式三：使用完整接口
type Error interface {
    error
    Code() string
    Message() string
    Details() []interface{}
    Unwrap() error
    Stack() []string
}

if customErr, ok := err.(Error); ok {
    code := customErr.Code()
    message := customErr.Message()
    details := customErr.Details()
    cause := customErr.Unwrap()
    stack := customErr.Stack()
}

// 在测试中使用类型断言
func TestError(t *testing.T) {
    err := NewSystemError("test error", nil)
    
    // 方式一：直接断言接口方法
    assert.Equal(t, codes.SystemError, err.(interface{ Code() string }).Code())
    
    // 方式二：使用工具函数（推荐）
    assert.Equal(t, codes.SystemError, errors.GetErrorCode(err))
}
```
## 7. 特定场景错误处理

### 7.1 数据库操作
```go
// 数据库连接错误
err := NewDBConnError("无法连接到数据库", cause)
// 查询错误
err := NewDBQueryError("查询失败", cause)
// 事务错误
err := NewDBTransactionError("事务提交失败", cause)
```
### 7.2 缓存操作
```go
// 缓存未命中
err := NewCacheMissError("数据不在缓存中", nil)
// 缓存过期
err := NewCacheExpiredError("缓存数据已过期", nil)
// 缓存已满
err := NewCacheFullError("缓存空间已满", nil)
```
### 7.3 文件操作
```go
// 文件不存在
err := NewFileNotFoundError("文件不存在", nil)
// 文件上传错误
err := NewFileUploadError("文件上传失败", cause)
// 文件类型错误
err := NewInvalidFileTypeError("不支持的文件类型", nil)
```
## 8. 最佳实践

### 8.1 错误创建
```go
// 推荐：使用专门的构造函数
err := NewDataAccessError("数据访问失败", cause)
// 不推荐：直接使用 NewError
err := NewError(codes.DataAccessError, "数据访问失败", cause)
```
### 8.2 错误包装
```go
// 推荐：添加有意义的上下文信息
err = Wrap(err, "处理用户登录请求时发生错误")
// 不推荐：简单重复错误信息
err = Wrap(err, "错误")
```
### 8.3 错误处理
```go
// 推荐：详细的错误处理
if err != nil {
if IsBusinessError(err) {
// 处理业务错误
HandleError(c, err)
} else {
// 处理系统错误
log.Error("系统错误", err)
HandleError(c, NewSystemError("系统内部错误", err))
}
return
}
```
## 附录：错误码列表

### 系统错误 (1000-1099)
- 1000: 系统内部错误
- 1001: 配置错误
- 1002: 网络错误
- ...

### 中间件错误 (1100-1199)
- 1100: 中间件错误
- 1101: 认证中间件错误
- 1102: 限流中间件错误
- ...

### 业务错误 (2000-2099)
- 2000: 无效参数
- 2001: 未授权
- 2002: 禁止访问
- ...

[完整错误码列表请参考 codes.go 文件](../errors/codes/codes.go)