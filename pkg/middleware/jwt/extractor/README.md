# JWT Token 提取器

## 目录
- [功能说明](#功能说明)
- [接口定义](#接口定义)
- [实现组件](#实现组件)
  - [Header提取器](#header提取器)
  - [Cookie提取器](#cookie提取器)
  - [Query提取器](#query提取器)
  - [链式提取器](#链式提取器)
- [使用示例](#使用示例)
- [性能测试](#性能测试)
- [最佳实践](#最佳实践)

## 功能说明

提供多种从 HTTP 请求中提取 JWT Token 的方式，支持：
- 从请求头提取 Token
- 从 Cookie 提取 Token
- 从 URL 查询参数提取 Token
- 链式组合多个提取器

## 接口定义

所有提取器都实现了 TokenExtractor 接口：

```go
type TokenExtractor interface {
    // Extract 从gin.Context中提取token
    Extract(c *gin.Context) (string, error)
}
```

## 实现组件

### Header提取器

从 HTTP 请求头提取 Token：

```go
// 创建默认的Header提取器（使用Authorization头）
extractor := NewHeaderExtractor("", "")

// 创建自定义Header提取器
extractor := NewHeaderExtractor("X-Token", "Bearer ")
```

### Cookie提取器

从 HTTP Cookie 提取 Token：

```go
// 创建默认的Cookie提取器（使用jwt作为cookie名）
extractor := NewCookieExtractor("")

// 创建自定义Cookie提取器
extractor := NewCookieExtractor("custom_token")
```

### Query提取器

从 URL 查询参数提取 Token：

```go
// 创建默认的Query提取器（使用token作为参数名）
extractor := NewQueryExtractor("")

// 创建自定义Query提取器
extractor := NewQueryExtractor("access_token")
```

### 链式提取器

按顺序尝试多个提取器：

```go
// 创建链式提取器
extractor := ChainExtractor{
    NewHeaderExtractor("Authorization", "Bearer "),
    NewCookieExtractor("jwt"),
    NewQueryExtractor("token"),
}
```

## 使用示例

在 Gin 中间件中使用：

```go
func TokenMiddleware(extractor TokenExtractor) gin.HandlerFunc {
    return func(c *gin.Context) {
        token, err := extractor.Extract(c)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
            c.Abort()
            return
        }
        c.Set("token", token)
        c.Next()
    }
}
```

## 性能测试

详细的性能测试报告和基准测试结果可以在以下位置找到：
- [基准测试代码](tests/benchmark/extractor_bench_test.go)
- [压力测试代码](tests/stress/extractor_stress_test.go)
- [性能测试报告](tests/benchmark/README.md)

## 最佳实践

1. 提取器选择：
   - API服务：优先使用 Header 提取器
   - Web应用：优先使用 Cookie 提取器
   - 特殊场景：使用 Query 提取器（如：文件下载链接）
   - 混合场景：使用链式提取器

2. 性能优化：
   - 确定Token位置时使用单一提取器
   - 链式提取器中将最可能成功的提取器放在前面
   - 高性能场景优先使用 Header 或 Query 提取器

3. 错误处理：
   - 所有提取器都返回标准化的错误
   - Token不存在：TokenNotFoundError
   - Token无效：TokenInvalidError

4. 安全建议：
   - 避免在URL查询参数中传递敏感Token
   - 使用Cookie时建议设置适当的安全属性
   - Header中使用标准的Authorization头
