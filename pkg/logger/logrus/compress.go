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
}

// LogCompressor 日志压缩器
type LogCompressor struct {
	config CompressConfig
	done   chan struct{}
}

// NewLogCompressor 创建日志压缩器
func NewLogCompressor(config CompressConfig) *LogCompressor {
	return &LogCompressor{
		config: config,
		done:   make(chan struct{}),
	}
}

// Start 启动压缩服务
func (c *LogCompressor) Start() {
	if !c.config.Enable {
		return
	}

	go c.run()
}

// Stop 停止压缩服务
func (c *LogCompressor) Stop() {
	close(c.done)
}

// run 运行压缩服务
func (c *LogCompressor) run() {
	ticker := time.NewTicker(c.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.compressOldLogs()
		case <-c.done:
			return
		}
	}
}

// compressOldLogs 压缩旧日志文件
func (c *LogCompressor) compressOldLogs() {
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

		for _, file := range files {
			if strings.HasSuffix(file, ".gz") {
				continue
			}

			if err := c.compressFile(file); err != nil {
				fmt.Printf("compress file error: %v\n", err)
			}
		}
	}
}

// compressFile 压缩单个文件
func (c *LogCompressor) compressFile(filename string) error {
	// 打开源文件
	source, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("open source file error: %v", err)
	}
	defer source.Close()

	// 创建目标文件
	target, err := os.Create(filename + ".gz")
	if err != nil {
		return fmt.Errorf("create target file error: %v", err)
	}
	defer target.Close()

	// 创建gzip写入器
	gw, err := gzip.NewWriterLevel(target, c.config.Level)
	if err != nil {
		return fmt.Errorf("create gzip writer error: %v", err)
	}
	defer gw.Close()

	// 写入文件头信息
	gw.Header.Name = filepath.Base(filename)
	gw.Header.ModTime = time.Now()

	// 复制文件内容
	if _, err := io.Copy(gw, source); err != nil {
		return fmt.Errorf("copy file content error: %v", err)
	}

	// 如果配置了删除源文件
	if c.config.DeleteSource {
		if err := os.Remove(filename); err != nil {
			return fmt.Errorf("remove source file error: %v", err)
		}
	}

	return nil
}
