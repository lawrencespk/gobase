package unit

import (
	"testing"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/elk"

	"github.com/stretchr/testify/assert"
)

func TestHandleELKError(t *testing.T) {
	t.Run("ConnectionError", func(t *testing.T) {
		err := errors.NewELKConnectionError("connection failed", nil)
		elk.HandleELKError(err)
		assert.Equal(t, codes.ELKConnectionError, err.(interface{ Code() string }).Code())
	})

	t.Run("IndexError", func(t *testing.T) {
		err := errors.NewELKIndexError("index operation failed", nil)
		elk.HandleELKError(err)
		assert.Equal(t, codes.ELKIndexError, err.(interface{ Code() string }).Code())
	})

	t.Run("QueryError", func(t *testing.T) {
		err := errors.NewELKQueryError("query failed", nil)

		elk.HandleELKError(err)
		assert.Equal(t, codes.ELKQueryError, err.(interface{ Code() string }).Code())
	})

	t.Run("BulkError", func(t *testing.T) {
		err := errors.NewELKBulkError("bulk operation failed", nil)
		elk.HandleELKError(err)
		assert.Equal(t, codes.ELKBulkError, err.(interface{ Code() string }).Code())
	})

	t.Run("ErrorChain", func(t *testing.T) {
		rootErr := errors.NewSystemError("root error", nil)
		wrappedErr := errors.NewELKConnectionError("connection failed", rootErr)
		elk.HandleELKError(wrappedErr)

		assert.Equal(t, codes.ELKConnectionError, wrappedErr.(interface{ Code() string }).Code())
		assert.Equal(t, rootErr, wrappedErr.(interface{ Unwrap() error }).Unwrap())
	})
}

func TestELKErrorTypes(t *testing.T) {
	t.Run("BulkProcessorErrors", func(t *testing.T) {
		assert.Equal(t, codes.ELKBulkError, elk.ErrProcessorClosed.(interface{ Code() string }).Code())
		assert.Equal(t, codes.ELKBulkError, elk.ErrDocumentTooLarge.(interface{ Code() string }).Code())
		assert.Equal(t, codes.ELKBulkError, elk.ErrRetryExhausted.(interface{ Code() string }).Code())
		assert.Equal(t, codes.ELKTimeoutError, elk.ErrCloseTimeout.(interface{ Code() string }).Code())
	})

	t.Run("ErrorWrapping", func(t *testing.T) {
		baseErr := errors.NewSystemError("base error", nil)
		wrappedErr := errors.Wrap(baseErr, "wrapped message")

		assert.True(t, errors.Is(wrappedErr, baseErr))
	})
}

func TestErrorFormatting(t *testing.T) {
	t.Run("ErrorString", func(t *testing.T) {
		err := errors.NewELKConnectionError("connection failed", nil)
		assert.Contains(t, err.Error(), "connection failed")
		assert.Contains(t, err.Error(), codes.ELKConnectionError)
	})
}
