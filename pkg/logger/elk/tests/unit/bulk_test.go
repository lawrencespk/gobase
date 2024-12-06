package unit

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"gobase/pkg/logger/elk"
	"gobase/pkg/logger/elk/tests/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBulkProcessor(t *testing.T) {
	ctx := context.Background()

	t.Run("Basic Bulk Operations", func(t *testing.T) {
		client := mock.NewMockElkClient()
		require.NoError(t, client.Connect(nil))

		processor := elk.NewBulkProcessor(client, &elk.BulkProcessorConfig{
			BatchSize:  2,
			FlushBytes: 1024,
			Interval:   50 * time.Millisecond,
		})
		defer processor.Close()

		// 添加文档
		doc1 := map[string]interface{}{"id": 1, "message": "test1"}
		doc2 := map[string]interface{}{"id": 2, "message": "test2"}

		err := processor.Add(ctx, "test-index", doc1)
		require.NoError(t, err)
		err = processor.Add(ctx, "test-index", doc2)
		require.NoError(t, err)

		// 手动刷新并等待
		err = processor.Flush(ctx)
		require.NoError(t, err)

		mockClient := mock.AsMockClient(client)

		docs := mockClient.GetDocuments()
		assert.Len(t, docs, 2)
	})

	t.Run("Manual Flush", func(t *testing.T) {
		client := mock.NewMockElkClient()
		require.True(t, client.IsConnected())

		processor := elk.NewBulkProcessor(client, &elk.BulkProcessorConfig{
			BatchSize:  10,
			FlushBytes: 1024,
			Interval:   1 * time.Second,
			RetryCount: 1,
			RetryWait:  10 * time.Millisecond,
		})

		doc := map[string]interface{}{"id": 1, "message": "test"}
		err := processor.Add(ctx, "test-index", doc)
		require.NoError(t, err)

		err = processor.Flush(ctx)
		require.NoError(t, err)

		mockClient := mock.AsMockClient(client)
		docs := mockClient.GetDocuments()
		assert.Len(t, docs, 1)
	})

	t.Run("Batch Size Trigger", func(t *testing.T) {
		client := mock.NewMockElkClient()
		require.True(t, client.IsConnected())

		processor := elk.NewBulkProcessor(client, &elk.BulkProcessorConfig{
			BatchSize:  2,
			FlushBytes: 1024,
			Interval:   1 * time.Second,
			RetryCount: 1,
			RetryWait:  10 * time.Millisecond,
		})

		for i := 0; i < 2; i++ {
			doc := map[string]interface{}{"id": i, "message": "test"}
			err := processor.Add(ctx, "test-index", doc)
			require.NoError(t, err)
		}

		time.Sleep(100 * time.Millisecond)
		mockClient := mock.AsMockClient(client)
		docs := mockClient.GetDocuments()
		assert.Len(t, docs, 2)
	})

	t.Run("Error Handling", func(t *testing.T) {
		client := mock.NewMockElkClient()
		mockClient := mock.AsMockClient(client)
		mockClient.SetShouldFailOps(true)

		processor := elk.NewBulkProcessor(client, &elk.BulkProcessorConfig{
			BatchSize:  1,
			FlushBytes: 1024,
			Interval:   100 * time.Millisecond,
			RetryCount: 1,
			RetryWait:  10 * time.Millisecond,
		})

		doc := map[string]interface{}{"id": 1, "message": "test"}
		err := processor.Add(ctx, "test-index", doc)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mock bulk index failure")

		mockClient.SetShouldFailOps(false)
	})

	t.Run("Concurrent Operations", func(t *testing.T) {
		client := mock.NewMockElkClient()
		require.NoError(t, client.Connect(nil))
		mockClient := mock.AsMockClient(client)

		processor := elk.NewBulkProcessor(client, &elk.BulkProcessorConfig{
			BatchSize:  10,
			FlushBytes: 1024,
			Interval:   100 * time.Millisecond,
		})

		// 并发添加文档
		const numDocs = 20
		errCh := make(chan error, numDocs)

		for i := 0; i < numDocs; i++ {
			go func(id int) {
				doc := map[string]interface{}{"id": id, "message": "test"}
				errCh <- processor.Add(ctx, "test-index", doc)
			}(i)
		}

		// 收集错误
		for i := 0; i < numDocs; i++ {
			err := <-errCh
			assert.NoError(t, err)
		}

		// 手动刷新并等待处理完成
		err := processor.Flush(ctx)
		require.NoError(t, err)

		docs := mockClient.GetDocuments()
		assert.Len(t, docs, numDocs)
	})

	t.Run("Close Processor", func(t *testing.T) {
		client := mock.NewMockElkClient()
		require.True(t, client.IsConnected())

		processor := elk.NewBulkProcessor(client, &elk.BulkProcessorConfig{
			BatchSize:  10,
			FlushBytes: 1024,
			Interval:   100 * time.Millisecond,
			RetryCount: 1,
			RetryWait:  10 * time.Millisecond,
		})

		doc := map[string]interface{}{"id": 1, "message": "test"}
		err := processor.Add(ctx, "test-index", doc)
		require.NoError(t, err)

		err = processor.Close()
		require.NoError(t, err)

		mockClient := mock.AsMockClient(client)
		docs := mockClient.GetDocuments()
		assert.Len(t, docs, 1)

		err = processor.Add(ctx, "test-index", doc)
		assert.Error(t, err)
	})

	t.Run("FlushBytes Trigger", func(t *testing.T) {
		client := mock.NewMockElkClient()
		require.True(t, client.IsConnected())

		processor := elk.NewBulkProcessor(client, &elk.BulkProcessorConfig{
			BatchSize:  100,
			FlushBytes: 1024,
			Interval:   1 * time.Second,
		})

		// 添加两个小文档
		for i := 0; i < 2; i++ {
			doc := map[string]interface{}{
				"id":      i,
				"message": "test",
			}
			err := processor.Add(ctx, "test-index", doc)
			require.NoError(t, err)
		}

		err := processor.Flush(ctx)
		require.NoError(t, err)

		mockClient := mock.AsMockClient(client)
		docs := mockClient.GetDocuments()
		assert.Equal(t, 2, len(docs))
	})

	t.Run("Retry Mechanism", func(t *testing.T) {
		client := mock.NewMockElkClient()
		mockClient := mock.AsMockClient(client)

		var callCount int32
		mockClient.SetCustomFailure(func() bool {
			count := atomic.AddInt32(&callCount, 1)
			t.Logf("调用次数: %d", count)
			return count == 1
		})

		processor := elk.NewBulkProcessor(client, &elk.BulkProcessorConfig{
			BatchSize:  1,
			FlushBytes: 1024,
			Interval:   50 * time.Millisecond,
			RetryCount: 1,
			RetryWait:  10 * time.Millisecond,
		})

		doc := map[string]interface{}{"id": 1, "message": "test"}
		err := processor.Add(ctx, "test-index", doc)
		require.NoError(t, err)

		// 等待第一次处理
		time.Sleep(100 * time.Millisecond)

		// 手动刷新并关闭
		err = processor.Flush(ctx)
		require.NoError(t, err)

		err = processor.Close()
		require.NoError(t, err)

		// 验证结果
		docs := mockClient.GetDocuments()
		count := atomic.LoadInt32(&callCount)

		t.Logf("最终文档数: %d", len(docs))
		t.Logf("总调用次数: %d", count)

		assert.Equal(t, 1, len(docs), "应该只有一个文档被索引")
		assert.Equal(t, 2, int(count), "应该只重试一次")
	})

	t.Run("Graceful Shutdown", func(t *testing.T) {
		client := mock.NewMockElkClient()
		require.True(t, client.IsConnected())

		processor := elk.NewBulkProcessor(client, &elk.BulkProcessorConfig{
			BatchSize:    10,
			FlushBytes:   1024,
			Interval:     1 * time.Second,
			CloseTimeout: 100 * time.Millisecond,
		})

		// 添加一些文档
		for i := 0; i < 5; i++ {
			doc := map[string]interface{}{"id": i, "message": "test"}
			err := processor.Add(ctx, "test-index", doc)
			require.NoError(t, err)
		}

		// 优雅关闭
		err := processor.Close()
		require.NoError(t, err)

		mockClient := mock.AsMockClient(client)
		docs := mockClient.GetDocuments()
		assert.Len(t, docs, 5)
	})
}
