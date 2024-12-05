package unit

import (
	"context"
	"testing"

	"gobase/pkg/logger/elk"
	"gobase/pkg/logger/elk/tests/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIndexOperations(t *testing.T) {
	t.Run("CreateIndex", func(t *testing.T) {
		client := mock.NewMockElkClient()
		require.NoError(t, client.Connect(nil))

		mapping := elk.DefaultIndexMapping()
		err := client.CreateIndex(context.Background(), "test-index", mapping)
		assert.NoError(t, err)
	})

	t.Run("IndexExists", func(t *testing.T) {
		client := mock.NewMockElkClient()
		require.NoError(t, client.Connect(nil))

		// 先创建索引
		mapping := elk.DefaultIndexMapping()
		err := client.CreateIndex(context.Background(), "test-index", mapping)
		require.NoError(t, err)

		exists, err := client.IndexExists(context.Background(), "test-index")
		assert.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("GetIndexMapping", func(t *testing.T) {
		client := mock.NewMockElkClient()
		require.NoError(t, client.Connect(nil))

		// 先创建索引
		mapping := elk.DefaultIndexMapping()
		err := client.CreateIndex(context.Background(), "test-index", mapping)
		require.NoError(t, err)

		mapping, err = client.GetIndexMapping(context.Background(), "test-index")
		assert.NoError(t, err)
		assert.NotNil(t, mapping)
	})

	t.Run("DeleteIndex", func(t *testing.T) {
		client := mock.NewMockElkClient()
		require.NoError(t, client.Connect(nil))

		err := client.DeleteIndex(context.Background(), "test-index")
		assert.NoError(t, err)
	})
}
