package cache

import "gobase/pkg/errors"

var (
	// ErrNotFound 缓存未找到错误
	ErrNotFound = errors.NewCacheNotFoundError("cache not found", nil)
)
