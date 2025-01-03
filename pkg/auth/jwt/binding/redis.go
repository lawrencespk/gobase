package binding

import (
	"context"
	"encoding/json"
	"time"

	"gobase/pkg/client/redis"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
)

const (
	// key前缀
	deviceBindingKeyPrefix = "auth:binding:device:"
	ipBindingKeyPrefix     = "auth:binding:ip:"
	// 默认过期时间
	defaultExpiration = 24 * time.Hour
)

// RedisStore Redis存储实现
type RedisStore struct {
	client redis.Client
}

// NewRedisStore 创建Redis存储
func NewRedisStore(client redis.Client) (Store, error) {
	if client == nil {
		return nil, errors.NewError(codes.InvalidParams, "redis client is required", nil)
	}

	// 初始化 metrics
	InitMetrics()
	// 不在这里注册 collector

	return &RedisStore{
		client: client,
	}, nil
}

// SaveDeviceBinding 保存设备绑定
func (s *RedisStore) SaveDeviceBinding(ctx context.Context, userID, deviceID string, info *DeviceInfo) error {
	if userID == "" || deviceID == "" || info == nil {
		RecordError()
		return errors.NewError(codes.InvalidParams, "invalid parameters", nil)
	}

	// 序列化设备信息
	data, err := json.Marshal(info)
	if err != nil {
		RecordError()
		return errors.NewError(codes.StoreErrSet, "failed to marshal device info", err)
	}

	// 保存设备绑定
	key := deviceBindingKeyPrefix + deviceID
	err = s.client.Set(ctx, key, string(data), defaultExpiration)
	if err != nil {
		RecordError()
		return errors.NewError(codes.StoreErrSet, "failed to save device binding", err)
	}

	RecordDeviceBinding()
	return nil
}

// GetDeviceBinding 获取设备绑定
func (s *RedisStore) GetDeviceBinding(ctx context.Context, deviceID string) (*DeviceInfo, error) {
	if deviceID == "" {
		RecordError()
		return nil, errors.NewError(codes.InvalidParams, "device ID is required", nil)
	}

	// 获取设备绑定
	key := deviceBindingKeyPrefix + deviceID
	data, err := s.client.Get(ctx, key)
	if err != nil {
		RecordError()
		return nil, errors.NewError(codes.StoreErrGet, "failed to get device binding", err)
	}

	// 反序列化设备信息
	var info DeviceInfo
	err = json.Unmarshal([]byte(data), &info)
	if err != nil {
		RecordError()
		return nil, errors.NewError(codes.StoreErrGet, "failed to unmarshal device info", err)
	}

	RecordIPBinding()
	return &info, nil
}

// SaveIPBinding 保存IP绑定
func (s *RedisStore) SaveIPBinding(ctx context.Context, userID, deviceID, ip string) error {
	if userID == "" || deviceID == "" || ip == "" {
		RecordError()
		return errors.NewError(codes.InvalidParams, "invalid parameters", nil)
	}

	// 保存到Redis
	key := ipBindingKeyPrefix + deviceID
	err := s.client.Set(ctx, key, ip, defaultExpiration)
	if err != nil {
		RecordError()
		return errors.NewError(codes.StoreErrSet, "failed to save IP binding", err)
	}

	RecordIPBinding()
	return nil
}

// GetIPBinding 获取IP绑定
func (s *RedisStore) GetIPBinding(ctx context.Context, deviceID string) (string, error) {
	if deviceID == "" {
		return "", errors.NewError(codes.InvalidParams, "device ID is required", nil)
	}

	// 从Redis获取
	key := ipBindingKeyPrefix + deviceID
	ip, err := s.client.Get(ctx, key)
	if err != nil {
		return "", errors.NewError(codes.StoreErrGet, "failed to get IP binding", err)
	}

	// 处理空值
	if ip == "" {
		return "", errors.NewError(codes.StoreErrNotFound, "IP binding not found", nil)
	}

	return ip, nil
}

// DeleteBinding 删除绑定
func (s *RedisStore) DeleteBinding(ctx context.Context, deviceID string) error {
	if deviceID == "" {
		return errors.NewError(codes.InvalidParams, "device ID is required", nil)
	}

	// 删除设备绑定
	deviceKey := deviceBindingKeyPrefix + deviceID
	_, err := s.client.Del(ctx, deviceKey)
	if err != nil {
		return errors.NewError(codes.StoreErrDelete, "failed to delete device binding", err)
	}

	// 删除IP绑定
	ipKey := ipBindingKeyPrefix + deviceID
	_, err = s.client.Del(ctx, ipKey)
	if err != nil {
		return errors.NewError(codes.StoreErrDelete, "failed to delete IP binding", err)
	}

	return nil
}

// Close 关闭存储
func (s *RedisStore) Close() error {
	return s.client.Close()
}
