package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"time"

	"gobase/pkg/context"

	"github.com/gin-gonic/gin"
)

// JSONFormatter JSON格式化器
type JSONFormatter struct {
	// 是否美化输出
	PrettyPrint bool
	// 是否记录请求头
	EnableRequestHeader bool
	// 是否记录请求体
	EnableRequestBody bool
	// 是否记录响应头
	EnableResponseHeader bool
	// 是否记录响应体
	EnableResponseBody bool
	// 请求体大小限制
	RequestBodyLimit int64
	// 响应体大小限制
	ResponseBodyLimit int64
	// 屏蔽的请求头
	SkipRequestHeaders []string
	// 屏蔽的响应头
	SkipResponseHeaders []string
}

// NewJSONFormatter 创建JSON格式化器
func NewJSONFormatter(config *Config) LogFormatter {
	return &JSONFormatter{
		PrettyPrint:          false,
		EnableRequestHeader:  true,
		EnableRequestBody:    true,
		EnableResponseHeader: true,
		EnableResponseBody:   true,
		RequestBodyLimit:     config.RequestBodyLimit,
		ResponseBodyLimit:    config.ResponseBodyLimit,
		SkipRequestHeaders:   []string{"Authorization", "Cookie"},
		SkipResponseHeaders:  []string{"Set-Cookie"},
	}
}

// Format 实现LogFormatter接口
func (f *JSONFormatter) Format(c *gin.Context, param *LogFormatterParam) map[string]interface{} {
	data := map[string]interface{}{
		"timestamp":  param.StartTime.Format(time.RFC3339),
		"level":      getLogLevel(param.StatusCode, param.Error),
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"query":      c.Request.URL.RawQuery,
		"ip":         c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
		"latency_ms": param.Latency.Milliseconds(),
		"status":     param.StatusCode,
		"req_size":   param.RequestSize,
		"resp_size":  param.ResponseSize,
	}

	// 获取自定义上下文
	customCtx := context.FromGinContext(c)
	if requestID := context.GetRequestID(customCtx); requestID != "" {
		data["request_id"] = requestID
	}

	if traceID := context.GetTraceID(customCtx); traceID != "" {
		data["trace_id"] = traceID
	}

	// 添加请求头
	if f.EnableRequestHeader {
		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			if !containsString(f.SkipRequestHeaders, k) {
				headers[k] = v[0]
			}
		}
		if len(headers) > 0 {
			data["request_headers"] = headers
		}
	}

	// 添加请求体
	if f.EnableRequestBody && c.Request.Body != nil {
		if body, err := f.getRequestBody(c); err == nil {
			data["request_body"] = body
		}
	}

	// 添加响应头
	if f.EnableResponseHeader {
		headers := make(map[string]string)
		for k, v := range c.Writer.Header() {
			if !containsString(f.SkipResponseHeaders, k) {
				headers[k] = v[0]
			}
		}
		if len(headers) > 0 {
			data["response_headers"] = headers
		}
	}

	// 添加响应体
	if f.EnableResponseBody {
		if body := f.getResponseBody(c); body != nil {
			data["response_body"] = body
		}
	}

	// 添加错误信息
	if param.Error != nil {
		data["error"] = param.Error.Error()
	}

	// 添加额外字段
	for k, v := range param.Fields {
		data[k] = v
	}

	return data
}

// getRequestBody 获取请求体
func (f *JSONFormatter) getRequestBody(c *gin.Context) (interface{}, error) {
	if c.Request.Body == nil {
		return nil, nil
	}

	// 读取body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}

	// 重置body
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 限制大小
	if f.RequestBodyLimit > 0 && int64(len(bodyBytes)) > f.RequestBodyLimit {
		return string(bodyBytes[:f.RequestBodyLimit]) + "...(truncated)", nil
	}

	// 尝试解析JSON
	var jsonBody interface{}
	if err := json.Unmarshal(bodyBytes, &jsonBody); err == nil {
		return jsonBody, nil
	}

	// 返回字符串
	return string(bodyBytes), nil
}

// getResponseBody 获取响应体
func (f *JSONFormatter) getResponseBody(c *gin.Context) interface{} {
	if writer, ok := c.Writer.(*bodyWriter); ok {
		body := writer.Body()

		// 限制大小
		if f.ResponseBodyLimit > 0 && int64(len(body)) > f.ResponseBodyLimit {
			return string(body[:f.ResponseBodyLimit]) + "...(truncated)"
		}

		// 尝试解析JSON
		var jsonBody interface{}
		if err := json.Unmarshal(body, &jsonBody); err == nil {
			return jsonBody
		}

		return string(body)
	}
	return nil
}

// getLogLevel 根据状态码和错误获取日志级别
func getLogLevel(statusCode int, err error) string {
	if err != nil {
		return "error"
	}
	if statusCode >= 500 {
		return "error"
	}
	if statusCode >= 400 {
		return "warn"
	}
	return "info"
}

// containsString 检查字符串数组是否包含指定字符串
func containsString(arr []string, str string) bool {
	for _, s := range arr {
		if s == str {
			return true
		}
	}
	return false
}
