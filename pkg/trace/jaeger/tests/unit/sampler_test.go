package unit

import (
	"testing"

	"gobase/pkg/trace/jaeger"
)

func TestNewSampler(t *testing.T) {
	tests := []struct {
		name      string
		config    *jaeger.SamplerConfig
		wantError bool
	}{
		{
			name: "const sampler valid",
			config: &jaeger.SamplerConfig{
				Type:  "const",
				Param: 1,
			},
			wantError: false,
		},
		{
			name: "const sampler invalid param",
			config: &jaeger.SamplerConfig{
				Type:  "const",
				Param: 0.5, // 应该是 0 或 1
			},
			wantError: true,
		},
		{
			name: "probabilistic sampler valid",
			config: &jaeger.SamplerConfig{
				Type:  "probabilistic",
				Param: 0.5,
			},
			wantError: false,
		},
		{
			name: "probabilistic sampler invalid param",
			config: &jaeger.SamplerConfig{
				Type:  "probabilistic",
				Param: 2.0, // 超出范围
			},
			wantError: true,
		},
		{
			name: "rateLimiting sampler valid",
			config: &jaeger.SamplerConfig{
				Type:      "rateLimiting",
				RateLimit: 100,
			},
			wantError: false,
		},
		{
			name: "rateLimiting sampler invalid rate",
			config: &jaeger.SamplerConfig{
				Type:      "rateLimiting",
				RateLimit: -1, // 无效的速率
			},
			wantError: true,
		},
		{
			name: "remote sampler valid",
			config: &jaeger.SamplerConfig{
				Type:            "remote",
				ServerURL:       "http://localhost:5778/sampling",
				RefreshInterval: 60,
			},
			wantError: false,
		},
		{
			name: "remote sampler missing URL",
			config: &jaeger.SamplerConfig{
				Type:            "remote",
				RefreshInterval: 60,
			},
			wantError: true,
		},
		{
			name: "unknown sampler type",
			config: &jaeger.SamplerConfig{
				Type: "unknown",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sampler, err := jaeger.NewSampler(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("NewSampler() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if err == nil {
				if sampler == nil {
					t.Error("NewSampler() returned nil sampler")
					return
				}
				sampler.Close()
			}
		})
	}
}

func TestSamplerDecisions(t *testing.T) {
	tests := []struct {
		name       string
		config     *jaeger.SamplerConfig
		operation  string
		iterations int
		checkFunc  func([]bool) error
	}{
		{
			name: "const sampler always sample",
			config: &jaeger.SamplerConfig{
				Type:  "const",
				Param: 1,
			},
			iterations: 100,
			checkFunc: func(decisions []bool) error {
				for _, d := range decisions {
					if !d {
						t.Error("const sampler with param 1 should always sample")
					}
				}
				return nil
			},
		},
		{
			name: "const sampler never sample",
			config: &jaeger.SamplerConfig{
				Type:  "const",
				Param: 0,
			},
			iterations: 100,
			checkFunc: func(decisions []bool) error {
				for _, d := range decisions {
					if d {
						t.Error("const sampler with param 0 should never sample")
					}
				}
				return nil
			},
		},
		{
			name: "probabilistic sampler",
			config: &jaeger.SamplerConfig{
				Type:  "probabilistic",
				Param: 0.5,
			},
			iterations: 1000,
			checkFunc: func(decisions []bool) error {
				trueCount := 0
				for _, d := range decisions {
					if d {
						trueCount++
					}
				}
				ratio := float64(trueCount) / float64(len(decisions))
				if ratio < 0.4 || ratio > 0.6 {
					t.Errorf("probabilistic sampler ratio %f is outside expected range [0.4, 0.6]", ratio)
				}
				return nil
			},
		},
		{
			name: "rateLimiting sampler",
			config: &jaeger.SamplerConfig{
				Type:      "rateLimiting",
				RateLimit: 50,
			},
			iterations: 100,
			checkFunc: func(decisions []bool) error {
				trueCount := 0
				for _, d := range decisions {
					if d {
						trueCount++
					}
				}
				if trueCount > 50 {
					t.Errorf("rateLimiting sampler allowed %d samples, expected <= 50", trueCount)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sampler, err := jaeger.NewSampler(tt.config)
			if err != nil {
				t.Fatalf("NewSampler() error = %v", err)
			}
			defer sampler.Close()

			decisions := make([]bool, tt.iterations)
			for i := 0; i < tt.iterations; i++ {
				sampled, _ := sampler.IsSampled(tt.operation)
				decisions[i] = sampled
			}

			if err := tt.checkFunc(decisions); err != nil {
				t.Errorf("Sampling check failed: %v", err)
			}
		})
	}
}

func TestSamplerTags(t *testing.T) {
	tests := []struct {
		name         string
		config       *jaeger.SamplerConfig
		expectedTags map[string]interface{}
	}{
		{
			name: "const sampler tags",
			config: &jaeger.SamplerConfig{
				Type:  "const",
				Param: 1,
			},
			expectedTags: map[string]interface{}{
				"sampler.type":  "const",
				"sampler.param": float64(1),
			},
		},
		{
			name: "probabilistic sampler tags",
			config: &jaeger.SamplerConfig{
				Type:  "probabilistic",
				Param: 0.5,
			},
			expectedTags: map[string]interface{}{
				"sampler.type":  "probabilistic",
				"sampler.param": float64(0.5),
			},
		},
		{
			name: "rateLimiting sampler tags",
			config: &jaeger.SamplerConfig{
				Type:      "rateLimiting",
				RateLimit: 100,
			},
			expectedTags: map[string]interface{}{
				"sampler.type":  "ratelimiting",
				"sampler.param": float64(100),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sampler, err := jaeger.NewSampler(tt.config)
			if err != nil {
				t.Fatalf("NewSampler() error = %v", err)
			}
			defer sampler.Close()

			_, tags := sampler.IsSampled("test-operation")

			// 检查标签数量
			if len(tags) != len(tt.expectedTags) {
				t.Errorf("Got %d tags, want %d", len(tags), len(tt.expectedTags))
			}

			// 计算实际的标签数量
			tagCount := 0
			for range tags {
				tagCount++
			}

			if tagCount != len(tt.expectedTags) {
				t.Errorf("Expected %d tags, but found %d", len(tt.expectedTags), tagCount)
			}
		})
	}
}
