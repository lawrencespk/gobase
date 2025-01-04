package stress

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"gobase/pkg/auth/jwt/blacklist"
	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
)

func TestBlacklistStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	// 创建日志记录器 (保留以供将来使用)
	log, err := logger.NewLogger(logger.WithLevel(types.ErrorLevel))
	require.NoError(t, err)
	_ = log // 暂时不使用，但保留以供将来扩展

	// 创建存储实例
	store := blacklist.NewMemoryStore()
	defer store.Close()

	ctx := context.Background()
	const (
		numWorkers    = 100
		numOperations = 1000
		expiration    = time.Minute
	)

	var wg sync.WaitGroup
	errCh := make(chan error, numWorkers*numOperations)

	// 启动工作协程
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for i := 0; i < numOperations; i++ {
				tokenID := fmt.Sprintf("token-%d-%d", workerID, i)
				reason := fmt.Sprintf("reason-%d-%d", workerID, i)

				// 添加token
				if err := store.Add(ctx, tokenID, reason, expiration); err != nil {
					errCh <- fmt.Errorf("worker %d add error: %v", workerID, err)
					continue
				}

				// 获取token
				if _, err := store.Get(ctx, tokenID); err != nil {
					errCh <- fmt.Errorf("worker %d get error: %v", workerID, err)
					continue
				}

				// 移除token
				if err := store.Remove(ctx, tokenID); err != nil {
					errCh <- fmt.Errorf("worker %d remove error: %v", workerID, err)
					continue
				}
			}
		}(w)
	}

	// 等待所有工作协程完成
	wg.Wait()
	close(errCh)

	// 检查错误
	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		t.Errorf("stress test encountered %d errors:", len(errors))
		for _, err := range errors {
			t.Error(err)
		}
	}
}
