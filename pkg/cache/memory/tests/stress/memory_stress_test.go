package stress

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobase/pkg/cache/memory"
	pkgerrors "gobase/pkg/errors"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
)

// 压力测试配置
const (
	testDuration       = 1 * time.Minute // 测试持续时间
	numGoroutines      = 50              // 并发goroutine数量
	maxEntries         = 100000          // 最大缓存条目数
	operationsPerBatch = 100             // 每个批次的操作数
	logInterval        = 5 * time.Second // 日志记录间隔
	maxLatencySamples  = 10000           // 最大延迟采样数
)

// 操作统计
type operationStats struct {
	sets      atomic.Uint64
	gets      atomic.Uint64
	deletes   atomic.Uint64
	errors    atomic.Uint64
	latencies struct {
		sync.RWMutex
		samples map[string][]time.Duration
	}
}

// 测试指标
type testMetrics struct {
	startMemory uint64
	maxMemory   uint64
	gcCount     uint32
	gcTime      time.Duration
	mutex       sync.Mutex
}

// TestCache_StressTest 缓存压力测试
func TestCache_StressTest(t *testing.T) {
	// 创建日志记录器
	logger, err := createLogger()
	require.NoError(t, err)

	// 创建缓存配置
	config := &memory.Config{
		DefaultTTL:      time.Hour,
		CleanupInterval: time.Minute,
		MaxEntries:      maxEntries,
	}

	// 创建缓存实例
	cache, err := memory.NewCache(config, logger)
	require.NoError(t, err)

	// 初始化统计数据
	stats := &operationStats{}

	// 初始化测试指标
	metrics := &testMetrics{
		startMemory: getMemoryUsage(),
	}

	// 创建上下文和取消函数
	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	// 创建等待组
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// 创建完成通道
	done := make(chan struct{})
	defer close(done)

	// 启动资源监控
	go monitorResources(metrics, done)

	// 启动统计打印
	go printStats(t, stats, metrics, done)

	// 启动工作协程
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			runWorker(ctx, cache, stats, id)
		}(i)
	}

	// 等待所有工作协程完成
	wg.Wait()

	// 验证测试结果
	verifyTestResults(t, stats, metrics)
}

// runWorker 运行工作协程
func runWorker(ctx context.Context, cache *memory.Cache, stats *operationStats, workerID int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

	// 预分配一个key池，减少随机性带来的miss
	keyPool := make([]string, 1000)
	for i := range keyPool {
		keyPool[i] = fmt.Sprintf("key-%d-%d", workerID, i)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// 从key池中选择key，增加命中率
			keyIdx := r.Intn(len(keyPool))
			key := keyPool[keyIdx]
			value := fmt.Sprintf("value-%d", r.Int())

			start := time.Now()
			var err error

			switch r.Intn(10) {
			case 0, 1, 2: // 30% Set操作
				err = cache.Set(ctx, key, value, time.Hour)
				recordLatency(stats, "set", time.Since(start))
				stats.sets.Add(1)

			case 3, 4, 5, 6, 7, 8: // 60% Get操作
				_, err = cache.Get(ctx, key)
				recordLatency(stats, "get", time.Since(start))
				stats.gets.Add(1)

			default: // 10% Delete操作
				err = cache.Delete(ctx, key)
				recordLatency(stats, "delete", time.Since(start))
				stats.deletes.Add(1)
			}

			if err != nil {
				// 只统计非预期的错误
				notFoundErr := pkgerrors.NewCacheNotFoundError("cache miss", nil)
				expiredErr := pkgerrors.NewCacheExpiredError("cache expired", nil)

				// 使用 errors.Is 来判断错误类型
				if !pkgerrors.Is(err, notFoundErr) &&
					!pkgerrors.Is(err, expiredErr) {
					stats.errors.Add(1)
				}
			}

			// 添加短暂休眠以减少资源竞争
			time.Sleep(time.Microsecond)
		}
	}
}

// recordLatency 记录操作延迟
func recordLatency(stats *operationStats, op string, latency time.Duration) {
	stats.latencies.Lock()
	defer stats.latencies.Unlock()

	// 延迟初始化
	if stats.latencies.samples == nil {
		stats.latencies.samples = make(map[string][]time.Duration)
	}

	key := fmt.Sprintf("latency-%s", op)
	samples := stats.latencies.samples[key]

	// 如果样本数量超过限制，删除旧的样本
	if len(samples) >= maxLatencySamples {
		// 保留后半部分的样本
		samples = samples[len(samples)/2:]
	}

	// 添加新样本
	samples = append(samples, latency)
	stats.latencies.samples[key] = samples
}

