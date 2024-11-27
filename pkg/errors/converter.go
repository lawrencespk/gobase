package errors

import (
	"gobase/pkg/errors/codes"
	"gobase/pkg/errors/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HTTPStatusMapping 定义错误码到 HTTP 状态码的映射
var HTTPStatusMapping = map[string]int{
	// 4xx - 客户端错误
	codes.InvalidParams:   http.StatusBadRequest,      // 400
	codes.Unauthorized:    http.StatusUnauthorized,    // 401
	codes.Forbidden:       http.StatusForbidden,       // 403
	codes.NotFound:        http.StatusNotFound,        // 404
	codes.AlreadyExists:   http.StatusConflict,        // 409
	codes.InvalidToken:    http.StatusUnauthorized,    // 401
	codes.TokenExpired:    http.StatusUnauthorized,    // 401
	codes.TooManyRequests: http.StatusTooManyRequests, // 429

	// 5xx - 服务器错误
	codes.SystemError:        http.StatusInternalServerError, // 500
	codes.DatabaseError:      http.StatusInternalServerError, // 500
	codes.CacheError:         http.StatusInternalServerError, // 500
	codes.ServiceUnavailable: http.StatusServiceUnavailable,  // 503
}

// ErrorResponse 定义统一的错误响应结构
type ErrorResponse struct {
	Code    string      `json:"code"`              // 错误码
	Message string      `json:"message"`           // 错误信息
	Details interface{} `json:"details,omitempty"` // 错误详情（可选）
}

// ToHTTPResponse 将错误转换为 HTTP 响应
func ToHTTPResponse(err error) (int, *ErrorResponse) {
	if err == nil {
		return http.StatusOK, nil
	}

	// 转换为自定义错误类型
	var e types.Error
	if customErr, ok := err.(types.Error); ok {
		e = customErr
	} else {
		// 非自定义错误，包装为系统错误
		e = NewSystemError(err.Error(), err).(types.Error)
	}

	// 获取 HTTP 状态码
	status := HTTPStatusMapping[e.Code()]
	if status == 0 {
		status = http.StatusInternalServerError
	}

	// 构造响应
	response := &ErrorResponse{
		Code:    e.Code(),
		Message: e.Message(),
	}

	// 如果有详情信息，添加到响应中
	if details := e.Details(); len(details) > 0 {
		response.Details = details
	}

	return status, response
}

// HandleError Gin 错误处理函数
func HandleError(c *gin.Context, err error) {
	status, response := ToHTTPResponse(err)
	c.JSON(status, response)
}

// ErrorHandler Gin 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 处理请求
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			// 获取最后一个错误
			err := c.Errors.Last().Err
			HandleError(c, err)
			return
		}
	}
}
