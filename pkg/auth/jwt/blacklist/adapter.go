package blacklist

import (
	"context"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

// StoreAdapter 将 Store 接口适配为 TokenBlacklist 接口
type StoreAdapter struct {
	store Store
}

// NewStoreAdapter 创建一个新的存储适配器
func NewStoreAdapter(store Store) TokenBlacklist {
	return &StoreAdapter{store: store}
}

// Add 实现 TokenBlacklist 接口
func (a *StoreAdapter) Add(ctx context.Context, token string, expiration time.Duration) error {
	return a.store.Add(ctx, token, "", expiration)
}

// IsBlacklisted 实现 TokenBlacklist 接口
func (a *StoreAdapter) IsBlacklisted(ctx context.Context, token string) (bool, error) {
	_, err := a.store.Get(ctx, token)
	if err != nil {
		// 使用 errors.Is 和 codes.StoreErrNotFound 来检查
		notFoundErr := errors.NewError(codes.StoreErrNotFound, "", nil)
		if errors.Is(err, notFoundErr) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Remove 实现 TokenBlacklist 接口
func (a *StoreAdapter) Remove(ctx context.Context, token string) error {
	return a.store.Remove(ctx, token)
}

// Clear 实现 TokenBlacklist 接口
func (a *StoreAdapter) Clear(ctx context.Context) error {
	// 由于 Store 接口没有 Clear 方法，这里返回未实现错误
	return errors.NewSystemError("clear operation not supported", nil)
}

// Close 实现 TokenBlacklist 接口
func (a *StoreAdapter) Close() error {
	return a.store.Close()
}
