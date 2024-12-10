package requestid_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"gobase/pkg/utils/requestid"
)

func TestUUIDGenerator(t *testing.T) {
	tests := []struct {
		name    string
		opts    *requestid.Options
		checkFn func(t *testing.T, id string)
	}{
		{
			name: "默认配置-生成标准UUID",
			opts: nil,
			checkFn: func(t *testing.T, id string) {
				if len(id) != 36 { // UUID标准长度
					t.Errorf("无效的UUID长度: %d", len(id))
				}
			},
		},
		{
			name: "带前缀配置-生成带前缀UUID",
			opts: &requestid.Options{
				Type:   "uuid",
				Prefix: "test",
			},
			checkFn: func(t *testing.T, id string) {
				if !strings.HasPrefix(id, "test-") {
					t.Errorf("期望前缀'test-', 实际获得: %s", id)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := requestid.NewGenerator(tt.opts)
			id := g.Generate()
			tt.checkFn(t, id)
		})
	}
}

func TestSnowflakeGenerator(t *testing.T) {
	opts := &requestid.Options{
		Type:         "snowflake",
		Prefix:       "test",
		WorkerID:     1,
		DatacenterID: 1,
	}

	g := requestid.NewGenerator(opts)

	// 生成多个ID并验证唯一性
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := g.Generate()

		// 验证唯一性
		if ids[id] {
			t.Errorf("生成了重复的ID: %s", id)
		}
		ids[id] = true

		// 验证前缀
		if !strings.HasPrefix(id, "test-") {
			t.Errorf("期望前缀'test-', 实际获得: %s", id)
		}
	}

	// 验证时间序列性
	id1 := g.Generate()
	time.Sleep(time.Millisecond)
	id2 := g.Generate()

	// 移除前缀后比较数值大小
	num1 := strings.TrimPrefix(id1, "test-")
	num2 := strings.TrimPrefix(id2, "test-")
	if num1 >= num2 {
		t.Errorf("Snowflake ID应该严格递增: id1=%s, id2=%s", id1, id2)
	}
}

func TestCustomGenerator(t *testing.T) {
	counter := 0
	g := requestid.NewCustomGenerator("test", func() string {
		counter++
		return fmt.Sprintf("custom-%d", counter)
	})

	tests := []struct {
		name     string
		expected string
	}{
		{"第一次生成", "test-custom-1"},
		{"第二次生成", "test-custom-2"},
		{"第三次生成", "test-custom-3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := g.Generate()
			if id != tt.expected {
				t.Errorf("期望 %s, 实际获得: %s", tt.expected, id)
			}
		})
	}
}

func BenchmarkGenerators(b *testing.B) {
	benchmarks := []struct {
		name string
		opts *requestid.Options
	}{
		{
			name: "UUID生成器(无前缀)",
			opts: &requestid.Options{Type: "uuid"},
		},
		{
			name: "UUID生成器(带前缀)",
			opts: &requestid.Options{
				Type:   "uuid",
				Prefix: "test",
			},
		},
		{
			name: "Snowflake生成器(无前缀)",
			opts: &requestid.Options{
				Type:         "snowflake",
				WorkerID:     1,
				DatacenterID: 1,
			},
		},
		{
			name: "Snowflake生成器(带前缀)",
			opts: &requestid.Options{
				Type:         "snowflake",
				Prefix:       "test",
				WorkerID:     1,
				DatacenterID: 1,
			},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			g := requestid.NewGenerator(bm.opts)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				g.Generate()
			}
		})
	}
}
