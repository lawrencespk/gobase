package redis

import (
	"time"

	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metric"
	"gobase/pkg/trace/jaeger"
)

// Options Redis客户端配置选项
type Options struct {
	// 基础配置
	Addresses  []string // Redis地址列表
	Username   string   // 用户名
	Password   string   // 密码
	DB         int      // 数据库索引
	MaxRetries int      // 最大重试次数

	// 连接池配置
	PoolSize     int           // 连接池大小
	MinIdleConns int           // 最小空闲连接数
	IdleTimeout  time.Duration // 空闲超时时间
	PoolTimeout  time.Duration // 连接池超时时间

	// 超时配置
	DialTimeout  time.Duration // 连接超时
	ReadTimeout  time.Duration // 读取超时
	WriteTimeout time.Duration // 写入超时

	// TLS配置
	EnableTLS   bool   // 是否启用TLS
	TLSCertFile string // TLS证书文件
	TLSKeyFile  string // TLS密钥文件
	SkipVerify  bool   // 是否跳过证书验证

	// 集群配置
	EnableCluster bool // 是否启用集群
	RouteRandomly bool // 是否随机路由

	// 监控配置
	EnableMetrics    bool   // 是否启用指标收集
	MetricsNamespace string // 指标命名空间
	EnableTracing    bool   // 是否启用链路追踪

	// 日志和追踪
	Logger types.Logger
	Tracer *jaeger.Provider

	// 重试配置
	RetryBackoff time.Duration // 重试间隔时间
	ConnTimeout  time.Duration // 连接超时时间

	// 使用 metric.Collector
	Collector metric.Collector // Prometheus collector
}

// DefaultOptions 返回默认配置选项
func DefaultOptions() *Options {
	return &Options{
		Addresses:        []string{},             // 默认不设置地址
		Username:         "",                     // 默认不设置用户名
		Password:         "",                     // 默认不设置密码
		DB:               0,                      // 默认不设置数据库
		MaxRetries:       3,                      // 默认最大重试次数
		PoolSize:         10,                     // 默认连接池大小
		MinIdleConns:     0,                      // 默认最小空闲连接数
		IdleTimeout:      time.Minute * 5,        // 默认空闲超时时间
		PoolTimeout:      time.Second * 4,        // 默认连接池超时时间
		DialTimeout:      time.Second * 5,        // 默认连接超时时间
		ReadTimeout:      time.Second * 3,        // 默认读取超时时间
		WriteTimeout:     time.Second * 3,        // 默认写入超时时间
		RetryBackoff:     time.Millisecond * 100, // 默认重试间隔时间
		ConnTimeout:      time.Second * 3,        // 默认连接超时时间
		Logger:           &types.NoopLogger{},    // 默认不设置日志记录器
		Tracer:           nil,                    // 默认不设置链路追踪器
		EnableMetrics:    false,                  // 默认不启用指标收集
		MetricsNamespace: "redis_client",         // 默认指标命名空间
		EnableTracing:    false,                  // 默认不启用链路追踪
		Collector:        nil,                    // 默认不设置 Prometheus collector
	}
}

// WithCollector 设置 Prometheus collector
func WithCollector(c metric.Collector) Option {
	return func(o *Options) {
		o.Collector = c
	}
}

// Option 配置选项函数
type Option func(*Options)

// WithLogger 设置日志记录器
func WithLogger(logger types.Logger) Option {
	return func(o *Options) {
		o.Logger = logger
	}
}

// WithTracer 设置链路追踪器
func WithTracer(tracer *jaeger.Provider) Option {
	return func(o *Options) {
		o.Tracer = tracer
	}
}

// WithPoolSize 设置连接池大小
func WithPoolSize(size int) Option {
	return func(o *Options) {
		o.PoolSize = size
	}
}

// WithMinIdleConns 设置最小空闲连接数
func WithMinIdleConns(n int) Option {
	return func(o *Options) {
		o.MinIdleConns = n
	}
}

