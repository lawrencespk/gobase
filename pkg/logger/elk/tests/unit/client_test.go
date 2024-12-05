package unit

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/logger/elk"
	"gobase/pkg/logger/elk/tests/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestElkClient(t *testing.T) {
	t.Run("Connect", func(t *testing.T) {
		client := mock.NewMockElkClient()
		config := &elk.ElkConfig{
			Addresses: []string{"http://localhost:9200"},
			Timeout:   30,
		}

		err := client.Connect(config)
		assert.NoError(t, err)
		assert.True(t, client.IsConnected())
	})

	t.Run("ConnectFailure", func(t *testing.T) {
		client := mock.NewMockElkClient()
		mockClient := mock.AsMockClient(client)
		mockClient.SetShouldFailOps(true)

		config := &elk.ElkConfig{
			Addresses: []string{"http://localhost:9200"},
			Timeout:   30,
		}

		err := client.Connect(config)
		assert.Error(t, err)
		assert.False(t, client.IsConnected())
	})

	t.Run("IndexDocument", func(t *testing.T) {
		client := mock.NewMockElkClient()
		mockClient := mock.AsMockClient(client)
		require.NoError(t, client.Connect(nil))

		doc := map[string]interface{}{
			"message": "test message",
			"time":    time.Now(),
		}

		err := client.IndexDocument(context.Background(), "test-index", doc)
		assert.NoError(t, err)

		docs := mockClient.GetDocuments()
		assert.Len(t, docs, 1)
		assert.Equal(t, doc, docs[0])
	})

	t.Run("BulkIndexDocuments", func(t *testing.T) {
		client := mock.NewMockElkClient()
		mockClient := mock.AsMockClient(client)
		require.NoError(t, client.Connect(nil))

		docs := []interface{}{
			map[string]interface{}{"message": "test1"},
			map[string]interface{}{"message": "test2"},
		}

		err := client.BulkIndexDocuments(context.Background(), "test-index", docs)
		assert.NoError(t, err)

		indexedDocs := mockClient.GetDocuments()
		assert.Len(t, indexedDocs, 2)
	})

	t.Run("Query", func(t *testing.T) {
		client := mock.NewMockElkClient()
		require.NoError(t, client.Connect(nil))

		// 先索引一些文档
		docs := []interface{}{
			map[string]interface{}{"message": "test1"},
			map[string]interface{}{"message": "test2"},
		}
		require.NoError(t, client.BulkIndexDocuments(context.Background(), "test-index", docs))

		query := map[string]interface{}{
			"query": map[string]interface{}{
				"match": map[string]interface{}{
					"message": "test1",
				},
			},
		}

		result, err := client.Query(context.Background(), "test-index", query)
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Close", func(t *testing.T) {
		client := mock.NewMockElkClient()
		require.NoError(t, client.Connect(nil))

		err := client.Close()
		assert.NoError(t, err)
		assert.False(t, client.IsConnected())
	})
}