// monitorResources 监控系统资源
func monitorResources(metrics *testMetrics, done <-chan struct{}) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			metrics.mutex.Lock()
			currentMemory := getMemoryUsage()
			if currentMemory > metrics.maxMemory {
				metrics.maxMemory = currentMemory
			}

			var stats runtime.MemStats
			runtime.ReadMemStats(&stats)
			metrics.gcCount = stats.NumGC
			metrics.gcTime = time.Duration(stats.PauseTotalNs)
			metrics.mutex.Unlock()
		}
	}
}

// printStats 打印统计信息
func printStats(t *testing.T, stats *operationStats, metrics *testMetrics, done <-chan struct{}) {
	ticker := time.NewTicker(logInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			metrics.mutex.Lock()
			t.Logf("Stats - Sets: %d, Gets: %d, Deletes: %d, Errors: %d",
				stats.sets.Load(),
				stats.gets.Load(),
				stats.deletes.Load(),
				stats.errors.Load(),
			)
			t.Logf("Memory - Current: %d MB, Max: %d MB, GC Count: %d, GC Time: %v",
				getMemoryUsage()/1024/1024,
				metrics.maxMemory/1024/1024,
				metrics.gcCount,
				metrics.gcTime,
			)
			metrics.mutex.Unlock()

			// 打印延迟统计
			printLatencyStats(t, stats)
		}
	}
}

// printLatencyStats 打印延迟统计信息
func printLatencyStats(t *testing.T, stats *operationStats) {
	ops := []string{"get", "set", "delete"}
	for _, op := range ops {
		if v, ok := stats.latencies.samples[fmt.Sprintf("latency-%s", op)]; ok {
			if len(v) > 0 {
				p50, p95, p99 := getLatencyStats(stats, op)
				t.Logf("%s Latency - P50: %v, P95: %v, P99: %v",
					op, p50, p95, p99)
			}
		}
	}
}

// getLatencyStats 获取延迟统计信息
func getLatencyStats(stats *operationStats, op string) (p50, p95, p99 time.Duration) {
	stats.latencies.RLock()
	defer stats.latencies.RUnlock()

	key := fmt.Sprintf("latency-%s", op)
	samples := stats.latencies.samples[key]
	if len(samples) == 0 {
		return 0, 0, 0
	}

	// 创建副本并排序
	sorted := make([]time.Duration, len(samples))
	copy(sorted, samples)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// 计算百分位数
	p50idx := len(sorted) * 50 / 100
	p95idx := len(sorted) * 95 / 100
	p99idx := len(sorted) * 99 / 100

	return sorted[p50idx], sorted[p95idx], sorted[p99idx]
}

// verifyTestResults 验证测试结果
func verifyTestResults(t *testing.T, stats *operationStats, metrics *testMetrics) {
	// 验证错误率
	totalOps := stats.sets.Load() + stats.gets.Load() + stats.deletes.Load()
	errorRate := float64(stats.errors.Load()) / float64(totalOps)
	assert.Less(t, errorRate, 0.01, "Error rate too high")

	// 验证内存使用
	memoryIncrease := metrics.maxMemory - metrics.startMemory
	assert.Less(t, memoryIncrease, uint64(maxEntries*1000),
		"Memory usage increased too much")

	// 验证延迟
	ops := []string{"get", "set", "delete"}
	for _, op := range ops {
		if v, ok := stats.latencies.samples[fmt.Sprintf("latency-%s", op)]; ok {
			if len(v) > 0 {
				_, p95, _ := getLatencyStats(stats, op)
				assert.Less(t, p95, 100*time.Millisecond,
					fmt.Sprintf("%s 95th percentile latency too high", op))
			}
		}
	}
}

// getMemoryUsage 获取当前内存使用量
func getMemoryUsage() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

// createLogger 创建日志记录器
func createLogger() (types.Logger, error) {
	// 创建文件管理器
	fm := logrus.NewFileManager(logrus.FileOptions{
		BufferSize:    32 * 1024,
		FlushInterval: time.Second,
		MaxOpenFiles:  100,
		DefaultPath:   "logs/cache_stress_test.log",
	})

	// 创建日志选项
	options := &logrus.Options{
		Level:        types.DebugLevel,
		ReportCaller: true,
		AsyncConfig: logrus.AsyncConfig{
			Enable: true,
		},
		CompressConfig: logrus.CompressConfig{
			Enable: false,
		},
	}

	// 创建日志记录器
	return logrus.NewLogger(fm, logrus.QueueConfig{}, options)
}
