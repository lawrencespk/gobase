# JWT Token 提取器

## 功能说明

提供多种从 HTTP 请求中提取 JWT Token 的方式：

- Header 提取器：从请求头提取 Token
- Cookie 提取器：从 Cookie 提取 Token
- Query 提取器：从 URL 查询参数提取 Token
- 链式提取器：按顺序尝试多个提取器

## 性能测试报告

### 完整 HTTP 请求场景性能

| 提取器类型 | 操作耗时 (ns/op) | 内存分配 (B/op) | 分配次数 (allocs/op) |
|------------|-----------------|----------------|-------------------|
| Header提取器 | 507.2 | 539 | 8 |
| Cookie提取器 | 654.9 | 754 | 10 |
| Query提取器 | 683.0 | 970 | 11 |
| 链式提取器(最佳) | 522.4 | 538 | 8 |
| 链式提取器(最差) | 13994 | 6424 | 93 |

### Extract 方法直接调用性能

| 提取器类型 | 操作耗时 (ns/op) | 内存分配 (B/op) | 分配次数 (allocs/op) |
|------------|-----------------|----------------|-------------------|
| Header提取器 | 26.12 | 0 | 0 |
| Cookie提取器 | 134.30 | 200 | 2 |
| Query提取器 | 8.45 | 0 | 0 |

## 性能分析

1. 提取器性能排名（从快到慢）：
   - Query提取器（直接调用）
   - Header提取器（直接调用）
   - Cookie提取器（直接调用）

2. 内存分配情况：
   - Header和Query提取器在直接调用时无内存分配
   - Cookie提取器需要少量内存分配
   - 完整HTTP请求场景下都需要一定内存分配

3. 链式提取器特点：
   - 最佳情况：性能接近单个提取器
   - 最差情况：性能显著下降（约27倍）
   - 建议将最可能成功的提取器放在链的前面

## 使用建议

1. 优先选择：
   - 如果确定Token位置：使用单一提取器
   - 如果Token位置不固定：使用链式提取器，注意提取器顺序

2. 性能优化：
   - 将最常用的提取方式放在链式提取器的前面
   - 避免不必要的链式提取器使用
   - 在高性能要求场景下优先使用Header或Query提取器

3. 配置建议：
   - Header提取器：适用于API场景
   - Cookie提取器：适用于Web场景
   - Query提取器：适用于特殊场景（如下载链接）
   - 链式提取器：适用于混合场景

## 压力测试结果

在并发测试中（100个并发，每个并发1000请求）：
- 错误率：<0.1%
- QPS：能够稳定支持百万级别请求（在最佳情况下）
- 内存使用稳定，无明显泄漏 