package exporter

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"gobase/pkg/errors"
	loggerTypes "gobase/pkg/logger/types"
	"gobase/pkg/monitor/prometheus/collector"
	configTypes "gobase/pkg/monitor/prometheus/config/types"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Exporter Prometheus指标导出器
type Exporter struct {
	cfg *configTypes.Config
	log loggerTypes.Logger

	// 收集器
	httpCollector     *collector.HTTPCollector
	runtimeCollector  *collector.RuntimeCollector
	resourceCollector *collector.ResourceCollector
	businessCollector *collector.BusinessCollector

	// HTTP服务
	server     *http.Server
	once       sync.Once
	stopChan   chan struct{}
	metricPath string
}

// New 创建导出器
func New(cfg *configTypes.Config, log loggerTypes.Logger) (*Exporter, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	e := &Exporter{
		cfg:        cfg,
		log:        log,
		stopChan:   make(chan struct{}),
		metricPath: cfg.Path,
	}

	// 初始化收集器
	e.initCollectors()

	return e, nil
}

// initCollectors 初始化所有收集器
func (e *Exporter) initCollectors() {
	namespace := e.cfg.Labels["app"]

	for _, c := range e.cfg.Collectors {
		switch c {
		case "http":
			e.httpCollector = collector.NewHTTPCollector(namespace)
		case "runtime":
			e.runtimeCollector = collector.NewRuntimeCollector(namespace)
		case "resource":
			e.resourceCollector = collector.NewResourceCollector(namespace)
		case "business":
			e.businessCollector = collector.NewBusinessCollector(namespace)
		}
	}
}

// Start 启动指标导出服务
func (e *Exporter) Start(ctx context.Context) error {
	var err error
	e.once.Do(func() {
		// 注册所有收集器
		if err = e.registerCollectors(); err != nil {
			return
		}

		// 创建HTTP服务
		mux := http.NewServeMux()
		mux.Handle(e.metricPath, promhttp.Handler())

		e.server = &http.Server{
			Addr:    fmt.Sprintf(":%d", e.cfg.Port),
			Handler: mux,
		}

		// 启动服务
		e.log.Infof(ctx, "Starting prometheus metrics server on port %d", e.cfg.Port)
		if err := e.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			e.log.Errorf(ctx, "Prometheus metrics server error: %v", err)
		}

		// 启动系统指标收集
		if e.runtimeCollector != nil {
			go e.collectSystemMetrics(ctx)
		}
	})

	return err
}

// Stop 停止指标导出服务
func (e *Exporter) Stop(ctx context.Context) error {
	close(e.stopChan)
	if e.server != nil {
		return e.server.Shutdown(ctx)
	}
	return nil
}

// registerCollectors 注册所有收集器
func (e *Exporter) registerCollectors() error {
	if e.httpCollector != nil {
		if err := e.httpCollector.Register(); err != nil {
			return errors.Wrap(err, "failed to register http collector")
		}
	}

	if e.runtimeCollector != nil {
		if err := e.runtimeCollector.Register(); err != nil {
			return errors.Wrap(err, "failed to register runtime collector")
		}
	}

	if e.resourceCollector != nil {
		if err := e.resourceCollector.Register(); err != nil {
			return errors.Wrap(err, "failed to register resource collector")
		}
	}

	if e.businessCollector != nil {
		if err := e.businessCollector.Register(); err != nil {
			return errors.Wrap(err, "failed to register business collector")
		}
	}

	return nil
}

// collectSystemMetrics 定期收集系统指标
func (e *Exporter) collectSystemMetrics(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-e.stopChan:
			return
		case <-ticker.C:
			// 创建一个临时通道来收集指标
			ch := make(chan prometheus.Metric, 100)
			go func() {
				e.runtimeCollector.Collect(ch)
				close(ch)
			}()

			// 消费收集到的指标
			for range ch {
				// 指标已经被收集器处理，这里不需要额外处理
			}
		}
	}
}

// GetHTTPCollector 获取HTTP收集器
func (e *Exporter) GetHTTPCollector() *collector.HTTPCollector {
	return e.httpCollector
}

// GetBusinessCollector 获取业务收集器
func (e *Exporter) GetBusinessCollector() *collector.BusinessCollector {
	return e.businessCollector
}

// validateConfig 验证配置
func validateConfig(cfg *configTypes.Config) error {
	if cfg == nil {
		return errors.NewConfigError("prometheus config is nil", nil)
	}

	if !cfg.Enabled {
		return nil
	}

	if cfg.Port <= 0 || cfg.Port > 65535 {
		return errors.NewConfigError(fmt.Sprintf("invalid port: %d", cfg.Port), nil)
	}

	if cfg.Path == "" {
		return errors.NewConfigError("metrics path cannot be empty", nil)
	}

	return nil
}
