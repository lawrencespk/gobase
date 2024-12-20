package integration

import (
	"context"
	"testing"
	"time"

	loggertypes "gobase/pkg/logger/types"
	configtypes "gobase/pkg/monitor/prometheus/config/types"
	"gobase/pkg/monitor/prometheus/tests/testutils"
)

type IntegrationTestSuite struct {
	prometheus *testutils.PrometheusContainer
	cfg        *configtypes.Config
	logger     loggertypes.Logger
	ctx        context.Context
	cancel     context.CancelFunc
}

func setupTestSuite(t *testing.T) *IntegrationTestSuite {
	suite := &IntegrationTestSuite{}

	// 创建上下文
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 30*time.Second)

	// 启动Prometheus容器
	prom, err := testutils.StartPrometheusContainer(t)
	if err != nil {
		t.Fatalf("启动Prometheus容器失败: %v", err)
	}
	suite.prometheus = prom

	// 初始化配置
	suite.cfg = &configtypes.Config{
		Enabled: true,
		Port:    9091,
		Path:    "/metrics",
		Labels: map[string]string{
			"app": "testapp",
		},
		Collectors: []string{
			"business",
			"http",
		},
	}

	// 初始化日志
	suite.logger = testutils.NewTestLogger(t)

	return suite
}
