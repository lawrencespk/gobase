package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"gobase/pkg/logger/elk"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestElkHookIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	config := getElkConfig()
	t.Logf("ES 连接配置: %+v", config)

	client := elk.NewElkClient()
	err := client.Connect(config)
	require.NoError(t, err, "连接 Elasticsearch 失败")
	defer client.Close()

	// 验证 Elasticsearch 服务状态
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.HealthCheck(ctx)
	require.NoError(t, err, "Elasticsearch 服务未响应")

	// 设置测试索引
	testIndex := fmt.Sprintf("test-logs-%d", time.Now().Unix())
	t.Logf("使用测试索引: %s", testIndex)

	// 创建索引
	mapping := &elk.IndexMapping{
		Settings: map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		Mappings: map[string]interface{}{
			"properties": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"type": "date",
				},
				"message": map[string]interface{}{
					"type": "text",
				},
			},
		},
	}
	err = client.CreateIndex(ctx, testIndex, mapping)
	require.NoError(t, err, "创建索引失败")

	// 创建测试用的logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// 创建并配置Hook
	hook, err := elk.NewElkHook(elk.ElkHookOptions{
		Config: config,
		Levels: []logrus.Level{
			logrus.DebugLevel,
			logrus.InfoLevel,
			logrus.WarnLevel,
			logrus.ErrorLevel,
		},
		Index: testIndex,
		BatchConfig: &elk.BulkProcessorConfig{
			BatchSize:    100,
			FlushBytes:   1 * 1024 * 1024,
			Interval:     100 * time.Millisecond,
			RetryCount:   2,
			RetryWait:    500 * time.Millisecond,
			CloseTimeout: 5 * time.Second,
		},
		ErrorLogger: logger,
	})
	require.NoError(t, err)
	defer hook.Close()

	logger.AddHook(hook)

	// 运行子测试
	for name, tf := range map[string]func(t *testing.T, ctx context.Context, logger *logrus.Logger, client *elk.ElkClient, hook *elk.ElkHook, testIndex string){
		"LogLevels":     testLogLevels,
		"LogFields":     testLogFields,
		"BulkLogging":   testBulkLogging,
		"ErrorHandling": testErrorHandling,
		"Cleanup":       testCleanup,
	} {
		testCtx, testCancel := context.WithTimeout(context.Background(), 120*time.Second)
		t.Run(name, func(t *testing.T) {
			tf(t, testCtx, logger, client, hook, testIndex)
		})
		testCancel()
	}
}

func testLogLevels(t *testing.T, ctx context.Context, logger *logrus.Logger, client *elk.ElkClient, hook *elk.ElkHook, testIndex string) {
	// ... 实现日志级别测试 ...
}

func testLogFields(t *testing.T, ctx context.Context, logger *logrus.Logger, client *elk.ElkClient, hook *elk.ElkHook, testIndex string) {
	// ... 实现字段测试 ...
}

func testBulkLogging(t *testing.T, ctx context.Context, logger *logrus.Logger, client *elk.ElkClient, hook *elk.ElkHook, testIndex string) {
	const totalLogs = 1000
	const numWorkers = 4

	t.Logf("开始批量日志测试，计划写入 %d 条日志到索引: %s", totalLogs, testIndex)

	// 验证索引是否存在
	exists, err := client.IndexExists(ctx, testIndex)
	if err != nil {
		t.Logf("检查索引存在时发生错误: %v", err)
	} else {
		t.Logf("索引 %s 是否存在: %v", testIndex, exists)
	}

	// 添加等待时间，确保 bulk processor 已经准备好
	time.Sleep(1 * time.Second)

	var wg sync.WaitGroup
	logsPerWorker := totalLogs / numWorkers

	// 添加计时起点
	start := time.Now()

	// 并发写入日志
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			start := workerID * logsPerWorker
			end := start + logsPerWorker

			for i := start; i < end; i++ {
				entry := &logrus.Entry{
					Logger:  logger,
					Data:    logrus.Fields{"batch": "test", "worker": workerID},
					Time:    time.Now(),
					Level:   logrus.InfoLevel,
					Message: fmt.Sprintf("Test log message %d", i),
				}

				if err := hook.Fire(entry); err != nil {
					t.Logf("写入日志失败 [worker %d, log %d]: %v", workerID, i, err)
				}

				if i > start && (i-start)%25 == 0 {
					t.Logf("Worker %d: 已写入 %d 条日志", workerID, i-start)
				}
			}
		}(w)
	}

	// 等待所有写入完成
	wg.Wait()

	// 使用新添加的方法
	t.Log("正在刷新缓冲区...")
	if err = hook.GetBulkProcessor().Flush(ctx); err != nil {
		t.Logf("刷新失败: %v", err)
	}

	t.Logf("所有日志写入完成，开始验证")

	// 在验证之前手动刷新索引
	err = client.RefreshIndex(ctx, testIndex)
	if err != nil {
		t.Fatalf("刷新索引失败: %v", err)
	}

	// 增加等待时间确保文档可被搜索
	time.Sleep(5 * time.Second)

	// 修改查询条件确保正确匹配
	count, err := getLogCount(ctx, client, testIndex, "batch", "test")
	require.NoError(t, err, "查询日志计数失败")

	if count >= totalLogs {
		t.Logf("批量日志测试完成，总耗时：%v", time.Since(start))
		return
	}

	t.Logf("当前已索引日志数: %d/%d", count, totalLogs)

	// 在写入完成后，添加索引刷新和验证步骤
	t.Log("所有日志写入完成，等待索引刷新...")

	// 强制刷新索引
	if err := client.RefreshIndex(ctx, testIndex); err != nil {
		t.Logf("刷新索引失败: %v", err)
	}

	// 等待一段时间让文档变得可搜索
	time.Sleep(2 * time.Second)

	// 验证文档数量
	t.Log("开始验证文档数量...")
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"batch": "test",
			},
		},
	}

	result, err := client.Query(ctx, testIndex, query)
	if err != nil {
		t.Logf("查询失败: %v", err)
	} else {
		t.Logf("查询结果: %+v", result)
	}

	// 获取索引统计信息
	stats, err := client.GetIndexStats(ctx, testIndex)
	if err != nil {
		t.Logf("获取索引统计信息失败: %v", err)
	} else {
		t.Logf("索引统计信息: %+v", stats)
	}

	t.Fatal("未能在规定时间内完成日志写入和验证")
}

func testErrorHandling(t *testing.T, ctx context.Context, logger *logrus.Logger, client *elk.ElkClient, hook *elk.ElkHook, testIndex string) {
	// ... 实现错误处理测试 ...
}

func testCleanup(t *testing.T, ctx context.Context, logger *logrus.Logger, client *elk.ElkClient, hook *elk.ElkHook, testIndex string) {
	// ... 实现清理测试 ...
}

// 辅助函数获取日志计数
func getLogCount(ctx context.Context, client *elk.ElkClient, index, field, value string) (int, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"term": map[string]interface{}{
				field: value,
			},
		},
	}

	result, err := client.Query(ctx, index, query)
	if err != nil {
		return 0, fmt.Errorf("查询失败: %w", err)
	}

	hits, ok := result.(map[string]interface{})["hits"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("unexpected response format: %+v", result)
	}

	total, ok := hits["total"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("unexpected total format: %+v", hits)
	}

	valueFloat, ok := total["value"].(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected value format: %+v", total)
	}

	return int(valueFloat), nil
}
