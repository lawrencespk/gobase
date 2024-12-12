package sampler

import (
	"math/rand"
	"sync/atomic"
)

// Sampler 采样器
type Sampler struct {
	rate   float64
	counts int64
}

// NewSampler 创建采样器
func NewSampler(rate float64) *Sampler {
	return &Sampler{
		rate: rate,
	}
}

// ShouldSample 是否应该采样
func (s *Sampler) ShouldSample() bool {
	atomic.AddInt64(&s.counts, 1)
	return rand.Float64() < s.rate
}

// GetSampledCount 获取采样数
func (s *Sampler) GetSampledCount() int64 {
	return atomic.LoadInt64(&s.counts)
}
