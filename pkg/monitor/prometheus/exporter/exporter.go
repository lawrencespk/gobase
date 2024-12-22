package exporter

import (
	"context"
	"encoding/json"
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
	dto "github.com/prometheus/client_model/go"
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
	systemCollector   *collector.SystemCollector

	// HTTP服务
	server     *http.Server
	once       sync.Once
	stopChan   chan struct{}
	metricPath string

	// 自定义注册表
	registry *prometheus.Registry
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
		case "system":
			e.systemCollector = collector.NewSystemCollector()
		}
	}
}

// Start 启动指标导出服务
func (e *Exporter) Start(ctx context.Context) error {
	var startErr error
	e.once.Do(func() {
		// 创建新的注册表
		e.registry = prometheus.NewRegistry()

		// 注册所有收集器
		if err := e.registerCollectors(); err != nil {
			startErr = err
			return
		}

		// 创建 HTTP 处理器
		mux := http.NewServeMux()

		// 使用自定义注册表创建处理器
		handler := promhttp.HandlerFor(e.registry, promhttp.HandlerOpts{})
		mux.Handle(e.metricPath, handler)

		// 添加查询处理器
		mux.Handle("/api/v1/query", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			query := r.URL.Query().Get("query")
			if query == "" {
				http.Error(w, "missing query parameter", http.StatusBadRequest)
				return
			}

			// 从注册表中获取指标
			gatheredMetrics, err := e.registry.Gather()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// 查找匹配的指标
			var matchedMetric *dto.Metric
			var matchedValue float64
			for _, mf := range gatheredMetrics {
				if *mf.Name == query {
					if len(mf.Metric) > 0 && mf.Metric[0].Counter != nil {
						matchedMetric = mf.Metric[0]
						matchedValue = *matchedMetric.Counter.Value
						break
					}
				}
			}

			// 构造标签映射
			labels := make(map[string]string)
			if matchedMetric != nil {
				for _, label := range matchedMetric.Label {
					labels[*label.Name] = *label.Value
				}
			}

			// 构造 Prometheus 格式的响应
			result := map[string]interface{}{
				"data": map[string]interface{}{
					"resultType": "vector",
					"result": []map[string]interface{}{
						{
							"metric": labels,
							"value": []interface{}{
								float64(time.Now().Unix()),
								fmt.Sprintf("%g", matchedValue),
							},
						},
					},
				},
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(result); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}))

		e.server = &http.Server{
			Addr:    fmt.Sprintf(":%d", e.cfg.Port),
			Handler: mux,
		}

		// 启动服务器
		e.log.Infof(ctx, "Starting prometheus metrics server on port %d", e.cfg.Port)
		go func() {
			if err := e.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				e.log.Errorf(ctx, "Prometheus metrics server error: %v", err)
			}
		}()

		// 启动系统指标收集
		if e.runtimeCollector != nil || e.systemCollector != nil {
			go e.collectSystemMetrics(ctx)
		}

		// 等待服务器启动
		time.Sleep(100 * time.Millisecond)
	})

	return startErr
}

// Stop 停止指标导出服务
func (e *Exporter) Stop(ctx context.Context) error {
	close(e.stopChan)
	if e.server != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		return e.server.Shutdown(shutdownCtx)
	}
	return nil
}

// registerCollectors 注册所有收集器
func (e *Exporter) registerCollectors() error {
	// 使用自定义注册表而不是默认注册表
	if e.httpCollector != nil {
		if err := e.registry.Register(e.httpCollector); err != nil {
			return errors.Wrap(err, "failed to register http collector")
		}
	}

	if e.runtimeCollector != nil {
		if err := e.registry.Register(e.runtimeCollector); err != nil {
			return errors.Wrap(err, "failed to register runtime collector")
		}
	}

	if e.resourceCollector != nil {
		if err := e.registry.Register(e.resourceCollector); err != nil {
			return errors.Wrap(err, "failed to register resource collector")
		}
	}

	if e.businessCollector != nil {
		if err := e.registry.Register(e.businessCollector); err != nil {
			return errors.Wrap(err, "failed to register business collector")
		}
	}

	if e.systemCollector != nil {
		if err := e.registry.Register(e.systemCollector); err != nil {
			return errors.Wrap(err, "failed to register system collector")
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

	// 确保路径以 "/" 开头
	if cfg.Path[0] != '/' {
		cfg.Path = "/" + cfg.Path
	}

	return nil
}
