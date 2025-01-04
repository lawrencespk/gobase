package benchmark

import (
	"context"
	"fmt"
	"os"
	"testing"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/binding"
	"gobase/pkg/client/redis"

	"github.com/prometheus/client_golang/prometheus"
)

func TestMain(m *testing.M) {
	// 在所有测试开始前初始化 metrics
	binding.InitMetrics()
	if err := binding.RegisterCollector(); err != nil {
		fmt.Printf("Failed to register metrics collector: %v\n", err)
		os.Exit(1)
	}

	// 运行测试
	code := m.Run()

	// 清理 metrics
	prometheus.Unregister(binding.GetCollector())

	os.Exit(code)
}

func BenchmarkDeviceValidator(b *testing.B) {
	// 初始化 Redis 客户端
	redisClient, err := redis.NewClient(
		redis.WithAddresses([]string{"localhost:6379"}),
	)
	if err != nil {
		b.Fatal(err)
	}
	defer redisClient.Close()

	// 创建 Redis Store
	store, err := binding.NewRedisStore(redisClient)
	if err != nil {
		b.Fatal(err)
	}
	defer store.Close()

	// 创建设备验证器 - 不再在这里初始化 metrics
	validator, err := binding.NewDeviceValidator(store)
	if err != nil {
		b.Fatal(err)
	}

	// 准备测试数据
	ctx := context.Background()
	userID := "bench-user"
	deviceID := "bench-device"
	device := &binding.DeviceInfo{
		ID:          deviceID,
		Type:        "mobile",
		Name:        "Bench Device",
		OS:          "iOS",
		Browser:     "Safari",
		Fingerprint: "bench-fp",
	}

	claims := jwt.NewStandardClaims(
		jwt.WithUserID(userID),
		jwt.WithDeviceID(deviceID),
	)

	// 保存绑定关系
	err = store.SaveDeviceBinding(ctx, userID, deviceID, device)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = validator.ValidateDevice(ctx, claims, device)
		}
	})
}

func BenchmarkIPValidator(b *testing.B) {
	// 初始化 Redis 客户端
	redisClient, err := redis.NewClient(
		redis.WithAddresses([]string{"localhost:6379"}),
	)
	if err != nil {
		b.Fatal(err)
	}
	defer redisClient.Close()

	// 创建 Redis Store
	store, err := binding.NewRedisStore(redisClient)
	if err != nil {
		b.Fatal(err)
	}
	defer store.Close()

	// 创建 IP 验证器
	validator, err := binding.NewIPValidator(store)
	if err != nil {
		b.Fatal(err)
	}

	// 准备测试数据
	ctx := context.Background()
	userID := "bench-user"
	deviceID := "bench-device"
	ip := "127.0.0.1"

	claims := jwt.NewStandardClaims(
		jwt.WithUserID(userID),
		jwt.WithDeviceID(deviceID),
	)

	// 保存绑定关系
	err = store.SaveIPBinding(ctx, userID, deviceID, ip)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = validator.ValidateIP(ctx, claims, ip)
		}
	})
}

func BenchmarkRedisStore_SaveBinding(b *testing.B) {
	// 初始化 Redis 客户端
	redisClient, err := redis.NewClient(
		redis.WithAddress("localhost:6379"),
	)
	if err != nil {
		b.Fatal(err)
	}
	defer redisClient.Close()

	// 创建 Redis Store
	store, err := binding.NewRedisStore(redisClient)
	if err != nil {
		b.Fatal(err)
	}
	defer store.Close()

	ctx := context.Background()

	b.Run("SaveDeviceBinding", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				i++
				device := &binding.DeviceInfo{
					ID:          fmt.Sprintf("bench-device-%d", i),
					Type:        "mobile",
					Name:        fmt.Sprintf("Bench Device %d", i),
					OS:          "iOS",
					Browser:     "Safari",
					Fingerprint: fmt.Sprintf("bench-fp-%d", i),
				}
				_ = store.SaveDeviceBinding(ctx, fmt.Sprintf("user-%d", i), device.ID, device)
			}
		})
	})

	b.Run("SaveIPBinding", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				i++
				_ = store.SaveIPBinding(ctx, fmt.Sprintf("user-%d", i),
					fmt.Sprintf("device-%d", i), "127.0.0.1")
			}
		})
	})
}
