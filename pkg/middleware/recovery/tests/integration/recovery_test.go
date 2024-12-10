//go:build windows
// +build windows

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/errors"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/middleware/recovery"

	"golang.org/x/sys/windows"
)

// TokenPrivileges Windows API 结构体
type TokenPrivileges struct {
	PrivilegeCount uint32
	Privileges     [1]LUIDAndAttributes
}

// LUIDAndAttributes Windows API 结构体
type LUIDAndAttributes struct {
	Luid       windows.LUID
	Attributes uint32
}

func init() {
	if runtime.GOOS == "windows" {
		var privileges = []string{
			"SeCreateGlobalPrivilege",
			"SeSecurityPrivilege",
		}

		for _, name := range privileges {
			var luid windows.LUID
			err := windows.LookupPrivilegeValue(nil, windows.StringToUTF16Ptr(name), &luid)
			if err != nil {
				continue
			}

			attrs := TokenPrivileges{
				PrivilegeCount: 1,
				Privileges: [1]LUIDAndAttributes{
					{
						Luid:       luid,
						Attributes: windows.SE_PRIVILEGE_ENABLED,
					},
				},
			}

			token := windows.Token(0)
			err = windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_ADJUST_PRIVILEGES, &token)
			if err != nil {
				continue
			}
			defer token.Close()

			err = windows.AdjustTokenPrivileges(
				token,
				false,
				(*windows.Tokenprivileges)(unsafe.Pointer(&attrs)),
				0,
				nil,
				nil,
			)
			if err != nil {
				continue
			}
		}
	}
}

func TestRecoveryIntegration(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	t.Run("should_integrate_with_error_system", func(t *testing.T) {
		// 创建测试路由
		r := gin.New()
		r.Use(recovery.Recovery())

		// 添加会panic的路由
		r.GET("/panic", func(c *gin.Context) {
			panic(errors.NewSystemError("system error", nil))
		})

		// 发送请求
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
		r.ServeHTTP(w, req)

		// 验证响应状态码
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		// 验证响应内容
		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "1000", response["code"])
		assert.Contains(t, response["message"], "system error")
	})

	t.Run("should_integrate_with_logger", func(t *testing.T) {
		// 创建临时日志目录
		tmpDir := filepath.Join(os.TempDir(), "TestRecoveryIntegration", "should_integrate_with_logger"+strconv.FormatInt(time.Now().UnixNano(), 10))
		t.Log("临时日志目录:", tmpDir)

		// 确保目录存在
		err := os.MkdirAll(tmpDir, 0755)
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// 创建日志配置
		opts := logrus.DefaultOptions()
		opts.OutputPaths = []string{filepath.Join(tmpDir, "app.log")}

		// 修复：设置正确的刷新间隔
		queueConfig := logrus.QueueConfig{
			MaxSize:       1000,
			BatchSize:     100,
			Workers:       1,
			FlushInterval: time.Second, // 设置为1秒
		}

		// 创建文件管理器
		fm := logrus.NewFileManager(logrus.FileOptions{
			BufferSize:   32 * 1024,
			MaxOpenFiles: 100,
			DefaultPath:  filepath.Join(tmpDir, "app.log"),
		})

		// 创建日志器
		logger, err := logrus.NewLogger(fm, queueConfig, opts)
		require.NoError(t, err)
		defer logger.Close()

		// 创建 Gin 路由
		gin.SetMode(gin.TestMode)
		r := gin.New()
		t.Log("Gin路由已创建")

		// 添加恢复中间件
		r.Use(recovery.Recovery(recovery.WithLogger(logger)))

		// 添加会触发 panic 的路由
		r.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})
		t.Log("Panic路由已添加")

		// 发送请求
		t.Log("开始发送请求...")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
		r.ServeHTTP(w, req)
		t.Log("请求已完成")

		// 验证响应
		t.Logf("响应状态码: %d", w.Code)
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var resp map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		t.Logf("响应内容: %v", resp)

		// 等待日志写入
		t.Log("等待2秒确保日志写入...")
		time.Sleep(2 * time.Second)

		// 读取并验证日志文件
		content, err := os.ReadFile(filepath.Join(tmpDir, "app.log"))
		if err != nil {
			t.Errorf("读取日志文件失败: %v", err)
			// 打印目录内容以帮助调试
			files, _ := os.ReadDir(tmpDir)
			t.Logf("目录 %s 内容:", tmpDir)
			for _, f := range files {
				t.Logf("- %s", f.Name())
			}
			return
		}

		assert.Contains(t, string(content), "panic recovered: test panic")
	})

	t.Run("should handle concurrent panics", func(t *testing.T) {
		// 创建测试路由
		r := gin.New()
		r.Use(recovery.Recovery())

		// 添加会panic的路由
		r.GET("/panic", func(c *gin.Context) {
			panic("concurrent panic")
		})

		// 并发发送请求
		concurrency := 10
		var wg sync.WaitGroup
		wg.Add(concurrency)

		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()
				w := httptest.NewRecorder()
				req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
				r.ServeHTTP(w, req)
				assert.Equal(t, http.StatusInternalServerError, w.Code)
			}()
		}

		wg.Wait()
	})
}
