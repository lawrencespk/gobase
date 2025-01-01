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
	codes.InvalidParams:   http.StatusBadRequest,            // 400 无效参数
	codes.Unauthorized:    http.StatusUnauthorized,          // 401 未授权
	codes.Forbidden:       http.StatusForbidden,             // 403 禁止访问
	codes.NotFound:        http.StatusNotFound,              // 404 资源不存在
	codes.AlreadyExists:   http.StatusConflict,              // 409 资源已存在
	codes.TooManyRequests: http.StatusTooManyRequests,       // 429 请求过多
	codes.BadRequest:      http.StatusBadRequest,            // 400 错误的请求
	codes.DataConflict:    http.StatusConflict,              // 409 数据冲突
	codes.RequestTimeout:  http.StatusRequestTimeout,        // 408 请求超时
	codes.InvalidFileType: http.StatusBadRequest,            // 400 无效的文件类型
	codes.FileTooLarge:    http.StatusRequestEntityTooLarge, // 413 文件过大

	// 5xx - 服务器错误
	codes.SystemError:        http.StatusInternalServerError, // 500 系统错误
	codes.DatabaseError:      http.StatusInternalServerError, // 500 数据库错误
	codes.CacheError:         http.StatusInternalServerError, // 500 缓存错误
	codes.ServiceUnavailable: http.StatusServiceUnavailable,  // 503 服务不可用
	codes.NetworkError:       http.StatusBadGateway,          // 502 网络错误
	codes.TimeoutError:       http.StatusGatewayTimeout,      // 504 超时错误
	codes.ThirdPartyError:    http.StatusBadGateway,          // 502 第三方服务错误

	// JWT相关错误码映射
	codes.TokenInvalid:      http.StatusUnauthorized, // 401 Token无效
	codes.TokenExpired:      http.StatusUnauthorized, // 401 Token过期
	codes.TokenRevoked:      http.StatusUnauthorized, // 401 Token已被吊销
	codes.TokenNotFound:     http.StatusUnauthorized, // 401 Token不存在
	codes.TokenTypeMismatch: http.StatusUnauthorized, // 401 Token类型不匹配

	codes.ClaimsMissing: http.StatusBadRequest,   // 400 Claims缺失
	codes.ClaimsInvalid: http.StatusBadRequest,   // 400 Claims无效
	codes.ClaimsExpired: http.StatusUnauthorized, // 401 Claims过期

	codes.SignatureInvalid:  http.StatusUnauthorized,        // 401 签名无效
	codes.KeyInvalid:        http.StatusInternalServerError, // 500 密钥无效
	codes.AlgorithmMismatch: http.StatusBadRequest,          // 400 算法不匹配

	codes.BindingInvalid:  http.StatusBadRequest,   // 400 绑定信息无效
	codes.BindingMismatch: http.StatusUnauthorized, // 401 绑定信息不匹配

	codes.SessionInvalid:  http.StatusUnauthorized, // 401 会话无效
	codes.SessionExpired:  http.StatusUnauthorized, // 401 会话过期
	codes.SessionNotFound: http.StatusUnauthorized, // 401 会话不存在

	codes.PolicyViolation: http.StatusForbidden,           // 403 违反安全策略
	codes.RotationFailed:  http.StatusInternalServerError, // 500 密钥轮换失败
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

// GetHTTPStatus 获取HTTP状态码
func GetHTTPStatus(err error) int {
	code := GetErrorCode(err)
	if status, ok := HTTPStatusMapping[code]; ok {
		return status
	}
	// 默认返回500
	return http.StatusInternalServerError
}
