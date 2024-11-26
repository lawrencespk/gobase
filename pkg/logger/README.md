

# 日志模块设计文档

## 项目结构

```pkg/
└── logger/
├── bootstrap/
│   └── bootstrap.go      # 系统初始化和对外接口
├── config/
│   └── config.go         # 配置管理
├── manager/
│   └── manager.go        # 日志实例管理
├── types/
│   └── types.go          # 类型和接口定义
├── logrus/
│   └── adapter.go        # Logrus适配器实现
├── elk/
│   └── adapter.go        # ELK适配器实现
├── factory.go            # 日志工厂
└── README.md             # 模块文档
```

## 功能说明

### 核心组件

1. **types (types/types.go)**
   
   - 定义所有日志级别（Debug/Info/Warn/Error/Fatal/Panic）
   - 定义统一的日志接口
   - 定义配置结构和字段类型
   - 确保类型系统的一致性
2. **factory (factory.go)**
   
   - 实现工厂模式
   - 负责创建具体的日志实例
   - 支持动态扩展新的日志类型
3. **bootstrap (bootstrap/bootstrap.go)**
   
   - 系统初始化入口
   - 提供全局访问点
   - 管理默认配置和实例
4. **config (config/config.go)**
   
   - 管理日志配置
   - 提供默认配置
   - 支持动态配置更新
   - 配置验证和转换
5. **manager (manager/manager.go)**
   
   - 管理日志实例生命周期
   - 实现单例模式
   - 保证并发安全
   - 支持实例复用
6. **logrus adapter (logrus/adapter.go)**
   
   - Logrus 适配器实现
   - 支持所有日志级别
   - 支持结构化日志
   - 支持自定义格式化
   - 支持 Hook 机制
7. **elk adapter (elk/adapter.go)**
   
   - ELK 适配器实现
   - 支持所有日志级别
   - 支持结构化日志
   - 支持自定义索引
   - 支持异步写入

---

## 特性

1. **完整性**
   
   - 支持全部日志级别
   - 完整的错误处理
   - 完整的配置选项
   - 完整的上下文支持
2. **扩展性**
   
   - 支持新增日志适配器
   - 支持自定义配置
   - 支持自定义格式化
   - 支持 Hook 机制
3. **并发安全**
   
   - 线程安全的实例管理
   - 并发安全的配置访问
   - 原子操作保证
   - 锁机制优化
4. **性能优化**
   
   - 实例复用
   - 读写锁分离
   - 异步日志支持
   - 内存使用优化
5. **使用便捷**
   
   - 简单的初始化流程
   - 统一的接口定义
   - 链式调用支持
   - 完整的类型提示

---

## 使用示例

```go
// 初始化
bootstrap.Initialize()

// 获取日志器
logger, err := bootstrap.GetLogger("default")
if err != nil {
    panic(err)
}

// 使用日志器
logger.WithFields(types.Fields{
    "user": "john",
}).Info("User logged in")
```

## 注意事项

1. **初始化顺序**
   * 必须先调用 `bootstrap.Initialize()`
   * 配置更新需要重新初始化
2. **资源管理**
   * 注意 ELK 连接的释放
   * 避免创建过多实例
3. **配置管理**
   * 建议使用配置文件
   * 注意敏感信息保护
4. **性能考虑**
   * 合理使用日志级别
   * 避免过多字段
   * 考虑使用异步日志
