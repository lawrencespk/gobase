package unit

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/store"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/errors/types"

	"github.com/stretchr/testify/assert"
)

type memoryStoreTestSuite struct {
	store *store.MemoryStore
	ctx   context.Context
}

func setupMemoryTest() (*memoryStoreTestSuite, error) {
	store, err := store.NewMemoryStore(store.Options{
		CleanupInterval: time.Minute,
	})
	if err != nil {
		return nil, err
	}

	return &memoryStoreTestSuite{
		store: store,
		ctx:   context.Background(),
	}, nil
}

// testClaims 实现 jwt.Claims 接口的测试用结构体
type testClaims struct {
	jwt.StandardClaims
	UserID      string        `json:"user_id"`
	UserName    string        `json:"user_name"`
	Roles       []string      `json:"roles"`
	Permissions []string      `json:"permissions"`
	DeviceID    string        `json:"device_id"`
	IPAddress   string        `json:"ip_address"`
	TokenType   jwt.TokenType `json:"token_type"`
	TokenID     string        `json:"token_id"`
}

// 实现 jwt.Claims 接口的所有必需方法
func (c *testClaims) GetUserID() string           { return c.UserID }
func (c *testClaims) GetUserName() string         { return c.UserName }
func (c *testClaims) GetRoles() []string          { return c.Roles }
func (c *testClaims) GetPermissions() []string    { return c.Permissions }
func (c *testClaims) GetDeviceID() string         { return c.DeviceID }
func (c *testClaims) GetIPAddress() string        { return c.IPAddress }
func (c *testClaims) GetTokenType() jwt.TokenType { return c.TokenType }
func (c *testClaims) GetTokenID() string          { return c.TokenID }
func (c *testClaims) Validate() error             { return nil }

// createTestClaims 创建测试用的Claims
func createTestClaims(userID string) jwt.Claims {
	return &testClaims{
		StandardClaims: *jwt.NewStandardClaims(),
		UserID:         userID,
		TokenType:      jwt.AccessToken,
	}
}

func TestMemoryStore_Save(t *testing.T) {
	suite, err := setupMemoryTest()
	assert.NoError(t, err)

	t.Run("successful save", func(t *testing.T) {
		info := &jwt.TokenInfo{
			Raw:       "test_token",
			Type:      jwt.AccessToken,
			ExpiresAt: time.Now().Add(time.Hour),
			Claims:    createTestClaims("test_user"),
		}

		err := suite.store.Set(suite.ctx, info.Raw, info, time.Hour)
		assert.NoError(t, err)
	})

	t.Run("save with nil info", func(t *testing.T) {
		err := suite.store.Set(suite.ctx, "test_token", nil, time.Hour)
		assert.Error(t, err)
		assert.True(t, errors.IsTokenInvalidError(err))
	})
}

func TestMemoryStore_Get(t *testing.T) {
	suite, err := setupMemoryTest()
	assert.NoError(t, err)

	t.Run("get existing token", func(t *testing.T) {
		info := &jwt.TokenInfo{
			Raw:       "test_token",
			Type:      jwt.AccessToken,
			ExpiresAt: time.Now().Add(time.Hour),
			Claims:    createTestClaims("test_user"),
		}
		err := suite.store.Set(suite.ctx, info.Raw, info, time.Hour)
		assert.NoError(t, err)

		retrieved, err := suite.store.Get(suite.ctx, info.Raw)
		assert.NoError(t, err)
		assert.Equal(t, info.Raw, retrieved.Raw)
		assert.Equal(t, info.Claims.GetUserID(), retrieved.Claims.GetUserID())
	})

	t.Run("get non-existent token", func(t *testing.T) {
		_, err := suite.store.Get(suite.ctx, "non-existent")
		assert.Error(t, err)
		if e, ok := err.(types.Error); ok {
			assert.Equal(t, codes.StoreErrNotFound, e.Code())
		}
	})
}

func TestMemoryStore_Delete(t *testing.T) {
	suite, err := setupMemoryTest()
	assert.NoError(t, err)

	t.Run("delete existing token", func(t *testing.T) {
		info := &jwt.TokenInfo{
			Raw:       "test_token",
			Type:      jwt.AccessToken,
			ExpiresAt: time.Now().Add(time.Hour),
			Claims:    createTestClaims("test_user"),
		}
		err := suite.store.Set(suite.ctx, info.Raw, info, time.Hour)
		assert.NoError(t, err)

		err = suite.store.Delete(suite.ctx, info.Raw)
		assert.NoError(t, err)

		_, err = suite.store.Get(suite.ctx, info.Raw)
		assert.Error(t, err)
		if e, ok := err.(types.Error); ok {
			assert.Equal(t, codes.StoreErrNotFound, e.Code())
		}
	})
}
