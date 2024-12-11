package logger

import (
	"bytes"
	stdctx "context"
	"io"
	"time"

	"gobase/pkg/context"
	"gobase/pkg/errors"
	"gobase/pkg/logger/types"

	"github.com/gin-gonic/gin"
)

// Middleware 创建日志中间件
func Middleware(opts ...Option) gin.HandlerFunc {
	// 应用选项
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	// 验证选项
	if err := options.Validate(); err != nil {
		panic(err)
	}

	return func(c *gin.Context) {
		// 跳过不需要记录的路径
		if containsString(options.Config.SkipPaths, c.Request.URL.Path) {
			c.Next()
			return
		}

		// 采样判断
		if !options.Sampler.Sample(c) {
			c.Next()
			return
		}

		// 开始时间
		start := time.Now()

		// 包装ResponseWriter以捕获响应体
		writer := newBodyWriter(c, options.Config)
		c.Writer = writer

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 开始指标收集
		if options.MetricsCollector != nil {
			options.MetricsCollector.BeginRequest(c.Request.Method, c.Request.URL.Path)
		}

		// 开始追踪
		var span interface{}
		if options.Tracer != nil {
			var ctx = c.Request.Context()
			ctx, span = options.Tracer.StartSpan(ctx, "http_request")
			c.Request = c.Request.WithContext(ctx)
			defer options.Tracer.FinishSpan(span)
		}

		// 处理panic
		defer func() {
			if r := recover(); r != nil {
				// 将 recover() 的返回值转换为错误
				var err error
				switch v := r.(type) {
				case error:
					err = v
				case string:
					err = errors.NewSystemError(v, nil)
				default:
					err = errors.NewSystemError("unknown panic", nil)
				}

				// 记录panic错误
				sysErr := errors.NewSystemError("panic recovered", err)
				ctx := stdctx.Background()
				options.Logger.Error(ctx, "panic recovered", types.Error(sysErr))

				// 设置追踪错误
				if span != nil {
					options.Tracer.SetError(span, sysErr)
				}

				// 返回500错误
				c.AbortWithStatus(500)
			}
		}()

		// 继续处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)

		// 慢请求判断
		if options.Config.SlowThreshold > 0 && latency.Milliseconds() > options.Config.SlowThreshold {
			ctx := stdctx.Background()
			options.Logger.Warn(ctx, "slow request detected",
				types.Field{Key: "path", Value: c.Request.URL.Path},
				types.Field{Key: "latency", Value: latency.Milliseconds()},
			)
		}

		// 获取错误信息
		var err error
		if len(c.Errors) > 0 {
			err = c.Errors.Last().Err
		}

		// 设置追踪标签和错误
		if span != nil {
			options.Tracer.SetTag(span, "http.status_code", c.Writer.Status())
			options.Tracer.SetTag(span, "http.method", c.Request.Method)
			options.Tracer.SetTag(span, "http.path", c.Request.URL.Path)
			if err != nil {
				options.Tracer.SetError(span, err)
			}
		}

		// 格式化日志
		param := &LogFormatterParam{
			StartTime:    start,
			Latency:      latency,
			StatusCode:   c.Writer.Status(),
			RequestSize:  int64(len(requestBody)),
			ResponseSize: int64(len(writer.Body())),
			Error:        err,
			Fields:       make(map[string]interface{}),
		}

		// 添加追踪ID
		if ctx := context.FromGinContext(c); ctx != nil {
			if traceID := context.GetTraceID(ctx); traceID != "" {
				param.Fields["trace_id"] = traceID
			}
		}

		// 格式化并记录日志
		fields := options.Formatter.Format(c, param)
		logFunc := getLogFunc(options.Logger, c.Writer.Status(), err)
		logFunc("http request", fields)

		// 收集指标
		if options.MetricsCollector != nil {
			options.MetricsCollector.CollectRequest(c, &MetricsParam{
				Method:       c.Request.Method,
				Path:         c.Request.URL.Path,
				StatusCode:   c.Writer.Status(),
				Latency:      latency,
				RequestSize:  param.RequestSize,
				ResponseSize: param.ResponseSize,
				HasError:     err != nil,
			})
		}
	}
}

// getLogFunc 根据状态码和错误获取日志函数
func getLogFunc(logger types.Logger, statusCode int, err error) func(msg string, fields map[string]interface{}) {
	return func(msg string, fields map[string]interface{}) {
		// 转换 fields 为 types.Field 切片
		logFields := make([]types.Field, 0, len(fields))
		for k, v := range fields {
			logFields = append(logFields, types.Field{Key: k, Value: v})
		}

		ctx := stdctx.Background()

		if err != nil {
			logger.Error(ctx, msg, logFields...)
			return
		}
		if statusCode >= 500 {
			logger.Error(ctx, msg, logFields...)
			return
		}
		if statusCode >= 400 {
			logger.Warn(ctx, msg, logFields...)
			return
		}
		logger.Info(ctx, msg, logFields...)
	}
}
