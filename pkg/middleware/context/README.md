# Context 中间件

## 简介
Context中间件提供了一个统一的上下文管理解决方案,用于在HTTP请求处理过程中传递请求相关的上下文信息。主要功能包括:

- 请求ID的自动生成与传递
- 客户端IP的获取与存储
- 元数据的管理(设置、获取、删除)
- 自定义上下文的注入与获取

## 功能特性

### 1. 请求ID管理
- 自动生成唯一的请求ID
- 支持从请求头获取已有请求ID
- 可配置是否在响应头中返回请求ID
- 可自定义请求ID的响应头名称

### 2. 上下文管理
- 提供统一的上下文获取接口
- 支持上下文数据的存取
- 支持元数据的管理
- 与标准context.Context完全兼容

### 3. 配置选项
- 支持自定义请求ID生成器配置
- 支持自定义响应头配置
- 提供默认配置支持

## 快速开始

### 基础用法
```go
import (
    "github.com/gin-gonic/gin"
    "gobase/pkg/middleware/context"
)

func main() {
    r := gin.New()
    
    // 使用默认配置
    r.Use(context.Middleware(nil))
    
    r.GET("/example", func(c *gin.Context) {
        // 获取上下��
        ctx := context.GetContextFromGin(c)
        
        // 获取请求ID
        requestID := ctx.GetRequestID()
        
        // 设置元数据
        ctx.SetMetadata(map[string]interface{}{
            "key": "value",
        })
        
        c.JSON(200, gin.H{"requestID": requestID})
    })
}
```

### 自定义配置
```go
opts := &context.Options{
    RequestIDOptions: &requestid.Options{
        Length: 32,
        Prefix: "REQ-",
    },
    SetRequestIDHeader: true,
    RequestIDHeaderName: "X-Custom-Request-ID",
}

r.Use(context.Middleware(opts))
```

## API 说明

### Middleware 选项
```go
type Options struct {
    // 请求ID生成器配置
    RequestIDOptions *requestid.Options
    // 是否在响应头中设置请求ID
    SetRequestIDHeader bool
    // 请求ID的响应头名称
    RequestIDHeaderName string
}
```

### 上下文接口
```go
type Context interface {
    context.Context
    GetRequestID() string
    SetRequestID(string)
    GetClientIP() string
    SetClientIP(string)
    GetMetadata() map[string]interface{}
    SetMetadata(map[string]interface{})
    DeleteMetadata(key string)
}
```

## 最佳实践

### 1. 请求追踪
```go
r.Use(context.Middleware(nil))

r.GET("/api", func(c *gin.Context) {
    ctx := context.GetContextFromGin(c)
    
    // 记录请求日志
    logger.WithField("request_id", ctx.GetRequestID()).
           WithField("client_ip", ctx.GetClientIP()).
           Info("处理请求")
           
    // 业务处理
    // ...
})
```

### 2. 元数据传递
```go
r.Use(context.Middleware(nil))

r.GET("/api", func(c *gin.Context) {
    ctx := context.GetContextFromGin(c)
    
    // 设置业务元数据
    ctx.SetMetadata(map[string]interface{}{
        "user_id": "12345",
        "role": "admin",
    })
    
    // 在后续处理中获取元数据
    metadata := ctx.GetMetadata()
    userID := metadata["user_id"].(string)
})
```

## 测试

### 单元测试
```bash
go test -v ./pkg/middleware/context/tests/unit
```

### 集成测试
```bash
go test -v ./pkg/middleware/context/tests/integration
```

## 注意事项

1. 确保在使用中间件前已经设置了 `types.DefaultNewContext`
2. 建议在应用启动时就配置好中间件
3. 请求ID会自动生成,无需手动设置
4. 元数据的key应该使用有意义的命名
5. 在处理大量并发请求时,注意元数据的内存使用

## 常见问题

### Q: 为什么获取不到上下文?
A: 确保:
1. 已经正确设置了 `types.DefaultNewContext`
2. 中间件已经被正确加载
3. 使用 `GetContextFromGin` 方法获取上下文

### Q: 如何自定义请求ID格式?
A: 通过配置 `RequestIDOptions` 来自定义:
```go
opts := &context.Options{
    RequestIDOptions: &requestid.Options{
        Length: 32,
        Prefix: "CUSTOM-",
    },
}
```
