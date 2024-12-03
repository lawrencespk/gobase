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
	LogPaths   []string      // 需要清理的日志路径
}

// LogCleaner 日志清理器
type LogCleaner struct {
	config CleanupConfig // 清理配置
	done   chan struct{} // 结束通道
}

// NewLogCleaner 创建日志清理器
func NewLogCleaner(config CleanupConfig) *LogCleaner {
	return &LogCleaner{
		config: config,              // 清理配置
		done:   make(chan struct{}), // 结束通道
	}
}

// Start 启动清理服务
func (c *LogCleaner) Start() {
	if !c.config.Enable { // 如果未启用清理
		return
	}

	go c.run()
}

// Stop 停止清理服务
func (c *LogCleaner) Stop() {
	close(c.done) // 关闭结束通道
}

// run 运行清理服务
func (c *LogCleaner) run() {
	ticker := time.NewTicker(c.config.Interval) // 创建定时器
	defer ticker.Stop()                         // 停止定时器

	for {
		select {
		case <-ticker.C: // 定时器触发
			c.cleanupOldLogs() // 清理旧日志文件
		case <-c.done: // 结束通道触发
			return
		}
	}
}

// cleanupOldLogs 清理旧日志文件
func (c *LogCleaner) cleanupOldLogs() {
	for _, path := range c.config.LogPaths { // 遍历日志路径
		if path == "stdout" || path == "stderr" { // 跳过标准输出和标准错误
			continue
		}

		pattern := filepath.Join(path, "*.log") // 获取日志文件模式
		files, err := filepath.Glob(pattern)    // 获取所有匹配的文件
		if err != nil {
			continue
		}

		// 按修改时间排序
		sort.Slice(files, func(i, j int) bool {
			fi, _ := os.Stat(files[i])              // 获取文件信息
			fj, _ := os.Stat(files[j])              // 获取文件信息
			return fi.ModTime().After(fj.ModTime()) // 比较修改时间
		})

		// 先筛选出未过期的文件
		validFiles := make([]string, 0)
		for _, file := range files {
			fi, err := os.Stat(file) // 获取文件信息
			if err != nil {          // 如果获取失败，跳过
				continue
			}

			// 如果文件未过期，加入有效文件列表
			if time.Since(fi.ModTime()).Hours() <= float64(c.config.MaxAge*24) {
				validFiles = append(validFiles, file) // 加入有效文件列表
			} else {
				// 删除过期文件
				if err := os.Remove(file); err != nil { // 删除文件
					fmt.Printf("remove expired log file error: %v\n", err) // 打印错误
				}
			}
		}

		// 然后基于 MaxBackups 清理未过期的文件
		if len(validFiles) > c.config.MaxBackups {
			for _, file := range validFiles[c.config.MaxBackups:] {
				if err := os.Remove(file); err != nil { // 删除文件
					fmt.Printf("remove old log file error: %v\n", err) // 打印错误
				}
			}
		}
	}
}
