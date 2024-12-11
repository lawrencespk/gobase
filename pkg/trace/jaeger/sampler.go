package jaeger

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"

	"github.com/uber/jaeger-client-go"
)

// Sampler 采样器
type Sampler struct {
	sync.RWMutex
	// 采样类型
	samplerType string
	// 采样参数
	param float64
	// 采样服务器URL
	serverURL string
	// 最大操作数
	maxOperations int
	// 刷新间隔
	refreshInterval time.Duration
	// 采样计数
	count uint64
	// 采样率限制
	rateLimit float64
	// 随机数生成器
	random *rand.Rand
	// 是否启用自适应采样
	adaptive bool
	// 操作采样率映射
	operationRates map[string]float64
}

// NewSampler 创建采样器
func NewSampler(config *SamplerConfig) (*Sampler, error) {
	if err := validateSamplerConfig(config); err != nil {
		return nil, errors.NewThirdPartyError(fmt.Sprintf("[%s] invalid sampler config", codes.JaegerSamplerError), err)
	}

	s := &Sampler{
		samplerType:     config.Type,
		param:           config.Param,
		serverURL:       config.ServerURL,
		maxOperations:   config.MaxOperations,
		refreshInterval: time.Duration(config.RefreshInterval) * time.Second,
		rateLimit:       config.RateLimit,
		random:          rand.New(rand.NewSource(time.Now().UnixNano())),
		adaptive:        config.Adaptive,
		operationRates:  make(map[string]float64),
	}

	// 如果是远程采样,启动定时刷新
	if s.samplerType == "remote" {
		go s.refreshLoop()
	}

	return s, nil
}

// IsSampled 判断是否需要采样
func (s *Sampler) IsSampled(operation string) (bool, []jaeger.Tag) {
	switch s.samplerType {
	case "const":
		return s.constSampling()
	case "probabilistic":
		return s.probabilisticSampling()
	case "rateLimiting":
		return s.rateLimitingSampling()
	case "remote":
		return s.remoteSampling(operation)
	default:
		return false, nil
	}
}

// constSampling 固定采样
func (s *Sampler) constSampling() (bool, []jaeger.Tag) {
	return s.param > 0, []jaeger.Tag{
		jaeger.NewTag("sampler.type", "const"),
		jaeger.NewTag("sampler.param", s.param),
	}
}

// probabilisticSampling 概率采样
func (s *Sampler) probabilisticSampling() (bool, []jaeger.Tag) {
	sample := s.random.Float64() < s.param
	return sample, []jaeger.Tag{
		jaeger.NewTag("sampler.type", "probabilistic"),
		jaeger.NewTag("sampler.param", s.param),
	}
}

// rateLimitingSampling 限速采样
func (s *Sampler) rateLimitingSampling() (bool, []jaeger.Tag) {
	current := atomic.AddUint64(&s.count, 1)
	sample := float64(current) <= s.rateLimit
	return sample, []jaeger.Tag{
		jaeger.NewTag("sampler.type", "ratelimiting"),
		jaeger.NewTag("sampler.param", s.rateLimit),
	}
}

// remoteSampling 远程采样
func (s *Sampler) remoteSampling(operation string) (bool, []jaeger.Tag) {
	s.RLock()
	rate, exists := s.operationRates[operation]
	s.RUnlock()

	// 如果不存在采样率配置,使用默认值
	if !exists {
		rate = s.param
	}

	sample := s.random.Float64() < rate
	return sample, []jaeger.Tag{
		jaeger.NewTag("sampler.type", "remote"),
		jaeger.NewTag("sampler.param", rate),
	}
}

// refreshLoop 定时刷新远程采样配置
func (s *Sampler) refreshLoop() {
	ticker := time.NewTicker(s.refreshInterval)
	defer ticker.Stop()

	for range ticker.C {
		if err := s.refreshSamplingStrategies(); err != nil {
			// 记录错误但继续运行
			continue
		}
	}
}

// refreshSamplingStrategies 刷新采样策略
func (s *Sampler) refreshSamplingStrategies() error {
	// TODO: 实现远程采样策略获取
	// 这里需要调用采样服务器API获取最新的采样策略
	// 目前返回空以待实现
	return nil
}

// updateAdaptiveSampling 更新自适应采样
//
//lint:ignore U1000 预留给将来使用的自适应采样功能
func (s *Sampler) updateAdaptiveSampling(operation string, latency time.Duration, err error) {
	if !s.adaptive {
		return
	}

	s.Lock()
	defer s.Unlock()

	rate := s.operationRates[operation]

	// 根据延迟和错误调整采样率
	if err != nil || latency > time.Second {
		// 增加采样率以捕获更多问题请求
		rate = min(rate*1.5, 1.0)
	} else if latency < time.Millisecond*100 {
		// 降低采样率以减少正常请求的开销
		rate = max(rate*0.9, 0.001)
	}

	s.operationRates[operation] = rate
}

// validateSamplerConfig 验证采样配置
func validateSamplerConfig(config *SamplerConfig) error {
	if config == nil {
		return fmt.Errorf("sampler config is nil")
	}

	switch config.Type {
	case "const":
		if config.Param != 0 && config.Param != 1 {
			return fmt.Errorf("const sampler param must be 0 or 1")
		}
	case "probabilistic":
		if config.Param < 0 || config.Param > 1 {
			return fmt.Errorf("probabilistic sampler param must be between 0 and 1")
		}
	case "rateLimiting":
		if config.RateLimit <= 0 {
			return fmt.Errorf("rate limiting sampler rate limit must be positive")
		}
	case "remote":
		if config.ServerURL == "" {
			return fmt.Errorf("remote sampler requires server URL")
		}
		if config.RefreshInterval <= 0 {
			return fmt.Errorf("remote sampler refresh interval must be positive")
		}
	default:
		return fmt.Errorf("unknown sampler type: %s", config.Type)
	}

	return nil
}

// Close 关闭采样器
func (s *Sampler) Close() error {
	// 目前没有需要清理的资源
	return nil
}

// min 返回两个float64中的较小值
//
//lint:ignore U1000 预留给自适应采样功能使用
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// max 返回两个float64中的较大值
//
//lint:ignore U1000 预留给自适应采样功能使用
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
