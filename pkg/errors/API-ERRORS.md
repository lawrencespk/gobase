# 错误处理模块 API 参考

## 1. 错误创建接口

### 1.1 业务错误

```
go

// 参数相关
NewInvalidParamsError(message string, cause error) error
NewBadRequestError(message string, cause error) error
NewValidationError(message string, cause error) error
// 认证授权
NewUnauthorizedError(message string, cause error) error
NewForbiddenError(message string, cause error) error
NewInvalidTokenError(message string, cause error) error
NewTokenExpiredError(message string, cause error) error
// 资源相关
NewNotFoundError(message string, cause error) error
NewAlreadyExistsError(message string, cause error) error
NewResourceExhaustedError(message string, cause error) error
// 数据相关
NewDataAccessError(message string, cause error) error
NewDataCreateError(message string, cause error) error
NewDataUpdateError(message string, cause error) error
NewDataDeleteError(message string, cause error) error
NewDataQueryError(message string, cause error) error
NewDataConvertError(message string, cause error) error
NewDataValidateError(message string, cause error) error
NewDataCorruptError(message string, cause error) error
```

### 1.2 系统错误

```
go

NewSystemError(message string, cause error) error
NewConfigError(message string, cause error) error
NewNetworkError(message string, cause error) error
NewDatabaseError(message string, cause error) error
NewCacheError(message string, cause error) error
NewTimeoutError(message string, cause error) error
NewThirdPartyError(message string, cause error) error
NewInitializeError(message string, cause error) error
NewShutdownError(message string, cause error) error
NewMemoryError(message string, cause error) error
NewDiskError(message string, cause error) error
```

### 1.3 数据库错误

```
go

NewDBConnError(message string, cause error) error
NewDBQueryError(message string, cause error) error
NewDBTransactionError(message string, cause error) error
NewDBDeadlockError(message string, cause error) error
```

### 1.4 缓存错误

```
go

NewCacheMissError(message string, cause error) error
NewCacheExpiredError(message string, cause error) error
NewCacheFullError(message string, cause error) error
```

## 2. 错误包装接口

### 2.1 基础包装

```
go

// 基础包装
Wrap(err error, message string) error
WrapWithCode(err error, code string, message string) error
// 格式化包装
Wrapf(err error, format string, args ...interface{}) error
WrapWithCodef(err error, code string, format string, args ...interface{}) error
```

## 3. 错误检查接口

### 3.1 错误类型检查

```
go

// 错误类型判断
IsBusinessError(err error) bool
IsSystemError(err error) bool
IsRetryableError(err error) bool
// 错误码检查
HasErrorCode(err error, code string) bool
IsErrorType(err error, startCode, endCode string) bool
```

### 3.2 错误信息获取

```
go

// 获取错误信息
GetErrorCode(err error) string
GetErrorMessage(err error) string
GetErrorDetails(err error) []interface{}
GetErrorStack(err error) []string
FormatErrorChain(err error) string
```
## 4. 错误重试接口

### 4.1 重试配置
```
go

type RetryConfig struct {
MaxAttempts int // 最大重试次数
InitialInterval time.Duration // 初始重试间隔
MaxInterval time.Duration // 最大重试间隔
Multiplier float64 // 重试间隔增长因子
}
// 默认配置
var DefaultRetryConfig = RetryConfig{
MaxAttempts: 3,
InitialInterval: 100 time.Millisecond,
MaxInterval: 2 time.Second,
Multiplier: 2.0,
}
```
### 4.2 重试执行
```
go

// 重试函数类型
type Retryable func(ctx context.Context) error
// 执行重试
WithRetry(ctx context.Context, op Retryable, opts ...RetryConfig) error
```
## 5. 错误分组接口

### 5.1 分组创建和管理
```
go

// 创建错误组
NewErrorGroup() Group
// Group 方法
type Group interface {
Add(err error) // 添加错误
HasErrors() bool // 检查是否有错误
GetErrors() []error // 获取所有错误
Error() string // 实现 error 接口
Go(ctx context.Context, fns ...func(ctx context.Context) error) // 并发执行
Clear() // 清除所有错误
First() error // 获取第一个错误
Len() int // 获取错误数量
}
```
## 6. HTTP 错误处理接口

### 6.1 错误转换
```
go

// 错误响应结构
type ErrorResponse struct {
Code string json:"code"
Message string json:"message"
Details interface{} json:"details,omitempty"
}
// 转换方法
ToHTTPResponse(err error) (int, ErrorResponse)
GetHTTPStatus(err error) int
```
### 6.2 中间件配置
```
go

// 中间件选项
type ErrorHandlerOptions struct {
IncludeStack bool
IncludeDetails bool
EnableLogging bool
ResponseFormatter func(c gin.Context, err error) interface{}
}
// 创建中间件
ErrorHandler(opts ...ErrorHandlerOptions) gin.HandlerFunc

```
## 7. 错误链接口

### 7.1 错误链操作
```
go

// 获取错误链
GetErrorChain(err error) ErrorChain
FirstError(err error) types.Error
LastError(err error) types.Error
RootCause(err error) error
```
## 8. 使用示例

### 8.1 基础错误处理
```
go

// 创建错误
err := NewInvalidParamsError("用户ID不能为空", nil)
// 包装错误
err = Wrap(err, "验证用户信息失败")
// 检查错误类型
if IsBusinessError(err) {
// 处理业务错误
}
```
### 8.2 重试示例
```
go

err := WithRetry(ctx, func(ctx context.Context) error {
return someOperation()
}, RetryConfig{
MaxAttempts: 3,
InitialInterval: time.Second,
})
```
### 8.3 错误分组示例

```
go

group := NewErrorGroup()
group.Go(ctx,
func(ctx context.Context) error {
return task1()
},
func(ctx context.Context) error {
return task2()
},
)
if group.HasErrors() {
errors := group.GetErrors()
// 处理错误
}
```


























