package config

import (
	"context"
	"encoding/json"
	"fmt"
	"gobase/pkg/config"
	configTypes "gobase/pkg/config/types"
	"gobase/pkg/errors"
	loggerTypes "gobase/pkg/logger/types"

	"gopkg.in/yaml.v2"
)

type Manager struct {
	cfg    *configTypes.Config
	logger loggerTypes.Logger
}

func NewManager(cfg *configTypes.Config, logger loggerTypes.Logger) *Manager {
	return &Manager{
		cfg:    cfg,
		logger: logger,
	}
}

// GetDashboardConfig 获取仪表盘配置
func (m *Manager) GetDashboardConfig(name string) (map[string]interface{}, error) {
	var dashboardJSON string

	switch name {
	case "redis":
		dashboardJSON = m.cfg.Grafana.Dashboards.Redis
	case "http":
		dashboardJSON = m.cfg.Grafana.Dashboards.HTTP
	case "logger":
		dashboardJSON = m.cfg.Grafana.Dashboards.Logger
	case "runtime":
		dashboardJSON = m.cfg.Grafana.Dashboards.Runtime
	case "system":
		dashboardJSON = m.cfg.Grafana.Dashboards.System
	default:
		return nil, fmt.Errorf("unknown dashboard: %s", name)
	}

	var dashboard map[string]interface{}
	if err := json.Unmarshal([]byte(dashboardJSON), &dashboard); err != nil {
		return nil, fmt.Errorf("failed to parse dashboard JSON: %v", err)
	}

	return dashboard, nil
}

// GetAlertRules 获取告警规则
func (m *Manager) GetAlertRules() (map[string]interface{}, error) {
	rulesYAML := m.cfg.Grafana.Alerts.Rules
	loggerRulesYAML := m.cfg.Grafana.Alerts.Logger

	// 合并规则
	var rules, loggerRules map[string]interface{}

	if err := yaml.Unmarshal([]byte(rulesYAML), &rules); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal alert rules")
	}

	if err := yaml.Unmarshal([]byte(loggerRulesYAML), &loggerRules); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal logger alert rules")
	}

	// 合并规则组
	rulesGroups := rules["groups"].([]interface{})
	loggerGroups := loggerRules["groups"].([]interface{})
	allGroups := append(rulesGroups, loggerGroups...)

	return map[string]interface{}{
		"groups": allGroups,
	}, nil
}

// WatchConfig 监听配置变更
func (m *Manager) WatchConfig(ctx context.Context) error {
	// 使用 config 包提供的 Watch 函数
	return config.Watch(ctx, m.cfg, func() {
		m.logger.Info(ctx, "Grafana configuration updated", loggerTypes.Field{
			Key:   "component",
			Value: "grafana_config",
		})
	})
}
