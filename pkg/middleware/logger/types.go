package logger

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// LogFormatter 日志格式化接口
type LogFormatter interface {
	// Format 格式化请求日志
	Format(ctx *gin.Context, param *LogFormatterParam) map[string]interface{}
}

// LogFormatterParam 格式化参数
type LogFormatterParam struct {
	// 请求开始时间
	StartTime time.Time
	// 请求处理时长
	Latency time.Duration
	// 请求状态码
	StatusCode int
	// 请求体大小
	RequestSize int64
	// 响应体大小
	ResponseSize int64
	// 错误信息
	Error error
	// 额外字段
	Fields map[string]interface{}
}

// Sampler 采样接口
type Sampler interface {
	// Sample 是否采样该请求
	Sample(ctx *gin.Context) bool
}

// MetricsCollector 指标收集接口(为Prometheus准备)
type MetricsCollector interface {
	// BeginRequest 开始请求监控
	BeginRequest(method, path string)
	// CollectRequest 收集请求指标
	CollectRequest(ctx context.Context, param *MetricsParam)
}

// MetricsParam 指标参数
type MetricsParam struct {
	// 请求方法
	Method string
	// 请求路径
	Path string
	// 状态码
	StatusCode int
	// 请求处理时长
	Latency time.Duration
	// 请求体大小
	RequestSize int64
	// 响应体大小
	ResponseSize int64
	// 是否有错误
	HasError bool
	// 额外标签
	Labels map[string]string
}

// TracingConfig 追踪配置
type TracingConfig struct {
	// 是否启用追踪
	Enable bool
	// 采样率
	SamplingRate float64
	// 追踪标签
	Tags map[string]string
}

// Tracer 追踪接口(为Jaeger准备)
type Tracer interface {
	// StartSpan 开始一个追踪span
	StartSpan(ctx context.Context, operationName string) (context.Context, interface{})
	// FinishSpan 结束追踪span
	FinishSpan(span interface{})
	// SetTag 设置标签
	SetTag(span interface{}, key string, value interface{})
	// SetError 设置错误
	SetError(span interface{}, err error)
}

// BodyWriter 响应体Writer接口
type BodyWriter interface {
	gin.ResponseWriter
	// Body 获取响应体
	Body() []byte
	// Status 获取状态码
	Status() int
	// Size 获取响应大小
	Size() int
	// Close 关闭writer
	Close() error
}
