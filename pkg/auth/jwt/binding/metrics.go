package binding

import (
	"gobase/pkg/monitor/prometheus/collector"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	bindingCollector *collector.BusinessCollector
	once             sync.Once
)

// InitMetrics 初始化指标收集器
func InitMetrics() {
	once.Do(func() {
		bindingCollector = collector.NewBusinessCollector("auth_binding")
	})
}

// RegisterCollector 注册指标收集器
func RegisterCollector() error {
	if bindingCollector != nil {
		return prometheus.Register(bindingCollector)
	}
	return nil
}

// GetCollector 获取指标收集器
func GetCollector() *collector.BusinessCollector {
	return bindingCollector
}

// RecordDeviceBinding 记录设备绑定指标
func RecordDeviceBinding() {
	if bindingCollector != nil {
		bindingCollector.ObserveOperation("device_binding", 0, nil)
	}
}

// RecordIPBinding 记录IP绑定指标
func RecordIPBinding() {
	if bindingCollector != nil {
		bindingCollector.ObserveOperation("ip_binding", 0, nil)
	}
}

// RecordError 记录错误指标
func RecordError() {
	if bindingCollector != nil {
		bindingCollector.ObserveOperation("binding_error", 0, nil)
	}
}
