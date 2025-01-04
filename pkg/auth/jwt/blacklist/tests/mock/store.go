package mock

import (
	"context"
	"sync"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

// BlacklistItem 表示黑名单项
type BlacklistItem struct {
	Reason    string
	ExpiredAt time.Time
}

// MockStore 实现Store接口的mock存储
type MockStore struct {
	mu sync.RWMutex
	// 存储黑名单信息
	blacklist map[string]BlacklistItem
	// mock错误控制
	shouldError bool
}

// NewMockStore 创建一个新的mock存储
func NewMockStore() *MockStore {
	return &MockStore{
		blacklist: make(map[string]BlacklistItem),
	}
}

// Add 添加到黑名单
func (m *MockStore) Add(ctx context.Context, tokenID, reason string, expiration time.Duration) error {
	if m.shouldError {
		return errors.NewCacheError("mock error", nil)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.blacklist[tokenID] = BlacklistItem{
		Reason:    reason,
		ExpiredAt: time.Now().Add(expiration),
	}
	return nil
}

// Get 获取黑名单信息
func (m *MockStore) Get(ctx context.Context, tokenID string) (string, error) {
	if m.shouldError {
		return "", errors.NewCacheError("mock error", nil)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if item, exists := m.blacklist[tokenID]; exists {
		if time.Now().Before(item.ExpiredAt) {
			return item.Reason, nil
		}
		delete(m.blacklist, tokenID)
	}
	return "", errors.NewError(codes.StoreErrNotFound, "token not found in blacklist", nil)
}

// Remove 从黑名单中移除
func (m *MockStore) Remove(ctx context.Context, tokenID string) error {
	if m.shouldError {
		return errors.NewCacheError("mock error", nil)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.blacklist, tokenID)
	return nil
}

// SetError 设置是否返回mock错误
func (m *MockStore) SetError(shouldError bool) {
	m.shouldError = shouldError
}

// Close 关闭存储
func (m *MockStore) Close() error {
	return nil
}
