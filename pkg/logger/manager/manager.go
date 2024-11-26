package manager

import (
	"fmt"
	"sync"

	"gobase/pkg/logger"
	"gobase/pkg/logger/types"
)

// LoggerManager 日志管理器
type LoggerManager struct {
	factory *logger.Factory
	loggers map[string]types.Logger
	mu      sync.RWMutex
}

// NewLoggerManager 创建新的日志管理器
func NewLoggerManager() *LoggerManager {
	return &LoggerManager{
		factory: &logger.Factory{},
		loggers: make(map[string]types.Logger),
	}
}

// GetOrCreate 获取或创建日志实例
func (m *LoggerManager) GetOrCreate(name string, cfg types.Config) (types.Logger, error) {
	m.mu.RLock()
	if logger, exists := m.loggers[name]; exists {
		m.mu.RUnlock()
		return logger, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// 双重检查
	if logger, exists := m.loggers[name]; exists {
		return logger, nil
	}

	logger, err := m.factory.NewLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger %s: %w", name, err)
	}

	m.loggers[name] = logger
	return logger, nil
}

// Get 获取已存在的日志实例
func (m *LoggerManager) Get(name string) (types.Logger, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	logger, ok := m.loggers[name]
	return logger, ok
}

// Remove 移除日志实例
func (m *LoggerManager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.loggers, name)
}
