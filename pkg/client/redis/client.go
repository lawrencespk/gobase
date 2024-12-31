package redis

import (
	"context"
	"crypto/tls"

	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"
	"gobase/pkg/logger/types"

	"github.com/go-redis/redis/v8"
	"github.com/opentracing/opentracing-go"
)

// DisableConnectionCheck 用于测试时禁用连接检查
var DisableConnectionCheck bool

// client Redis客户端实现
type client struct {
	client  redis.UniversalClient
	logger  types.Logger
	tracer  opentracing.Tracer
	options *Options
}

// NewClient 创建一个新的Redis客户端
func NewClient(opts ...Option) (Client, error) {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	// 基本验证
	if len(options.Addresses) == 0 {
		return nil, errors.NewError(codes.CacheError, "redis address is required", nil)
	}

	// 验证数据库编号
	if options.DB < 0 {
		return nil, errors.NewError(codes.InvalidParams, "invalid database number", nil)
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
					return nil, errors.NewError(codes.CacheError, "failed to load TLS certificate", err)
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
					return nil, errors.NewError(codes.CacheError, "failed to load TLS certificate", err)
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
			return nil, errors.NewError(codes.CacheError, "failed to connect to redis", err)
		}
	}

	return &client{
		client:  rdb,
		logger:  options.Logger,
		tracer:  options.Tracer,
		options: options,
	}, nil
}

// NewClientFromConfig 从配置创建Redis客户端
func NewClientFromConfig(cfg *Config) (Client, error) {
	if cfg == nil {
		return nil, errors.NewError(codes.CacheError, "config is required", nil)
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
