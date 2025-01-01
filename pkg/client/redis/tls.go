package redis

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"gobase/pkg/errors"
)

// loadTLSConfig 加载TLS配置
func loadTLSConfig(opts *Options) (*tls.Config, error) {
	if !opts.EnableTLS {
		return nil, nil
	}

	// 验证TLS配置参数
	if opts.TLSCertFile == "" || opts.TLSKeyFile == "" {
		return nil, errors.NewRedisInvalidConfigError("TLS certificate and key files are required", nil)
	}

	// 加载证书
	cert, err := tls.LoadX509KeyPair(opts.TLSCertFile, opts.TLSKeyFile)
	if err != nil {
		return nil, errors.NewRedisInvalidConfigError("failed to load TLS certificate", err)
	}

	// 创建证书池
	caCertPool := x509.NewCertPool()
	caCert, err := os.ReadFile(opts.TLSCertFile)
	if err != nil {
		return nil, errors.NewRedisInvalidConfigError("failed to read CA certificate", err)
	}

	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, errors.NewRedisInvalidConfigError("failed to append CA certificate: invalid certificate format", nil)
	}

	// 返回TLS配置
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}
