# Context 包

## 简介
Context 包提供了一个轻量级但功能强大的上下文管理解决方案,用于在整个请求生命周期中传递元数据、处理超时控制、携带用户信息等。

## 主要功能

### 1. 元数据管理 (metadata.go)
- 支持读写任意类型的键值对数据
- 支持批量设置和获取元数据
- 支持元数据合并
- 支持元数据拷贝
- 线程安全的实现
- 支持父子上下文元数据继承

### 2. 值管理 (values.go)
- 支持类型安全的值存取
- 支持默认值
- 支持值验证
- 支持值转换
- 支持链式操作
- 内置常用类型转换器

### 3. 验证器 (validator.go)
- 支持自定义验证规则
- 支持验证链
- 支持错误收集
- 支持验证组
- 内置常用验证器

### 4. 类型系统 (types/)
- 定义了统一的上下文接口
- 提供了常用常量定义
- 支持类型转换
- 类型安全的实现

## 使用示例

### 基础使用
```go
// 创建上下文
ctx := context.New()

// 设置元数据
ctx.SetMetadata("user_id", "12345")
ctx.SetMetadata("role", "admin")

// 获取元数据
userID := ctx.GetMetadata("user_id")
role := ctx.GetMetadata("role")

// 批量设置
ctx.SetMetadataMap(map[string]interface{}{
    "tenant_id": "t-001",
    "trace_id": "trace-123",
})

// 获取所有元数据
allMeta := ctx.Metadata()
```

### 值操作
```go
// 设置值
ctx.SetValue("count", 100)

// 获取值(带默认值)
count := ctx.GetIntValue("count", 0)

// 类型安全的值获取
if v, ok := ctx.GetInt("count"); ok {
    // 使用值
}

// 链式操作
v := ctx.GetString("key").
      WithDefault("default").
      Transform(strings.ToUpper).
      Value()
```

### 验证器使用
```go
// 创建验证器
v := ctx.Validator()

// 添加验证规则
v.Required("user_id", "用户ID不能为空")
v.Length("name", 2, 20, "名称长度必须在2-20之间")
v.Custom("age", func(v interface{}) error {
    // 自定义验证逻辑
    return nil
})

// 执行验证
if err := v.Validate(); err != nil {
    // 处理验证错误
}
```

## 最佳实践

### 1. 元数据管理
- 使用统一的键名规范
- 及时清理不再使用的元数据
- 避免存储过大的数据
- 注意类型安全

### 2. 值管理
- 优先使用类型安全的方法
- 合理使用默认值
- 谨慎使用类型转换
- 注意值的生命周期

### 3. 验证器
- 统一验证规则
- 合理组织验证组
- 优雅处理验证错误
- 注意验证性能

### 4. 性能优化
- 合理使用对象池
- 避免频繁创建上下文
- 及时释放资源
- 注意内存使用

## 注意事项
1. 上下文对象不是线程安全的,不要在多个 goroutine 中共享
2. 及时调用 Cancel 方法释放资源
3. 避免在上下文中存储过大的数据
4. 注意处理超时和取消信号
5. 正确处理错误情况
6. 遵循统一的命名规范
7. 注意内存泄漏问题
8. 合理使用默认值

## 扩展性
1. 支持自定义元数据序列化
2. 支持自定义验证规则
3. 支持自定义值转换器
4. 支持自定义超时处理
5. 支持自定义错误处理
6. 支持中间件集成