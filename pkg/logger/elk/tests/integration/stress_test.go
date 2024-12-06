package integration

import (
	"context"
	"fmt"
	"gobase/pkg/logger/elk"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

//go test -timeout 25m -run "^TestStressELKIntegration$" gobase/pkg/logger/elk/tests/integration -v

// generateRandomString 生成指定长度的随机字符串
func generateRandomString(size int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, size)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 生成随机大小的消息
func generateRandomMessage() string {
	// 定义消息大小范围
	const (
		minSize = 100    // 最小100字节
		maxSize = 100000 // 最大100KB
	)

	// 使用更复杂的分布
	var size int
	if rand.Float64() < 0.9 {
		// 90%的消息在100B-1KB之间
		size = minSize + rand.Intn(900)
	} else if rand.Float64() < 0.09 {
		// 9%的消息在1KB-10KB之间
		size = 1024 + rand.Intn(9*1024)
	} else {
		// 1%的消息在10KB-100KB之间
		size = 10*1024 + rand.Intn(90*1024)
	}
	return generateRandomString(size)
}

// 模拟突发流量
func generateBurst() bool {
	// 增加突发概率到2%
	return rand.Float64() < 0.02
}

// 添加延迟分布统计
type latencyStats struct {
	p50 int64 // 中位数延迟
	p90 int64 // 90分位延迟
	p99 int64 // 99分位延迟
	max int64 // 最大延迟
}

// 模拟网络故障
func simulateNetworkFailure() bool {
	// 0.1%概率出现网络完全中断
	if rand.Float64() < 0.001 {
		time.Sleep(time.Second) // 模拟1秒网络中断
		return true
	}
	return false
}

func TestStressELKIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// 配置参数
	const (
		testDuration  = 5 * time.Minute
		batchSize     = 10000                  // 增加批处理大小
		flushInterval = 100 * time.Millisecond // 减少刷新间隔
		maxDocSize    = 20 * 1024 * 1024       // 增加到20MB
		burstWorkers  = 10                     // 增加突发工作者数量
		burstDuration = 10 * time.Second       // 增加突发持续时间
	)

	// 动态计算工作者数量
	numWorkers := runtime.NumCPU() * 2 // 使用CPU核心数的2倍

	// 使用共享的ELK配置
	elkConfig := getElkConfig()
	elkConfig.Index = fmt.Sprintf("stress-test-%d", time.Now().Unix())
	elkConfig.Timeout = 60 * time.Second

	// 创建Hook配置
	hookOpts := elk.ElkHookOptions{
		Config: elkConfig,
		Levels: []logrus.Level{
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.ErrorLevel,
		},
		BatchConfig: &elk.BulkProcessorConfig{
			BatchSize:    batchSize,
			FlushBytes:   maxDocSize,
			Interval:     flushInterval,
			RetryCount:   3,
			RetryWait:    1 * time.Second,
			CloseTimeout: 30 * time.Second, // 增加关闭超时
		},
		MaxDocSize: maxDocSize,
	}

	// 创建Hook
	hook, err := elk.NewElkHook(hookOpts)
	require.NoError(t, err)
	defer hook.Close()

	// 创建logger
	logger := logrus.New()
	logger.AddHook(hook)

	// 统计计数器
	var (
		successCount uint64
		errorCount   uint64
		bytesWritten uint64
		burstCount   uint64
		maxLatency   int64 // 使用 int64 存储纳秒
		totalLatency int64 // 使用 int64 存储纳秒
		messageCount int64
		startTime    = time.Now()
		activeBursts sync.WaitGroup
		latencies    []time.Duration
		latencyMutex sync.Mutex
	)

	// 创建上下文用于控制测试时间
	ctx, cancel := context.WithTimeout(context.Background(), testDuration)
	defer cancel()

	// 创建等待组
	var wg sync.WaitGroup

	// 监控 goroutine
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		var (
			lastSuccess uint64
			lastBytes   uint64
			lastTime    = time.Now()
		)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				now := time.Now()
				currentSuccess := atomic.LoadUint64(&successCount)
				currentErrors := atomic.LoadUint64(&errorCount)
				currentBytes := atomic.LoadUint64(&bytesWritten)
				currentBursts := atomic.LoadUint64(&burstCount)

				// 计算这个时间窗口的速率
				intervalRate := float64(currentSuccess-lastSuccess) / now.Sub(lastTime).Seconds()
				overallRate := float64(currentSuccess) / time.Since(startTime).Seconds()
				throughput := float64(currentBytes-lastBytes) / (1024 * 1024) / now.Sub(lastTime).Seconds()

				// 计算平均延迟
				count := atomic.LoadInt64(&messageCount)
				var avgLatency float64
				if count > 0 {
					total := atomic.LoadInt64(&totalLatency)
					avgLatency = float64(total) / float64(count)
				}
				maxLat := time.Duration(atomic.LoadInt64(&maxLatency))

				t.Logf("Current Stats:\n"+
					"  Success: %d, Errors: %d\n"+
					"  Current Rate: %.2f msg/sec\n"+
					"  Overall Rate: %.2f msg/sec\n"+
					"  Throughput: %.2f MB/sec\n"+
					"  Total Data: %.2f MB\n"+
					"  Burst Events: %d\n"+
					"  Avg Latency: %.2fms\n"+
					"  Max Latency: %v\n"+
					"  Memory Usage: %s",
					currentSuccess, currentErrors,
					intervalRate,
					overallRate,
					throughput,
					float64(currentBytes)/(1024*1024),
					currentBursts,
					float64(avgLatency)/float64(time.Millisecond),
					maxLat,
					getMemStats(),
				)

				// 更新上一个时间窗口的值
				lastSuccess = currentSuccess
				lastTime = now
				lastBytes = currentBytes

				// 计算延迟分布
				latencyMutex.Lock()
				if len(latencies) > 0 {
					// 复制并排序延迟数据
					sortedLatencies := make([]time.Duration, len(latencies))
					copy(sortedLatencies, latencies)
					sort.Slice(sortedLatencies, func(i, j int) bool {
						return sortedLatencies[i] < sortedLatencies[j]
					})

					// 计算百分位数
					stats := latencyStats{
						p50: int64(sortedLatencies[len(sortedLatencies)*50/100]),
						p90: int64(sortedLatencies[len(sortedLatencies)*90/100]),
						p99: int64(sortedLatencies[len(sortedLatencies)*99/100]),
						max: int64(sortedLatencies[len(sortedLatencies)-1]),
					}

					// 清空切片以避免内存持续增长
					latencies = latencies[:0]

					t.Logf("Latency Distribution:\n"+
						"  P50: %v\n"+
						"  P90: %v\n"+
						"  P99: %v\n"+
						"  Max: %v",
						time.Duration(stats.p50),
						time.Duration(stats.p90),
						time.Duration(stats.p99),
						time.Duration(stats.max))
				}
				latencyMutex.Unlock()
			}
		}
	}()

	// 启动工作goroutine
	t.Logf("Starting %d workers...", numWorkers)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			burstMode := false
			var burstEndTime time.Time

			for {
				select {
				case <-ctx.Done():
					return
				default:
					if simulateNetworkFailure() {
						atomic.AddUint64(&errorCount, 1)
						continue
					}

					simulateNetworkLatency() // 模拟网络延迟
					// 检查是否需要触发突发流量
					if !burstMode && generateBurst() {
						burstMode = true
						burstEndTime = time.Now().Add(burstDuration)
						atomic.AddUint64(&burstCount, 1)

						// 启动额外的临时工作者
						for j := 0; j < burstWorkers; j++ {
							activeBursts.Add(1)
							go func(tmpID int) {
								defer activeBursts.Done()
								for {
									if time.Now().After(burstEndTime) || ctx.Err() != nil {
										return
									}
									start := time.Now()

									// 生成更大的随机消息
									randomData := generateRandomMessage()
									fields := logrus.Fields{
										"worker_id":   fmt.Sprintf("burst-%d", tmpID),
										"timestamp":   time.Now().UnixNano(),
										"random_data": randomData,
										"burst_mode":  true,
									}

									atomic.AddUint64(&bytesWritten, uint64(len(randomData)))

									select {
									case <-ctx.Done():
										return
									default:
										entry := logger.WithFields(fields)
										entry.Info("Burst mode message")
										atomic.AddUint64(&successCount, 1)

										// 记录延迟
										latency := time.Since(start)
										atomic.AddInt64(&totalLatency, int64(latency))
										atomic.AddInt64(&messageCount, 1)

										// 更新最大延迟
										for {
											current := atomic.LoadInt64(&maxLatency)
											if int64(latency) <= current {
												break
											}
											if atomic.CompareAndSwapInt64(&maxLatency, current, int64(latency)) {
												break
											}
										}

										latencyMutex.Lock()
										latencies = append(latencies, latency)
										latencyMutex.Unlock()

										time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
									}
								}
							}(j)
						}
					}

					// 检查是否需要结束突发模式
					if burstMode && time.Now().After(burstEndTime) {
						burstMode = false
					}

					start := time.Now()

					// 生成随机日志数据
					randomData := generateRandomMessage()
					fields := logrus.Fields{
						"worker_id":   workerID,
						"timestamp":   time.Now().UnixNano(),
						"random_data": randomData,
						"burst_mode":  burstMode,
					}

					atomic.AddUint64(&bytesWritten, uint64(len(randomData)))

					// 随机选择日志级别
					level := randomLogLevel()

					// 写入日志
					entry := logger.WithFields(fields)
					switch level {
					case logrus.InfoLevel:
						entry.Info(fmt.Sprintf("Stress test message from worker %d", workerID))
					case logrus.WarnLevel:
						entry.Warn(fmt.Sprintf("Stress test message from worker %d", workerID))
					case logrus.ErrorLevel:
						entry.Error(fmt.Sprintf("Stress test message from worker %d", workerID))
					}
					atomic.AddUint64(&successCount, 1)

					// 记录延迟
					latency := time.Since(start)
					atomic.AddInt64(&totalLatency, int64(latency))
					atomic.AddInt64(&messageCount, 1)

					// 更新最大延迟
					for {
						current := atomic.LoadInt64(&maxLatency)
						if int64(latency) <= current {
							break
						}
						if atomic.CompareAndSwapInt64(&maxLatency, current, int64(latency)) {
							break
						}
					}

					latencyMutex.Lock()
					latencies = append(latencies, latency)
					latencyMutex.Unlock()

					// 动态调整休眠时间，在突发模式下减少休眠
					if burstMode {
						time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
					} else {
						time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
					}
				}
			}
		}(i)
	}

	// 等待测试完成
	<-ctx.Done()
	t.Log("Context done, waiting for workers to finish...")
	wg.Wait()
	t.Log("Workers finished, waiting for burst goroutines...")
	activeBursts.Wait()
	t.Log("All goroutines finished, closing hook...")

	// 最终统计
	finalSuccess := atomic.LoadUint64(&successCount)
	finalErrors := atomic.LoadUint64(&errorCount)
	finalBytes := atomic.LoadUint64(&bytesWritten)
	totalTime := time.Since(startTime)
	rate := float64(finalSuccess) / totalTime.Seconds()

	// 计算最终平均延迟
	var avgLatency float64
	count := atomic.LoadInt64(&messageCount)
	if count > 0 {
		total := atomic.LoadInt64(&totalLatency)
		avgLatency = float64(total) / float64(count)
	}
	maxLat := time.Duration(atomic.LoadInt64(&maxLatency))

	t.Logf("Final Stats:\n"+
		"  Duration: %v\n"+
		"  Success: %d\n"+
		"  Errors: %d\n"+
		"  Rate: %.2f msg/sec\n"+
		"  Total Data: %.2f MB\n"+
		"  Burst Events: %d\n"+
		"  Avg Latency: %.2fms\n"+
		"  Max Latency: %v\n"+
		"  Memory Usage: %s",
		totalTime, finalSuccess, finalErrors, rate,
		float64(finalBytes)/(1024*1024),
		atomic.LoadUint64(&burstCount),
		float64(avgLatency)/float64(time.Millisecond),
		maxLat,
		getMemStats(),
	)

	// 验证结果
	require.Greater(t, finalSuccess, uint64(0), "Should have processed some messages successfully")
	require.Less(t, float64(finalErrors)/float64(finalSuccess), 0.01,
		"Error rate should be less than 1%")
}

// 随机选择日志级别
func randomLogLevel() logrus.Level {
	levels := []logrus.Level{
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
	}
	return levels[rand.Intn(len(levels))]
}

// 获取内存统计信息
func getMemStats() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return fmt.Sprintf("Alloc=%.1fMB, Sys=%.1fMB, NumGC=%v",
		float64(m.Alloc)/(1024*1024),
		float64(m.Sys)/(1024*1024),
		m.NumGC,
	)
}

// 添加网络延迟模拟
func simulateNetworkLatency() {
	if rand.Float64() < 0.01 { // 1%概率出现网络延迟
		time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
	}
}
