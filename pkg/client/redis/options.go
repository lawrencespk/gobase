package redis

import (
	"time"

	"gobase/pkg/logger/types"

	"github.com/opentracing/opentracing-go"
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

	// 超时配置
	DialTimeout  time.Duration // 连接超时
	ReadTimeout  time.Duration // 读取超时
	WriteTimeout time.Duration // 写入超时

	// TLS配置
	EnableTLS   bool   // 是否启用TLS
	TLSCertFile string // TLS证书文件
	TLSKeyFile  string // TLS密钥文件

	// 集群配置
	EnableCluster bool // 是否启用集群
	RouteRandomly bool // 是否随机路由

	// 监控配置
	EnableMetrics bool // 是否启用监控
	EnableTracing bool // 是否启用链路追踪

	// 日志和追踪
	Logger types.Logger
	Tracer opentracing.Tracer
}

// DefaultOptions 返回默认配置选项
func DefaultOptions() *Options {
	return &Options{
		Addresses:    []string{"localhost:6379"},
		Username:     "",
		Password:     "",
		DB:           0,
		MaxRetries:   3,
		PoolSize:     10,
		MinIdleConns: 5,
		IdleTimeout:  time.Minute * 5,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
		Logger:       &types.BasicLogger{},
		Tracer:       opentracing.NoopTracer{},
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
func WithTracer(tracer opentracing.Tracer) Option {
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
