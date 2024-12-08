package logger

import (
	"fmt"
	"io"
	"sync"

	baseLogger "gobase/pkg/logger/types"

	"github.com/sirupsen/logrus"
)

type NacosLogAdapter struct {
	mu       sync.RWMutex
	logger   baseLogger.Logger
	writers  []io.WriteCloser
	instance *logrus.Logger
}

func NewNacosLogAdapter(logger baseLogger.Logger) *NacosLogAdapter {
	adapter := &NacosLogAdapter{
		logger:   logger,
		writers:  make([]io.WriteCloser, 0),
		instance: logrus.New(),
	}

	// 配置logrus实例
	adapter.instance.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
	})

	return adapter
}

// Close 关闭所有writers
func (a *NacosLogAdapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var errs []error
	for _, writer := range a.writers {
		if err := writer.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	// 清空writers列表
	a.writers = a.writers[:0]

	if len(errs) > 0 {
		return fmt.Errorf("failed to close writers: %v", errs)
	}
	return nil
}
