package unit_test

import (
	"testing"

	"gobase/pkg/auth/jwt/binding"
	"gobase/pkg/monitor/prometheus/collector"

	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	t.Run("初始化指标", func(t *testing.T) {
		// 确保可以多次调用初始化
		binding.InitMetrics()
		binding.InitMetrics() // 第二次调用不应该panic
	})

	t.Run("记录指标", func(t *testing.T) {
		// 确保记录指标方法不会panic
		assert.NotPanics(t, func() {
			binding.RecordDeviceBinding()
			binding.RecordIPBinding()
			binding.RecordError()
		})
	})

	t.Run("获取收集器", func(t *testing.T) {
		c := binding.GetCollector()
		assert.NotNil(t, c)
		// 直接检查类型是否匹配
		assert.IsType(t, &collector.BusinessCollector{}, c, "应该返回 BusinessCollector 类型")
	})
}
