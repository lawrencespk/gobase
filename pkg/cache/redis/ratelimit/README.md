# Redis Rate Limiter Store

Redis 限流存储适配器，提供基于 Redis 的限流器存储实现。

## 功能特性

- 提供 Redis 存储适配器，用于限流器的数据存储
- 支持 Lua 脚本执行，确保原子性操作
- 支持键值删除操作，用于重置限流器状态

## 使用示例

```go
import (
"gobase/pkg/client/redis"
"gobase/pkg/cache/redis/ratelimit"
)
// 创建 Redis 客户端
client, err := redis.NewClient(
redis.WithAddresses([]string{"localhost:6379"}),
redis.WithDB(0),
)
if err != nil {
panic(err)
}
// 创建限流存储适配器
store := ratelimit.NewStore(client)
```

## 接口说明
### Store 结构体

```go
type Store struct {
client redis.Client
}
```

### 主要方法
#### NewStore
创建新的 Redis 限流存储适配器。

```
go
func NewStore(client redis.Client) Store
```

参数：
- `client`: Redis 客户端接口实现

#### Eval

执行 Lua 脚本，用于原子性的限流操作。

```go
func (s Store) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error)
```

参数：
- `ctx`: 上下文
- `script`: Lua 脚本内容
- `keys`: Redis 键列表
- `args`: 脚本参数列表

返回：
- `interface{}`: 脚本执行结果
- `error`: 错误信息

#### Del

删除一个或多个键，用于重置限流器状态。
```go
func (s Store) Del(ctx context.Context, keys ...string) error
```

参数：
- `ctx`: 上下文
- `keys`: 要删除的键列表

返回：
- `error`: 错误信息

## 注意事项

1. 确保 Redis 客户端配置正确，包括连接参数、超时设置等
2. 在高并发场景下，建议适当调整 Redis 连接池大小
3. 使用 context 进行超时控制，避免操作长时间阻塞
4. 建议在生产环境中使用 Redis 集群以提高可用性

## 依赖

- `gobase/pkg/client/redis`: Redis 客户端接口
- `context`: 标准库上下文包

## 相关包

- `gobase/pkg/ratelimit/redis`: Redis 限流器实现
- `gobase/pkg/ratelimit/core`: 限流器核心接口定义