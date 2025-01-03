package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/collector"
	"gobase/pkg/trace/jaeger"

	"github.com/go-redis/redis/v8"
)

// DisableConnectionCheck 用于测试时禁用连接检查
var DisableConnectionCheck bool

// DisableTracing 用于测试时禁用追踪
var DisableTracing bool

// client Redis客户端实现
type client struct {
	client  redis.UniversalClient
	logger  types.Logger
	options *Options
	metrics *collector.RedisCollector
	tracer  *jaeger.Provider
	pool    Pool
}

// NewClient 创建一个新的Redis客户端
func NewClient(opts ...Option) (Client, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	// 验证连接池设置
	if options.PoolSize < 0 {
		return nil, errors.NewRedisInvalidConfigError("invalid pool size", nil)
	}
	if options.MinIdleConns < 0 {
		return nil, errors.NewRedisInvalidConfigError("invalid min idle connections", nil)
	}
	if options.IdleTimeout <= 0 {
		return nil, errors.NewRedisInvalidConfigError("invalid idle timeout", nil)
	}

	// 基本验证
	if len(options.Addresses) == 0 {
		return nil, errors.NewRedisInvalidConfigError("redis address is required", nil)
	}

	// 验证数据库编号
	if options.DB < 0 {
		return nil, errors.NewInvalidParamsError("invalid database number", nil)
	}

	var rdb redis.UniversalClient
	if options.EnableCluster {
		// 集群模式配置
		clusterOpts := &redis.ClusterOptions{
			Addrs:         options.Addresses,
			Username:      options.Username,
			Password:      options.Password,
			MaxRetries:    options.MaxRetries,
			PoolSize:      options.PoolSize,
			MinIdleConns:  options.MinIdleConns,
			IdleTimeout:   options.IdleTimeout,
			PoolTimeout:   options.PoolTimeout,
			DialTimeout:   options.DialTimeout,
			ReadTimeout:   options.ReadTimeout,
			WriteTimeout:  options.WriteTimeout,
			RouteRandomly: options.RouteRandomly,
		}

		// 配置TLS
		if options.EnableTLS {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: options.SkipVerify,
			}
			if options.TLSCertFile != "" && options.TLSKeyFile != "" {
				cert, err := tls.LoadX509KeyPair(options.TLSCertFile, options.TLSKeyFile)
				if err != nil {
					return nil, errors.NewRedisInvalidConfigError("failed to load TLS certificate", err)
				}
				tlsConfig.Certificates = []tls.Certificate{cert}
			}
			clusterOpts.TLSConfig = tlsConfig
		}

		rdb = redis.NewClusterClient(clusterOpts)
	} else {
		// 单机模式配置
		redisOpts := &redis.Options{
			Addr:         options.Addresses[0],
			Username:     options.Username,
			Password:     options.Password,
			DB:           options.DB,
			MaxRetries:   options.MaxRetries,
			PoolSize:     options.PoolSize,
			MinIdleConns: options.MinIdleConns,
			IdleTimeout:  options.IdleTimeout,
			PoolTimeout:  options.PoolTimeout,
			DialTimeout:  options.DialTimeout,
			ReadTimeout:  options.ReadTimeout,
			WriteTimeout: options.WriteTimeout,
		}

		// 配置TLS
		if options.EnableTLS {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: options.SkipVerify,
			}

			if options.TLSCertFile != "" && options.TLSKeyFile != "" {
				cert, err := tls.LoadX509KeyPair(options.TLSCertFile, options.TLSKeyFile)
				if err != nil {
					return nil, errors.NewRedisInvalidConfigError("failed to load TLS certificate", err)
				}
				tlsConfig.Certificates = []tls.Certificate{cert}
			}

			redisOpts.TLSConfig = tlsConfig
		}

		rdb = redis.NewClient(redisOpts)
	}

	// 仅在非测试模式下验证连接
	if !DisableConnectionCheck {
		ctx := context.Background()
		err := withRetry(ctx, options, func() error {
			return rdb.Ping(ctx).Err()
		})
		if err != nil {
			options.Logger.WithError(err).Error(ctx, "failed to connect to redis")
			return nil, errors.NewRedisConnError("failed to connect to redis", err)
		}
	}

	// 初始化Redis监控指标收集器
	var metrics *collector.RedisCollector
	if options.EnableMetrics {
		metrics = collector.NewRedisCollector(options.MetricsNamespace)
	}

	// 创建客户端
	c := &client{
		client:  rdb,
		logger:  options.Logger,
		options: options,
		metrics: metrics,
	}

	// 初始化连接池
	pool := NewPool(rdb, options.Logger, options, metrics)
	if pool == nil {
		return nil, errors.NewRedisInvalidConfigError("failed to create pool", nil)
	}
	c.pool = pool

	// 等待连接池初始化完成
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		stats := c.client.PoolStats()
		if stats.TotalConns >= uint32(options.PoolSize) {
			return c, nil
		}
		// 发送多个 ping 来建立连接
		var wg sync.WaitGroup
		for j := 0; j < options.PoolSize; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				_ = c.client.Ping(ctx)
			}()
		}
		wg.Wait()
		time.Sleep(100 * time.Millisecond)
	}

	// 初始化 tracer
	if options.EnableTracing && !DisableTracing {
		tracer, err := jaeger.NewProvider()
		if err != nil {
			return nil, errors.NewRedisInvalidConfigError("failed to create tracer", err)
		}
		c.tracer = tracer
	}

	return c, nil
}

