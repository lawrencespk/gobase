package unit

import (
	"encoding/json"
	"testing"
	"time"

	"gobase/pkg/auth/jwt/session"

	"github.com/stretchr/testify/assert"
)

func TestSessionTypes(t *testing.T) {
	t.Run("Session Serialization", func(t *testing.T) {
		now := time.Now()
		sess := &session.Session{
			UserID:    "user-1",
			TokenID:   "token-1",
			ExpiresAt: now,
			CreatedAt: now,
			UpdatedAt: now,
			Metadata: map[string]interface{}{
				"string": "value",
				"number": float64(123),
				"bool":   true,
			},
		}

		// 测试序列化
		data, err := json.Marshal(sess)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)

		// 测试反序列化
		var decoded session.Session
		err = json.Unmarshal(data, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, sess.UserID, decoded.UserID)
		assert.Equal(t, sess.TokenID, decoded.TokenID)
		assert.Equal(t, sess.Metadata["string"], decoded.Metadata["string"])
		assert.Equal(t, sess.Metadata["number"], decoded.Metadata["number"])
		assert.Equal(t, sess.Metadata["bool"], decoded.Metadata["bool"])
	})

	t.Run("Empty Session", func(t *testing.T) {
		sess := &session.Session{}
		data, err := json.Marshal(sess)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})

	t.Run("Nil Metadata", func(t *testing.T) {
		sess := &session.Session{
			UserID:  "user-1",
			TokenID: "token-1",
		}
		data, err := json.Marshal(sess)
		assert.NoError(t, err)
		assert.NotEmpty(t, data)
	})
}
