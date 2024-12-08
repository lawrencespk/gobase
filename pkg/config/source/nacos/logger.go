package nacos

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	baseLogger "gobase/pkg/logger/types"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// NacosLogAdapter 适配Nacos的日志接口
type NacosLogAdapter struct {
	mu       sync.RWMutex
	logger   baseLogger.Logger
	writers  []io.WriteCloser
	instance *logrus.Logger
	lumber   *lumberjack.Logger
}

// NewNacosLogAdapter 创建新的日志适配器
func NewNacosLogAdapter(logger baseLogger.Logger, logDir string) *NacosLogAdapter {
	// 创建 lumberjack logger
	lumber := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, "nacos-sdk.log"),
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     7, // days
		Compress:   true,
	}

	adapter := &NacosLogAdapter{
		logger:   logger,
		writers:  make([]io.WriteCloser, 0),
		instance: logrus.New(),
		lumber:   lumber,
	}

	// 配置 logrus
	adapter.instance.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
	})
	adapter.instance.SetOutput(lumber)

	return adapter
}

// GetLogrus 获取logrus实例
func (a *NacosLogAdapter) GetLogrus() *logrus.Logger {
	return a.instance
}

// Close 关闭所有writers
func (a *NacosLogAdapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 1. 停止 logrus 输出
	a.instance.SetOutput(io.Discard)

	// 2. 等待确保没有正在进行的写操作
	time.Sleep(200 * time.Millisecond)

	// 3. 关闭 lumberjack
	if a.lumber != nil {
		// 强制执行一次日志轮转
		_ = a.lumber.Rotate()

		// 等待文件系统操作完成
		time.Sleep(100 * time.Millisecond)

		// 关闭 lumberjack
		if err := a.lumber.Close(); err != nil {
			return fmt.Errorf("failed to close lumberjack: %v", err)
		}

		// 清空引用
		a.lumber = nil
	}

	// 4. 关闭其他 writers
	for _, writer := range a.writers {
		_ = writer.Close()
	}
	a.writers = nil

	// 5. 在 Windows 上，显式运行 GC
	if runtime.GOOS == "windows" {
		runtime.GC()
	}

	return nil
}
