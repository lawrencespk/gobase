# Grafana 监控模块
## 简介
模块提供了完整的 Grafana 监控解决方案，包括仪表盘模板、告警规则配置和告警通知管理。集成了系统指标、运行时指标、HTTP 指标等多维度监控。
## 目录结构

grafana/
├── config/
│ └── manager.go # 配置管理器
├── template/
│ ├── dashboard/ # 仪表盘模板
│ │ ├── http.json # HTTP指标仪表盘
│ │ ├── system.json # 系统指标仪表盘
│ │ ├── ratelimit.json # 限流指标仪表盘
│ │ ├── logger.json # 日志指标仪表盘
│ │ ├── redis.json # Redis指标仪表盘
│ │ └── runtime.json # 运行时指标仪表盘
│ ├── alert/
│ │ └── rules.yaml # 告警规则配置
│ └── alertmanager/
│ └── config.yaml # 告警管理器配置
└── README.md

## 功能特性

### 1. 仪表盘模板
- **HTTP 指标仪表盘**
  - 请求速率监控
  - 错误率监控
  - 延迟分布监控
  - 状态码分布

- **系统指标仪表盘**
  - CPU 使用率
  - 内存使用情况
  - 磁盘使用情况
  - 网络 IO 监控

- **运行时指标仪表盘**
  - Goroutine 数量
  - GC 监控
  - 堆内存使用
  - 系统线程数

- **Redis 指标仪表盘**
  - 操作速率监控
  - 操作延迟监控
  - 连接池状态
  - 错误率监控

### 2. 告警规则
- **HTTP 告警**
  - 高错误率告警
  - 慢响应告警
  - 高延迟告警
  - 流量突增告警

- **系统告警**
  - CPU 高使用率告警
  - 内存不足告警
  - 磁盘空间不足告警
  - Goroutine 数量过高告警

- **业务告警**
  - 业务错误率告警
  - 处理速率过低告警

- **Redis告警**
  - 高错误率告警
  - 高延迟告警
  - 连接池耗尽告警
  - 连接错误告警



### 3. 告警通知
- 支持多种通知渠道
  - 邮件通知
  - Slack 通知
  - 钉钉通知（可扩展）
- 告警分级管理
- 告警抑制规则
- 告警分组策略

## 快速开始

### 1. 配置文件准备

```bash
创建配置目录
mkdir -p /etc/grafana/provisioning/dashboards
mkdir -p /etc/grafana/provisioning/datasources
复制仪表盘模板
cp template/dashboard/.json /etc/grafana/provisioning/dashboards/
复制告警规则
cp template/alert/rules.yaml /etc/prometheus/rules/
复制告警管理器配置
cp template/alertmanager/config.yaml /etc/alertmanager/
```

### 2. 配置 Grafana 数据源

```yaml
apiVersion: 1
datasources:
name: Prometheus
type: prometheus
access: proxy
url: http://prometheus:9090
isDefault: true
```

### 3. 启动服务

```bash
启动 Grafana
docker run -d \
-p 3000:3000 \
-v /etc/grafana:/etc/grafana \
grafana/grafana
启动 Alertmanager
docker run -d \
-p 9093:9093 \
-v /etc/alertmanager:/etc/alertmanager \
prom/alertmanager
```

## 使用说明

### 1. 仪表盘导入
1. 登录 Grafana (默认 http://localhost:3000)
2. 导航到 "+" -> "Import"
3. 上传或粘贴仪表盘 JSON 文件内容
4. 选择 Prometheus 数据源
5. 点击 "Import" 完成导入

### 2. 告警配置
1. 在 Prometheus 中加载告警规则：

```yaml
rule_files:
/etc/prometheus/rules/.yaml
```

```yaml
alerting:
alertmanagers:
static_configs:
targets:
alertmanager:9093
```

### 3. 通知渠道配置
1. 配置邮件通知：
   - 修改 alertmanager/config.yaml 中的 SMTP 配置
   - 更新接收邮箱地址

2. 配置 Slack 通知：
   - 在 Slack 中创建 Webhook
   - 更新 alertmanager/config.yaml 中的 Webhook URL

## 最佳实践

1. **仪表盘组织**
   - 按照服务类型组织仪表盘
   - 使用文件夹管理不同环境的仪表盘
   - 统一命名规范

2. **告警配置**
   - 合理设置告警阈值
   - 配置告警优先级
   - 使用告警抑制避免告警风暴

3. **性能优化**
   - 合理设置采样率
   - 优化查询语句
   - 配置适当的数据保留期

## 常见问题

1. **仪表盘无数据**
   - 检查 Prometheus 数据源配置
   - 验证指标采集是否正常
   - 检查时间范围设置

2. **告警未触发**
   - 检查告警规则语法
   - 验证告警条件是否满足
   - 检查 Alertmanager 配置

3. **通知未送达**
   - 检查网络连接
   - 验证通知渠道配置
   - 查看 Alertmanager 日志

## 维护与支持

- 定期更新仪表盘模板
- 根据业务需求调整告警规则
- 监控系统性能指标
- 及时响应告警通知