// WithIdleTimeout 设置空闲超时时间
func WithIdleTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.IdleTimeout = timeout
	}
}

// WithAddress 设置Redis地址
func WithAddress(addr string) Option {
	return func(o *Options) {
		o.Addresses = []string{addr}
	}
}

// WithAddresses 设置多个Redis地址
func WithAddresses(addrs []string) Option {
	return func(o *Options) {
		o.Addresses = addrs
	}
}

// WithUsername 设置用户名
func WithUsername(username string) Option {
	return func(o *Options) {
		o.Username = username
	}
}

// WithPassword 设置密码
func WithPassword(password string) Option {
	return func(o *Options) {
		o.Password = password
	}
}

// WithDB 设置数据库索引
func WithDB(db int) Option {
	return func(o *Options) {
		o.DB = db
	}
}

// WithMaxRetries 设置最大重试次数
func WithMaxRetries(retries int) Option {
	return func(o *Options) {
		o.MaxRetries = retries
	}
}

// WithDialTimeout 设置连接超时时间
func WithDialTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.DialTimeout = timeout
	}
}

// WithReadTimeout 设置读取超时时间
func WithReadTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.ReadTimeout = timeout
	}
}

// WithWriteTimeout 设置写入超时时间
func WithWriteTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.WriteTimeout = timeout
	}
}

// WithTLS 设置是否启用TLS
func WithTLS(enable bool) Option {
	return func(o *Options) {
		o.EnableTLS = enable
	}
}

// WithTLSCert 设置TLS证书文件路径
func WithTLSCert(certFile string) Option {
	return func(o *Options) {
		o.TLSCertFile = certFile
	}
}

// WithTLSKey 设置TLS密钥文件路径
func WithTLSKey(keyFile string) Option {
	return func(o *Options) {
		o.TLSKeyFile = keyFile
	}
}

// WithCluster 设置是否启用集群模式
func WithCluster(enable bool) Option {
	return func(o *Options) {
		o.EnableCluster = enable
	}
}

// WithRouteRandomly 设置是否随机路由
func WithRouteRandomly(enable bool) Option {
	return func(o *Options) {
		o.RouteRandomly = enable
	}
}

// WithMetricsNamespace 设置指标命名空间
func WithMetricsNamespace(namespace string) Option {
	return func(o *Options) {
		o.MetricsNamespace = namespace
	}
}

// WithMetrics 设置是否启用指标收集
func WithMetrics(enable bool) Option {
	return func(o *Options) {
		o.EnableMetrics = enable
	}
}

// WithTracing 设置是否启用链路追踪
func WithTracing(enable bool) Option {
	return func(o *Options) {
		o.EnableTracing = enable
	}
}

// WithRetryBackoff 设置重试间隔时间
func WithRetryBackoff(backoff time.Duration) Option {
	return func(o *Options) {
		o.RetryBackoff = backoff
	}
}

// WithConnTimeout 设置连接超时时间
func WithConnTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.ConnTimeout = timeout
	}
}

// WithSkipVerify 设置是否跳过证书验证
func WithSkipVerify(skip bool) Option {
	return func(o *Options) {
		o.SkipVerify = skip
	}
}

// WithEnableCluster 是 WithCluster 的别名
func WithEnableCluster(enable bool) Option {
	return WithCluster(enable)
}

// WithEnableMetrics 是 WithMetrics 的包装函数
func WithEnableMetrics(enable bool) Option {
	return func(o *Options) {
		o.EnableMetrics = enable
		if enable {
			// 如果启用指标收集，但没有设置命名空间，则使用默认值
			if o.MetricsNamespace == "" {
				o.MetricsNamespace = "redis_client"
			}
		}
	}
}

// WithEnableTracing 是 WithTracing 的别名
func WithEnableTracing(enable bool) Option {
	return WithTracing(enable)
}
