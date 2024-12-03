package logrus

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CompressConfig 压缩配置
type CompressConfig struct {
	Enable       bool          // 是否启用压缩
	Algorithm    string        // 压缩算法: gzip, zstd 等
	Level        int           // 压缩级别
	DeleteSource bool          // 压缩后是否删除源文件
	Interval     time.Duration // 压缩检查间隔
	LogPaths     []string      // 需要监控的日志路径
}

// LogCompressor 日志压缩器
type LogCompressor struct {
	config CompressConfig // 压缩配置
	done   chan struct{}  // 结束通道
}

// NewLogCompressor 创建日志压缩器
func NewLogCompressor(config CompressConfig) *LogCompressor {
	return &LogCompressor{
		config: config,              // 压缩配置
		done:   make(chan struct{}), // 结束通道
	}
}

// Start 启动压缩服务
func (c *LogCompressor) Start() {
	if !c.config.Enable { // 如果未启用压缩
		return
	}

	go c.run() // 运行压缩服务
}

// Stop 停止压缩服务
func (c *LogCompressor) Stop() {
	close(c.done) // 关闭结束通道
}

// run 运行压缩服务
func (c *LogCompressor) run() {
	ticker := time.NewTicker(c.config.Interval) // 创建定时器
	defer ticker.Stop()                         // 停止定时器

	for {
		select {
		case <-ticker.C: // 定时器触发
			c.compressOldLogs() // 压缩旧日志文件
		case <-c.done: // 结束通道触发
			return // 退出
		}
	}
}

// compressOldLogs 压缩旧日志文件
func (c *LogCompressor) compressOldLogs() {
	for _, path := range c.config.LogPaths { // 使用配置中的路径
		if path == "stdout" || path == "stderr" { // 跳过标准输出和标准错误
			continue
		}

		dir := filepath.Dir(path)              // 获取目录
		pattern := filepath.Join(dir, "*.log") // 获取日志文件模式

		files, err := filepath.Glob(pattern) // 获取所有匹配的文件
		if err != nil {
			continue // 如果获取失败，跳过
		}

		for _, file := range files {
			if strings.HasSuffix(file, ".gz") { // 如果文件已压缩，跳过
				continue
			}

			if err := c.compressFile(file); err != nil { // 压缩文件
				fmt.Printf("compress file error: %v\n", err) // 打印错误
			}
		}
	}
}

// compressFile 压缩单个文件
func (c *LogCompressor) compressFile(filename string) error {
	// 打开源文件
	source, err := os.Open(filename) // 打开源文件
	if err != nil {
		return fmt.Errorf("open source file error: %v", err) // 如果打开失败，返回错误
	}

	// 读取所有内容并立即关闭源文件
	content, err := io.ReadAll(source) // 读取所有内容
	source.Close()                     // 立即关闭源文件
	if err != nil {
		return fmt.Errorf("read source file error: %v", err) // 如果读取失败，返回错误
	}

	// 创建目标文件
	target, err := os.Create(filename + ".gz") // 创建目标文件
	if err != nil {
		return fmt.Errorf("create target file error: %v", err) // 如果创建失败，返回错误
	}
	defer target.Close()

	// 创建gzip写入器
	gw, err := gzip.NewWriterLevel(target, c.config.Level) // 创建gzip写入器
	if err != nil {
		return fmt.Errorf("create gzip writer error: %v", err) // 如果创建失败，返回错误
	}
	defer gw.Close()

	// 写入文件头信息
	gw.Header.Name = filepath.Base(filename) // 设置文件名
	gw.Header.ModTime = time.Now()           // 设置修改时间

	// 写入内容
	if _, err := gw.Write(content); err != nil {
		return fmt.Errorf("write content error: %v", err) // 如果写入失败，返回错误
	}

	// 确保所有内容都已写入
	if err := gw.Close(); err != nil {
		return fmt.Errorf("close gzip writer error: %v", err) // 如果关闭失败，返回错误
	}

	if err := target.Close(); err != nil {
		return fmt.Errorf("close target file error: %v", err) // 如果关闭失败，返回错误
	}

	// 删除源文件
	if c.config.DeleteSource {
		if err := os.Remove(filename); err != nil {
			return fmt.Errorf("remove source file error: %v", err) // 如果删除失败，返回错误
		}
	}

	return nil
}
