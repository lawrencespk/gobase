package unit

import (
	"context"
	"testing"
	"time"

	"gobase/pkg/auth/jwt/session"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockStore struct {
	mock.Mock
}

// Save 实现 Store 接口
func (m *mockStore) Save(ctx context.Context, session *session.Session) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

// Get 实现 Store 接口
func (m *mockStore) Get(ctx context.Context, tokenID string) (*session.Session, error) {
	args := m.Called(ctx, tokenID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*session.Session), args.Error(1)
}

// Delete 实现 Store 接口
func (m *mockStore) Delete(ctx context.Context, tokenID string) error {
	args := m.Called(ctx, tokenID)
	return args.Error(0)
}

// Close 实现 Store 接口
func (m *mockStore) Close(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Ping 实现 Store 接口
func (m *mockStore) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// Refresh 实现 Store 接口
func (m *mockStore) Refresh(ctx context.Context, tokenID string, newExpiration time.Time) error {
	args := m.Called(ctx, tokenID, newExpiration)
	return args.Error(0)
}

func TestManager(t *testing.T) {
	store := new(mockStore)
	logger := &mockLogger{}
	metrics := metrics.NewJWTMetrics()

	manager := session.NewManager(store,
		session.WithLogger(logger),
		session.WithMetrics(metrics),
	)

	t.Run("NewManager", func(t *testing.T) {
		assert.NotNil(t, manager)
	})

	t.Run("Store Operations", func(t *testing.T) {
		ctx := context.Background()
		sess := &session.Session{
			TokenID:   "test-token",
			UserID:    "test-user",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		// 测试 Set
		store.On("Save", mock.AnythingOfType("*context.valueCtx"), sess).Return(nil)
		err := manager.Set(ctx, sess.TokenID, sess, time.Hour)
		assert.NoError(t, err)

		// 测试 Get
		store.On("Get", mock.AnythingOfType("*context.valueCtx"), sess.TokenID).Return(sess, nil)
		got, err := manager.Get(ctx, sess.TokenID)
		assert.NoError(t, err)
		assert.Equal(t, sess, got)

		// 测试 Delete
		store.On("Delete", mock.AnythingOfType("*context.valueCtx"), sess.TokenID).Return(nil)
		err = manager.Delete(ctx, sess.TokenID)
		assert.NoError(t, err)

		store.AssertExpectations(t)
	})
}

// Mock Logger
type mockLogger struct {
	types.Logger
}

func (m *mockLogger) Error(ctx context.Context, msg string, fields ...types.Field) {}
func (m *mockLogger) Debug(ctx context.Context, msg string, fields ...types.Field) {}
