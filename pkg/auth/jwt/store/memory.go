package store

import (
	"context"
	"sync"
	"time"

	"gobase/pkg/auth/jwt"
	"gobase/pkg/errors"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/collector"
	"gobase/pkg/trace/jaeger"
)

// MemoryStore 内存实现的Token存储
type MemoryStore struct {
	// token -> TokenInfo 映射
	tokens map[string]*jwt.TokenInfo
	// userID -> []token 映射
	userTokens map[string]map[string]struct{}
	mutex      sync.RWMutex
	logger     types.Logger
	metrics    *collector.BusinessCollector
	stopChan   chan struct{}
}

// NewMemoryStore 创建内存存储实例
func NewMemoryStore(opts Options) (*MemoryStore, error) {
	// 创建logger选项
	logOpts := logrus.DefaultOptions()
	logOpts.Level = types.InfoLevel
	logOpts.OutputPaths = []string{"stdout"}
	logOpts.Development = true

	// 创建文件管理器选项
	fileOpts := logrus.FileOptions{
		BufferSize:    32 * 1024,   // 32KB 缓冲区
		FlushInterval: time.Second, // 1秒刷新间隔
		MaxOpenFiles:  100,         // 最大打开文件数
		DefaultPath:   "app.log",   // 默认日志文件路径
	}
	fileManager := logrus.NewFileManager(fileOpts)

	// 创建队列配置
	queueConfig := logrus.QueueConfig{
		MaxSize:         1000,             // 队列最大大小
		BatchSize:       100,              // 批处理大小
		Workers:         1,                // 工作协程数量
		FlushInterval:   time.Second,      // 刷新间隔
		RetryCount:      3,                // 重试次数
		RetryInterval:   time.Second,      // 重试间隔
		MaxBatchWait:    time.Second * 5,  // 最大批处理等待时间
		ShutdownTimeout: time.Second * 10, // 关闭超时时间
	}

	log, err := logrus.NewLogger(fileManager, queueConfig, logOpts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create logger")
	}

	store := &MemoryStore{
		tokens:     make(map[string]*jwt.TokenInfo),
		userTokens: make(map[string]map[string]struct{}),
		logger:     log,
		stopChan:   make(chan struct{}),
	}

	// 初始化监控指标
	if opts.EnableMetrics {
		metrics := collector.NewBusinessCollector("jwt_store")
		if err := metrics.Register(); err != nil {
			return nil, errors.Wrap(err, "failed to register metrics collector")
		}
		store.metrics = metrics
	}

	// 启动清理协程
	if opts.CleanupInterval > 0 {
		go store.cleanup(opts.CleanupInterval)
	}

	return store, nil
}

// Save 保存Token信息
func (s *MemoryStore) Save(ctx context.Context, token string, tokenInfo *jwt.TokenInfo) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "store.memory.save")
	if span != nil {
		defer span.Finish()
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		if s.metrics != nil {
			s.metrics.ObserveOperation("save", duration, nil)
		}
	}()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查Token是否已过期
	if time.Now().After(tokenInfo.ExpiresAt) {
		s.logger.WithFields(types.Field{
			Key:   "token",
			Value: token,
		}).Warn(ctx, "token already expired")
		return errors.NewRedisKeyExpiredError("token expired", nil)
	}

	// 保存Token信息
	s.tokens[token] = tokenInfo

	// 保存用户Token映射
	userID := tokenInfo.Claims.GetUserID()
	if _, exists := s.userTokens[userID]; !exists {
		s.userTokens[userID] = make(map[string]struct{})
	}
	s.userTokens[userID][token] = struct{}{}

	s.logger.WithFields(
		types.Field{Key: "token", Value: token},
		types.Field{Key: "user_id", Value: userID},
	).Debug(ctx, "token info saved")

	return nil
}

// Get 获取Token信息
func (s *MemoryStore) Get(ctx context.Context, token string) (*jwt.TokenInfo, error) {
	span, ctx := jaeger.StartSpanFromContext(ctx, "store.memory.get")
	if span != nil {
		defer span.Finish()
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	tokenInfo, exists := s.tokens[token]
	if !exists {
		s.logger.WithFields(
			types.Field{Key: "token", Value: token},
		).Debug(ctx, "token not found")
		return nil, errors.NewRedisKeyNotFoundError("token not found", nil)
	}

	// 检查Token是否已过期
	if time.Now().After(tokenInfo.ExpiresAt) {
		s.logger.WithFields(
			types.Field{Key: "token", Value: token},
		).Debug(ctx, "token expired")
		return nil, errors.NewRedisKeyExpiredError("token expired", nil)
	}

	return tokenInfo, nil
}

// Delete 删除Token信息
func (s *MemoryStore) Delete(ctx context.Context, token string) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "store.memory.delete")
	if span != nil {
		defer span.Finish()
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	tokenInfo, exists := s.tokens[token]
	if !exists {
		s.logger.WithFields(
			types.Field{Key: "token", Value: token},
		).Debug(ctx, "token not found when deleting")
		return nil
	}

	// 从用户Token映射中删除
	userID := tokenInfo.Claims.GetUserID()
	if userTokens, exists := s.userTokens[userID]; exists {
		delete(userTokens, token)
		// 如果用户没有其他Token，删除用户映射
		if len(userTokens) == 0 {
			delete(s.userTokens, userID)
		}
	}

	// 删除Token
	delete(s.tokens, token)

	s.logger.WithFields(
		types.Field{Key: "token", Value: token},
		types.Field{Key: "user_id", Value: userID},
	).Debug(ctx, "token deleted")

	return nil
}

// DeleteByUserID 删除用户的所有Token
func (s *MemoryStore) DeleteByUserID(ctx context.Context, userID string) error {
	span, ctx := jaeger.StartSpanFromContext(ctx, "store.memory.delete_by_user_id")
	if span != nil {
		defer span.Finish()
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	userTokens, exists := s.userTokens[userID]
	if !exists {
		s.logger.WithFields(
			types.Field{Key: "user_id", Value: userID},
		).Debug(ctx, "no tokens found for user")
		return nil
	}

	// 删除所有Token
	for token := range userTokens {
		delete(s.tokens, token)
	}

	// 删除用户Token映射
	delete(s.userTokens, userID)

	s.logger.WithFields(
		types.Field{Key: "user_id", Value: userID},
		types.Field{Key: "token_count", Value: len(userTokens)},
	).Debug(ctx, "user tokens deleted")

	return nil
}

// Close 关闭存储
func (s *MemoryStore) Close() error {
	close(s.stopChan)
	return nil
}

// cleanup 定期清理过期的Token
func (s *MemoryStore) cleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupExpiredTokens()
		case <-s.stopChan:
			return
		}
	}
}

// cleanupExpiredTokens 清理过期的Token
func (s *MemoryStore) cleanupExpiredTokens() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	for token, info := range s.tokens {
		if now.After(info.ExpiresAt) {
			// 从用户Token映射中删除
			if userTokens, exists := s.userTokens[info.Claims.GetUserID()]; exists {
				delete(userTokens, token)
				// 如果用户没有其他Token，删除用户映射
				if len(userTokens) == 0 {
					delete(s.userTokens, info.Claims.GetUserID())
				}
			}
			// 删除Token
			delete(s.tokens, token)
		}
	}
}
