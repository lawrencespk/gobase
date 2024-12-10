package unit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/types"
	"gobase/pkg/middleware/recovery"
)

func TestRecovery(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	t.Run("should recover from panic", func(t *testing.T) {
		// 创建测试路由
		r := gin.New()
		r.Use(recovery.Recovery())

		// 添加会panic的路由
		r.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		// 创建测试请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
		r.ServeHTTP(w, req)

		// 验证响应
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		// 验证响应体
		var resp map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Contains(t, resp["message"], "test panic")
		assert.Equal(t, codes.SystemError, resp["code"])
	})

	t.Run("should use custom error handler", func(t *testing.T) {
		var handledErr error

		// 创建测试路由
		r := gin.New()
		r.Use(recovery.Recovery(
			recovery.WithErrorHandler(func(err error) {
				handledErr = err
			}),
		))

		// 添加会panic的路由
		r.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		// 创建测试请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
		r.ServeHTTP(w, req)

		// 验证错误处理
		assert.NotNil(t, handledErr)
		assert.Contains(t, handledErr.Error(), "test panic")
	})

	t.Run("should disable error response", func(t *testing.T) {
		// 创建测试路由
		r := gin.New()
		r.Use(recovery.Recovery(
			recovery.WithDisableErrorResponse(true),
		))

		// 添加会panic的路由
		r.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		// 创建测试请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
		r.ServeHTTP(w, req)

		// 验证响应是否为空
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Empty(t, w.Body.String())
	})

	t.Run("should log stack trace", func(t *testing.T) {
		var loggedFields []types.Field
		mockLogger := &mockLogger{
			errorFunc: func(ctx context.Context, msg string, fields ...types.Field) {
				loggedFields = fields
			},
		}

		// 创建测试路由
		r := gin.New()
		r.Use(recovery.Recovery(
			recovery.WithLogger(mockLogger),
			recovery.WithPrintStack(true),
		))

		// 添加会panic的路由
		r.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		// 创建测试请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
		r.ServeHTTP(w, req)

		// 验证日志字段
		var hasStack bool
		for _, field := range loggedFields {
			if field.Key == "stack" {
				hasStack = true
				assert.NotEmpty(t, field.Value)
				break
			}
		}
		assert.True(t, hasStack, "Stack trace should be logged")
	})
}

// mockLogger 用于测试的日志记录器
type mockLogger struct {
	errorFunc func(context.Context, string, ...types.Field)
	level     types.Level
}

// 基础日志方法
func (m *mockLogger) Debug(ctx context.Context, msg string, fields ...types.Field) {}
func (m *mockLogger) Info(ctx context.Context, msg string, fields ...types.Field)  {}
func (m *mockLogger) Warn(ctx context.Context, msg string, fields ...types.Field)  {}
func (m *mockLogger) Error(ctx context.Context, msg string, fields ...types.Field) {
	if m.errorFunc != nil {
		m.errorFunc(ctx, msg, fields...)
	}
}
func (m *mockLogger) Fatal(ctx context.Context, msg string, fields ...types.Field) {}

// 格式化日志方法
func (m *mockLogger) Debugf(ctx context.Context, format string, args ...interface{}) {}
func (m *mockLogger) Infof(ctx context.Context, format string, args ...interface{})  {}
func (m *mockLogger) Warnf(ctx context.Context, format string, args ...interface{})  {}
func (m *mockLogger) Errorf(ctx context.Context, format string, args ...interface{}) {}
func (m *mockLogger) Fatalf(ctx context.Context, format string, args ...interface{}) {}

// 链式调用方法
func (m *mockLogger) WithContext(ctx context.Context) types.Logger  { return m }
func (m *mockLogger) WithFields(fields ...types.Field) types.Logger { return m }
func (m *mockLogger) WithError(err error) types.Logger              { return m }
func (m *mockLogger) WithTime(t time.Time) types.Logger             { return m }
func (m *mockLogger) WithCaller(skip int) types.Logger              { return m }

// 配置方法
func (m *mockLogger) SetLevel(level types.Level) {
	m.level = level
}
func (m *mockLogger) GetLevel() types.Level {
	return m.level
}

// 同步方法
func (m *mockLogger) Sync() error {
	return nil
}
