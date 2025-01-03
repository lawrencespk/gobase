package mock

import (
	"context"
	"sync"

	"gobase/pkg/auth/jwt/binding"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

// MockStore 实现Store接口的mock存储
type MockStore struct {
	mu sync.RWMutex
	// 存储device绑定信息
	deviceBindings map[string]*binding.DeviceInfo
	// 存储IP绑定信息
	ipBindings map[string]string
	// mock错误控制
	shouldError bool
}

func NewMockStore() *MockStore {
	return &MockStore{
		deviceBindings: make(map[string]*binding.DeviceInfo),
		ipBindings:     make(map[string]string),
	}
}

func (m *MockStore) SaveDeviceBinding(ctx context.Context, userID, tokenID string, device *binding.DeviceInfo) error {
	if m.shouldError {
		return errors.NewError(codes.CacheError, "mock error", nil)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deviceBindings[tokenID] = device
	return nil
}

func (m *MockStore) GetDeviceBinding(ctx context.Context, tokenID string) (*binding.DeviceInfo, error) {
	if m.shouldError {
		return nil, errors.NewCacheError("mock error", nil)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	device, exists := m.deviceBindings[tokenID]
	if !exists {
		return nil, errors.NewError(codes.StoreErrNotFound, "device binding not found", nil)
	}
	return device, nil
}

func (m *MockStore) SaveIPBinding(ctx context.Context, userID, tokenID, ip string) error {
	if m.shouldError {
		return errors.NewError(codes.CacheError, "mock error", nil)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ipBindings[tokenID] = ip
	return nil
}

func (m *MockStore) GetIPBinding(ctx context.Context, tokenID string) (string, error) {
	if m.shouldError {
		return "", errors.NewCacheError("mock error", nil)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	ip, exists := m.ipBindings[tokenID]
	if !exists {
		return "", errors.NewError(codes.StoreErrNotFound, "IP binding not found", nil)
	}
	return ip, nil
}

func (m *MockStore) DeleteBinding(ctx context.Context, tokenID string) error {
	if m.shouldError {
		return errors.NewCacheError("mock error", nil)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.deviceBindings, tokenID)
	delete(m.ipBindings, tokenID)
	return nil
}

// SetError 设置是否返回mock错误
func (m *MockStore) SetError(shouldError bool) {
	m.shouldError = shouldError
}

func (m *MockStore) Close() error {
	return nil
}
