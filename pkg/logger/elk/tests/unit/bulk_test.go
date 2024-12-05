package unit

import (
	"context"
	"strings"
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
		mockClient := mock.AsMockClient(client)

		processor := elk.NewBulkProcessor(client, &elk.BulkProcessorConfig{
			BatchSize:  2,
			FlushBytes: 1024,
			Interval:   100 * time.Millisecond,
		})

		// 添加文档
		doc1 := map[string]interface{}{"id": 1, "message": "test1"}
		doc2 := map[string]interface{}{"id": 2, "message": "test2"}

		err := processor.Add(ctx, "test-index", doc1)
		require.NoError(t, err)
		err = processor.Add(ctx, "test-index", doc2)
		require.NoError(t, err)

		// 等待自动刷新
		time.Sleep(200 * time.Millisecond)

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
		require.NoError(t, err)

		time.Sleep(200 * time.Millisecond)
		docs := mockClient.GetDocuments()
		assert.Len(t, docs, 0)

		stats := processor.Stats()
		assert.Greater(t, stats.ErrorCount, int64(0), "Should have recorded errors")
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
			FlushBytes: 100,
			Interval:   1 * time.Second,
		})

		// 添加一个小文档
		smallDoc := map[string]interface{}{
			"id":      1,
			"message": "test",
		}
		err := processor.Add(ctx, "test-index", smallDoc)
		require.NoError(t, err)

		// 添加一个较大的文档，这应该会触发刷新
		largeDoc := map[string]interface{}{
			"id":      2,
			"message": strings.Repeat("test", 50),
		}
		err = processor.Add(ctx, "test-index", largeDoc)
		require.NoError(t, err)

		time.Sleep(100 * time.Millisecond)
		mockClient := mock.AsMockClient(client)
		docs := mockClient.GetDocuments()
		assert.Len(t, docs, 2)
	})

	t.Run("Retry Mechanism", func(t *testing.T) {
		client := mock.NewMockElkClient()
		mockClient := mock.AsMockClient(client)
		mockClient.SetShouldFailOps(true)

		processor := elk.NewBulkProcessor(client, &elk.BulkProcessorConfig{
			BatchSize:  1,
			FlushBytes: 1024,
			Interval:   100 * time.Millisecond,
			RetryCount: 3,
			RetryWait:  10 * time.Millisecond,
		})

		doc := map[string]interface{}{"id": 1, "message": "test"}
		err := processor.Add(ctx, "test-index", doc)
		require.NoError(t, err)

		time.Sleep(500 * time.Millisecond)

		stats := processor.Stats()
		assert.Equal(t, int64(5), stats.ErrorCount)
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
