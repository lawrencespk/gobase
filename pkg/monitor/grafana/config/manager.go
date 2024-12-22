package config

import (
	"fmt"
	"path/filepath"
	"sync"

	"gobase/pkg/errors"
	"gobase/pkg/logger/types"
)

type Manager struct {
	logger     types.Logger
	configPath string
	templates  map[string]string
	mu         sync.RWMutex
}

func NewManager(configPath string, logger types.Logger) *Manager {
	return &Manager{
		logger:     logger,
		configPath: configPath,
		templates:  make(map[string]string),
	}
}

func (m *Manager) LoadTemplates() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 加载所有仪表盘模板
	dashboards, err := filepath.Glob(filepath.Join(m.configPath, "dashboard", "*.json"))
	if err != nil {
		return errors.NewConfigError(
			fmt.Sprintf("failed to load dashboard templates: %v", err),
			err,
		)
	}

	// 加载所有告警规则
	rules, err := filepath.Glob(filepath.Join(m.configPath, "alert", "*.yaml"))
	if err != nil {
		return errors.NewConfigError(
			fmt.Sprintf("failed to load alert rules: %v", err),
			err,
		)
	}

	// 加载 Alertmanager 配置
	alertmanager, err := filepath.Glob(filepath.Join(m.configPath, "alertmanager", "*.yaml"))
	if err != nil {
		return errors.NewConfigError(
			fmt.Sprintf("failed to load alertmanager config: %v", err),
			err,
		)
	}

	// 合并所有模板
	templates := append(dashboards, rules...)
	templates = append(templates, alertmanager...)

	for _, template := range templates {
		name := filepath.Base(template)
		m.templates[name] = template
	}

	return nil
}

func (m *Manager) GetTemplate(name string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	template, ok := m.templates[name]
	if !ok {
		return "", errors.NewConfigError(
			fmt.Sprintf("template not found: %s", name),
			nil,
		)
	}

	return template, nil
}
