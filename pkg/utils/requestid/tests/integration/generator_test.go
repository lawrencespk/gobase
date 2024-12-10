package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"gobase/pkg/utils/requestid"

	"github.com/stretchr/testify/assert"
)

// TestConcurrentGeneration 测试并发生成请求ID
func TestConcurrentGeneration(t *testing.T) {
	tests := []struct {
		name        string
		opts        *requestid.Options
		goroutines  int
		iterations  int
		checkUnique bool
	}{
		{
			name: "UUID并发生成",
			opts: &requestid.Options{
				Type:       "uuid",
				Prefix:     "test",
				EnablePool: true,
			},
			goroutines:  100,
			iterations:  1000,
			checkUnique: true,
		},
		{
			name: "Snowflake并发生成",
			opts: &requestid.Options{
				Type:         "snowflake",
				Prefix:       "test",
				WorkerID:     1,
				DatacenterID: 1,
			},
			goroutines:  100,
			iterations:  1000,
			checkUnique: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := requestid.NewGenerator(tt.opts)
			var wg sync.WaitGroup
			ids := make(chan string, tt.goroutines*tt.iterations)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// 启动多个goroutine并发生成ID
			for i := 0; i < tt.goroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < tt.iterations; j++ {
						select {
						case <-ctx.Done():
							return
						default:
							ids <- g.Generate()
						}
					}
				}()
			}

			// 等待所有goroutine完成
			wg.Wait()
			close(ids)

			// 验证生成的ID
			if tt.checkUnique {
				uniqueIDs := make(map[string]bool)
				for id := range ids {
					assert.False(t, uniqueIDs[id], "发现重复ID: %s", id)
					uniqueIDs[id] = true
				}
				assert.Equal(t, tt.goroutines*tt.iterations, len(uniqueIDs), "生成的唯一ID数量不正确")
			}
		})
	}
}

// TestGeneratorRecovery 测试生成器的故障恢复能力
func TestGeneratorRecovery(t *testing.T) {
	opts := &requestid.Options{
		Type:         "snowflake",
		WorkerID:     1,
		DatacenterID: 1,
	}
	g := requestid.NewGenerator(opts)

	snowflakeGen, ok := g.(*requestid.SnowflakeGenerator)
	if !ok {
		t.Fatal("无法转换为 SnowflakeGenerator")
	}

	// 记录第一次生成的ID
	firstID := g.Generate()
	assert.NotEmpty(t, firstID)

	// 获取当前时间戳
	currentTime := time.Now().UnixNano() / 1000000

	// 模拟时钟回拨：将上一次时间戳设置为当前时间后的一个小时
	futureTime := currentTime + 3600000 // 增加一小时的毫秒数
	snowflakeGen.SetLastTimestamp(futureTime)

	// 使用 recover 来捕获 panic
	var panicMsg interface{}
	func() {
		defer func() {
			panicMsg = recover()
		}()
		g.Generate()
	}()

	// 验证是否发生了预期的 panic
	assert.NotNil(t, panicMsg, "应该发生时钟回拨 panic")
	assert.Contains(t, fmt.Sprint(panicMsg), "Clock moved backwards", "panic 消息不符合预期")

	// 重置时间戳到当前时间
	snowflakeGen.SetLastTimestamp(currentTime)

	// 等待一小段时间确保时间戳前进
	time.Sleep(time.Millisecond * 10)

	// 验证恢复后可以正常生成
	recoveredID := g.Generate()
	assert.NotEmpty(t, recoveredID)
	assert.NotEqual(t, firstID, recoveredID)
}

// TestGeneratorPerformance 测试生成器性能
func TestGeneratorPerformance(t *testing.T) {
	tests := []struct {
		name     string
		opts     *requestid.Options
		duration time.Duration
		minOps   int // 最小期望的操作数/秒
	}{
		{
			name: "UUID性能测试",
			opts: &requestid.Options{
				Type:       "uuid",
				EnablePool: true,
			},
			duration: 3 * time.Second,
			minOps:   10000,
		},
		{
			name: "Snowflake性能测试",
			opts: &requestid.Options{
				Type:         "snowflake",
				WorkerID:     1,
				DatacenterID: 1,
			},
			duration: 3 * time.Second,
			minOps:   50000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := requestid.NewGenerator(tt.opts)
			count := 0
			start := time.Now()

			for time.Since(start) < tt.duration {
				g.Generate()
				count++
			}

			opsPerSecond := float64(count) / tt.duration.Seconds()
			assert.GreaterOrEqual(t, opsPerSecond, float64(tt.minOps),
				"性能不满足要求: %.2f ops/s < %d ops/s", opsPerSecond, tt.minOps)
		})
	}
}

// TestGeneratorStability 测试生成器稳定性
func TestGeneratorStability(t *testing.T) {
	opts := &requestid.Options{
		Type:       "uuid",
		EnablePool: true,
	}
	g := requestid.NewGenerator(opts)

	// 长时间运行测试
	duration := 5 * time.Second
	interval := 100 * time.Millisecond
	iterations := int(duration / interval)

	for i := 0; i < iterations; i++ {
		id := g.Generate()
		assert.NotEmpty(t, id, "生成的ID不应为空")
		time.Sleep(interval)
	}
}

// TestGeneratorEdgeCases 测试边界情况
func TestGeneratorEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		opts        *requestid.Options
		expectPanic bool
	}{
		{
			name:        "空配置",
			opts:        nil,
			expectPanic: false,
		},
		{
			name: "无效生成器类型",
			opts: &requestid.Options{
				Type: "invalid",
			},
			expectPanic: false,
		},
		{
			name: "超长前缀",
			opts: &requestid.Options{
				Type:   "uuid",
				Prefix: string(make([]byte, 1000)),
			},
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					g := requestid.NewGenerator(tt.opts)
					g.Generate()
				})
			} else {
				assert.NotPanics(t, func() {
					g := requestid.NewGenerator(tt.opts)
					id := g.Generate()
					assert.NotEmpty(t, id)
				})
			}
		})
	}
}
