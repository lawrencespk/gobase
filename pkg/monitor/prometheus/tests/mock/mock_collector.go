package mock

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/mock"
)

// MockCollector 模拟收集器
type MockCollector struct {
	mock.Mock
}

func (m *MockCollector) Describe(ch chan<- *prometheus.Desc) {
	m.Called(ch)
}

func (m *MockCollector) Collect(ch chan<- prometheus.Metric) {
	m.Called(ch)
}

func (m *MockCollector) Register() error {
	args := m.Called()
	return args.Error(0)
}
