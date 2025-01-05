package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/auth/jwt/crypto"
	"gobase/pkg/auth/jwt/security"
	"gobase/pkg/logger"
	"gobase/pkg/monitor/prometheus/metrics"
)

func TestKeyRotator_Start(t *testing.T) {
	ctx := context.Background()
	log, err := logger.NewLogger()
	require.NoError(t, err)

	// 正确初始化 KeyManager
	keyManager, err := crypto.NewKeyManager(jwt.RS256, log)
	require.NoError(t, err)

	// 初始化密钥
	err = keyManager.InitializeKeys(ctx, nil) // 使用自动生成的密钥
	require.NoError(t, err)

	policy := &security.Policy{
		EnableRotation:   true,
		RotationInterval: time.Millisecond * 100,
		Metrics:          metrics.DefaultJWTMetrics,
	}

	rotator, err := security.NewKeyRotator(keyManager, policy, log)
	require.NoError(t, err)

	t.Run("正常轮换", func(t *testing.T) {
		err := rotator.Start(ctx)
		require.NoError(t, err)

		// 等待至少一次轮换
		time.Sleep(time.Millisecond * 150)
		rotator.Stop()
	})

	t.Run("轮换失败", func(t *testing.T) {
		// 创建一个新的带有错误的 KeyManager
		failingManager, err := crypto.NewKeyManager(jwt.HS256, log) // HMAC 不支持轮换
		require.NoError(t, err)

		err = failingManager.InitializeKeys(ctx, &jwt.KeyConfig{
			SecretKey: "test-secret",
		})
		require.NoError(t, err)

		failingRotator, err := security.NewKeyRotator(failingManager, policy, log)
		require.NoError(t, err)

		err = failingRotator.Start(ctx)
		require.NoError(t, err)

		// 等待至少一次轮换尝试
		time.Sleep(time.Millisecond * 150)
		failingRotator.Stop()
	})

	t.Run("未启用轮换", func(t *testing.T) {
		disabledPolicy := &security.Policy{
			EnableRotation:   false,
			RotationInterval: time.Millisecond * 100,
			Metrics:          metrics.DefaultJWTMetrics,
		}

		disabledRotator, err := security.NewKeyRotator(keyManager, disabledPolicy, log)
		require.NoError(t, err)

		err = disabledRotator.Start(ctx)
		require.NoError(t, err)

		// 等待一段时间
		time.Sleep(time.Millisecond * 150)
		disabledRotator.Stop()
	})
}

func TestKeyRotator_Stop(t *testing.T) {
	log, err := logger.NewLogger()
	require.NoError(t, err)

	keyManager, err := crypto.NewKeyManager(jwt.RS256, log)
	require.NoError(t, err)

	err = keyManager.InitializeKeys(context.Background(), nil)
	require.NoError(t, err)

	policy := &security.Policy{
		EnableRotation:   true,
		RotationInterval: time.Millisecond * 100,
		Metrics:          metrics.DefaultJWTMetrics,
	}

	rotator, err := security.NewKeyRotator(keyManager, policy, log)
	require.NoError(t, err)

	t.Run("停止轮换", func(t *testing.T) {
		err := rotator.Start(context.Background())
		require.NoError(t, err)

		// 等待至少一次轮换
		time.Sleep(time.Millisecond * 150)
		rotator.Stop()

		// 等待一段时间确认没有新的轮换
		time.Sleep(time.Millisecond * 150)
	})
}