// PoolStats 获取连接池统计信息
func (c *client) PoolStats() *PoolStats {
	if c.pool != nil {
		return c.pool.Stats()
	}
	return nil
}

// Close 关闭客户端连接
func (c *client) Close() error {
	if c.pool != nil {
		return c.pool.Close()
	}
	return nil
}

// NewClientFromConfig 从配置创建Redis客户端
func NewClientFromConfig(cfg *Config) (Client, error) {
	if cfg == nil {
		return nil, errors.NewRedisInvalidConfigError("config is required", nil)
	}

	// 将 Config 转换为 Options
	opts := []Option{
		WithAddresses(cfg.Addresses),
		WithUsername(cfg.Username),
		WithPassword(cfg.Password),
		WithDB(cfg.Database),
		WithPoolSize(cfg.PoolSize),
		WithMinIdleConns(cfg.MinIdleConns),
		WithMaxRetries(cfg.MaxRetries),
		WithDialTimeout(cfg.DialTimeout),
		WithReadTimeout(cfg.ReadTimeout),
		WithWriteTimeout(cfg.WriteTimeout),
	}

	// TLS配置
	if cfg.EnableTLS {
		opts = append(opts,
			WithTLS(true),
			WithTLSCert(cfg.TLSCertFile),
			WithTLSKey(cfg.TLSKeyFile),
		)
	}

	// 集群配置
	if cfg.EnableCluster {
		opts = append(opts,
			WithCluster(true),
			WithRouteRandomly(cfg.RouteRandomly),
		)
	}

	// 监控配置
	if cfg.EnableMetrics {
		opts = append(opts,
			WithMetrics(true),
			WithMetricsNamespace(cfg.MetricsNamespace),
		)
	}
	if cfg.EnableTracing {
		opts = append(opts, WithTracing(true))
	}

	// 使用已有的 NewClient 函数创建客户端
	return NewClient(opts...)
}

// withReconnect 尝试重新连接
func (c *client) withReconnect(ctx context.Context) error {
	maxRetries := c.options.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 5
	}

	backoff := c.options.RetryBackoff
	if backoff <= 0 {
		backoff = 2 * time.Second
	}

	for i := 0; i <= maxRetries; i++ {
		// 1. 先关闭现有连接
		if c.pool != nil {
			_ = c.client.Close()
			time.Sleep(time.Second)
		}

		// 2. 尝试重新建立连接
		err := c.withOperation(ctx, "Ping", func() error {
			return c.client.Ping(ctx).Err()
		})
		if err == nil {
			return nil
		}

		if i == maxRetries {
			return errors.NewRedisConnError("failed to reconnect", err)
		}

		// 3. 使用指数退避策略
		backoffDuration := backoff * time.Duration(1<<uint(i))
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoffDuration):
			continue
		}
	}

	return nil
}

// Do 方法修改
func (c *client) Do(ctx context.Context, cmd string, args ...interface{}) (interface{}, error) {
	span, ctx := startSpan(ctx, c.tracer, "redis."+cmd)
	if span != nil {
		defer span.Finish()
		span.SetTag("db.statement", fmt.Sprintf("%s %v", cmd, args))
	}

	startTime := time.Now()
	var result interface{}

	// 直接在 withRetry 中声明和赋值 err
	err := withRetry(ctx, c.options, func() error {
		cmdArgs := make([]interface{}, 0, len(args)+1)
		cmdArgs = append(cmdArgs, cmd)
		cmdArgs = append(cmdArgs, args...)

		r := c.client.Do(ctx, cmdArgs...)
		if r != nil {
			if err := r.Err(); err != nil && isNetworkError(err) {
				// 尝试重连
				if reconnErr := c.withReconnect(ctx); reconnErr != nil {
					return reconnErr
				}
				return err // 返回原始错误以触发重试
			}
			result = r
			return r.Err()
		}
		return nil
	})

	duration := time.Since(startTime)

	if c.metrics != nil {
		c.metrics.ObserveCommandExecution(duration.Seconds(), err)
	}

	if err != nil && span != nil {
		span.SetError(err)
	}

	return result, err
}
