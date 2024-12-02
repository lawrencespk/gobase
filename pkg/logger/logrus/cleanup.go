package logrus

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// CleanupConfig 清理配置
type CleanupConfig struct {
	Enable     bool          // 是否启用清理
	MaxBackups int           // 保留的旧日志文件个数
	MaxAge     int           // 日志文件的最大保留天数
	Interval   time.Duration // 清理检查间隔
}

// LogCleaner 日志清理器
type LogCleaner struct {
	config CleanupConfig
	done   chan struct{}
}

// NewLogCleaner 创建日志清理器
func NewLogCleaner(config CleanupConfig) *LogCleaner {
	return &LogCleaner{
		config: config,
		done:   make(chan struct{}),
	}
}

// Start 启动清理服务
func (c *LogCleaner) Start() {
	if !c.config.Enable {
		return
	}

	go c.run()
}

// Stop 停止清理服务
func (c *LogCleaner) Stop() {
	close(c.done)
}

// run 运行清理服务
func (c *LogCleaner) run() {
	ticker := time.NewTicker(c.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanupOldLogs()
		case <-c.done:
			return
		}
	}
}

// cleanupOldLogs 清理旧日志文件
func (c *LogCleaner) cleanupOldLogs() {
	for _, path := range defaultOptions.OutputPaths {
		if path == "stdout" || path == "stderr" {
			continue
		}

		dir := filepath.Dir(path)
		pattern := filepath.Join(dir, "*.log")

		files, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}

		// 按修改时间排序
		sort.Slice(files, func(i, j int) bool {
			fi, _ := os.Stat(files[i])
			fj, _ := os.Stat(files[j])
			return fi.ModTime().After(fj.ModTime())
		})

		// 删除多余的备份
		if len(files) > c.config.MaxBackups {
			for _, file := range files[c.config.MaxBackups:] {
				if err := os.Remove(file); err != nil {
					fmt.Printf("remove old log file error: %v\n", err)
				}
			}
		}

		// 删除过期的日志
		for _, file := range files {
			fi, err := os.Stat(file)
			if err != nil {
				continue
			}

			if time.Since(fi.ModTime()).Hours() > float64(c.config.MaxAge*24) {
				if err := os.Remove(file); err != nil {
					fmt.Printf("remove expired log file error: %v\n", err)
				}
			}
		}
	}
}
