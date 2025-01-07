package benchmark

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"gobase/pkg/middleware/jwt/extractor"
)

func setupRouter(e extractor.TokenExtractor) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.GET("/test", func(c *gin.Context) {
		token, err := e.Extract(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": token})
	})
	return r
}

func BenchmarkExtractors(b *testing.B) {
	benchmarks := []struct {
		name      string
		extractor extractor.TokenExtractor
		setup     func(*http.Request)
	}{
		{
			name:      "Header提取器",
			extractor: extractor.NewHeaderExtractor("Authorization", "Bearer "),
			setup: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer test.token.123")
			},
		},
		{
			name:      "Cookie提取器",
			extractor: extractor.NewCookieExtractor("jwt"),
			setup: func(r *http.Request) {
				r.AddCookie(&http.Cookie{
					Name:  "jwt",
					Value: "test.token.456",
				})
			},
		},
		{
			name:      "Query提取器",
			extractor: extractor.NewQueryExtractor("token"),
			setup: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("token", "test.token.789")
				r.URL.RawQuery = q.Encode()
			},
		},
		{
			name: "链式提取器-最佳情况",
			extractor: extractor.ChainExtractor{
				extractor.NewHeaderExtractor("Authorization", "Bearer "),
				extractor.NewCookieExtractor("jwt"),
				extractor.NewQueryExtractor("token"),
			},
			setup: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer test.token.chain")
			},
		},
		{
			name: "链式提取器-最差情况",
			extractor: extractor.ChainExtractor{
				extractor.NewHeaderExtractor("Authorization", "Bearer "),
				extractor.NewCookieExtractor("jwt"),
				extractor.NewQueryExtractor("token"),
			},
			setup: func(r *http.Request) {
				q := r.URL.Query()
				q.Add("token", "test.token.chain")
				r.URL.RawQuery = q.Encode()
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			router := setupRouter(bm.extractor)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/test", nil)
			bm.setup(req)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				router.ServeHTTP(w, req)
			}
		})
	}
}

// BenchmarkExtractors_Extract 直接测试Extract方法的性能
func BenchmarkExtractors_Extract(b *testing.B) {
	benchmarks := []struct {
		name      string
		extractor extractor.TokenExtractor
		setup     func(*gin.Context)
	}{
		{
			name:      "Header提取器-直接调用",
			extractor: extractor.NewHeaderExtractor("Authorization", "Bearer "),
			setup: func(c *gin.Context) {
				c.Request.Header.Set("Authorization", "Bearer test.token.123")
			},
		},
		{
			name:      "Cookie提取器-直接调用",
			extractor: extractor.NewCookieExtractor("jwt"),
			setup: func(c *gin.Context) {
				c.Request.AddCookie(&http.Cookie{
					Name:  "jwt",
					Value: "test.token.456",
				})
			},
		},
		{
			name:      "Query提取器-直接调用",
			extractor: extractor.NewQueryExtractor("token"),
			setup: func(c *gin.Context) {
				q := c.Request.URL.Query()
				q.Add("token", "test.token.789")
				c.Request.URL.RawQuery = q.Encode()
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)
			bm.setup(c)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, _ = bm.extractor.Extract(c)
			}
		})
	}
}
