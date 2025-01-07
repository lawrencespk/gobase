# JWT中间件

## 目录
- [功能特性](#功能特性)
- [使用方法](#使用方法)
- [Context操作](#context操作)
- [性能测试](#性能测试)
- [实现细节](#实现细节)

## 功能特性
- 支持JWT令牌的验证和解析
- 提供完整的Context操作API
- 支持自定义Claims扩展
- 支持访问令牌(Access Token)和刷新令牌(Refresh Token)
- 提供用户身份、角色、权限等信息的存取
- 支持设备ID和IP地址的追踪

## 使用方法

### 基础用法
```go
// 创建中间件
jwtMiddleware := jwt.New(
    jwt.WithSigningKey([]byte("your-secret-key")),
)

// 在gin中使用
router.Use(jwtMiddleware.Handle())
```

### 自定义Claims
```go
type CustomClaims struct {
    *jwt.StandardClaims
    // 自定义字段
}
```

## Context操作
支持以下Context操作:
- Claims操作: `WithClaims/GetClaims`
- Token操作: `WithToken/GetToken`
- TokenType操作: `WithTokenType/GetTokenType`
- UserID操作: `WithUserID/GetUserID`
- UserName操作: `WithUserName/GetUserName`
- Roles操作: `WithRoles/GetRoles`
- Permissions操作: `WithPermissions/GetPermissions`
- DeviceID操作: `WithDeviceID/GetDeviceID`
- IPAddress操作: `WithIPAddress/GetIPAddress`

### 示例
```go
// 获取用户ID
userID, err := jwtContext.GetUserID(ctx)

// 获取用户角色
roles, err := jwtContext.GetRoles(ctx)

// 获取用户权限
permissions, err := jwtContext.GetPermissions(ctx)
```

## 性能测试
性能测试报告请参考: [性能测试报告](tests/benchmark/README.md)

主要性能指标:
- 平均延迟: 0.156毫秒/请求
- 内存分配: 12.2KB/请求
- 分配次数: 185次/请求

## 实现细节

### 目录结构
```
jwt/
├── README.md
├── middleware.go       // 中间件主实现
├── context/           // Context相关实现
│   ├── context.go
│   └── tests/
├── tests/            // 测试相关
│   ├── benchmark/    // 性能测试
│   ├── integration/  // 集成测试
│   └── unit/        // 单元测试
```

### 测试覆盖
- 单元测试: Context操作、Token验证等
- 集成测试: 中间件集成测试
- 性能测试: 基准测试和压力测试
