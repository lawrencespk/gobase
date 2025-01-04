package stress

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/binding"
	"gobase/pkg/client/redis"
)

func TestBinding_StressConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// 初始化Redis客户端
	redisClient, err := redis.NewClient(
		redis.WithAddress("localhost:6379"),
	)
	require.NoError(t, err)
	defer redisClient.Close()

	// 初始化存储
	store, err := binding.NewRedisStore(redisClient)
	require.NoError(t, err)
	defer store.Close()

	// 初始化验证器
	deviceValidator, err := binding.NewDeviceValidator(store)
	require.NoError(t, err)
	ipValidator, err := binding.NewIPValidator(store)
	require.NoError(t, err)

	// 测试参数
	const (
		numUsers    = 1000 // 用户数量
		numDevices  = 5    // 每个用户的设备数量
		numWorkers  = 50   // 并发工作协程数
		testTimeout = 2 * time.Minute
	)

	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()

	// 准备测试数据
	devices := make([]binding.DeviceInfo, 0, numUsers*numDevices)
	claims := make([]jwt.Claims, 0, numUsers*numDevices)

	for i := 0; i < numUsers; i++ {
		for j := 0; j < numDevices; j++ {
			userID := fmt.Sprintf("user-%d", i)
			deviceID := fmt.Sprintf("device-%d-%d", i, j)

			// 创建设备信息
			device := binding.DeviceInfo{
				ID:          deviceID,
				Type:        "mobile",
				Name:        fmt.Sprintf("Device-%d-%d", i, j),
				OS:          "iOS",
				Browser:     "Safari",
				Fingerprint: fmt.Sprintf("fp-%d-%d", i, j),
			}
			devices = append(devices, device)

			// 创建claims
			claim := jwt.NewStandardClaims(
				jwt.WithUserID(userID),
				jwt.WithDeviceID(deviceID),
				jwt.WithIPAddress("127.0.0.1"),
			)
			claims = append(claims, claim)

			// 保存绑定关系
			err := store.SaveDeviceBinding(ctx, userID, deviceID, &device)
			require.NoError(t, err)
			err = store.SaveIPBinding(ctx, userID, deviceID, "127.0.0.1")
			require.NoError(t, err)
		}
	}

	// 并发验证测试
	t.Run("ConcurrentDeviceValidation", func(t *testing.T) {
		var wg sync.WaitGroup
		errorCh := make(chan error, numWorkers)

		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < len(devices); i++ {
					err := deviceValidator.ValidateDevice(ctx, claims[i], &devices[i])
					if err != nil {
						errorCh <- fmt.Errorf("device validation failed: %w", err)
						return
					}
				}
			}()
		}

		wg.Wait()
		close(errorCh)

		for err := range errorCh {
			t.Error(err)
		}
	})

	t.Run("ConcurrentIPValidation", func(t *testing.T) {
		var wg sync.WaitGroup
		errorCh := make(chan error, numWorkers)

		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < len(claims); i++ {
					err := ipValidator.ValidateIP(ctx, claims[i], "127.0.0.1")
					if err != nil {
						errorCh <- fmt.Errorf("IP validation failed: %w", err)
						return
					}
				}
			}()
		}

		wg.Wait()
		close(errorCh)

		for err := range errorCh {
			t.Error(err)
		}
	})
}
