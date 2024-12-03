package logrus

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
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
	config   CompressConfig // 压缩配置
	stopChan chan struct{}  // 停止信号通道
	wg       sync.WaitGroup // 等待组
	started  bool           // 是否启动
	mu       sync.Mutex     // 互斥锁
}

// NewLogCompressor 创建新的日志压缩器
func NewLogCompressor(config CompressConfig) *LogCompressor {
	return &LogCompressor{
		config:   config,              // 压缩配置
		stopChan: make(chan struct{}), // 停止信号通道
	}
}

// Start 启动压缩器
func (c *LogCompressor) Start() {
	c.mu.Lock()         // 锁定
	defer c.mu.Unlock() // 解锁

	if c.started { // 如果已经启动
		return
	}

	log.Printf("Starting compressor with config: %+v", c.config) // 打印压缩配置

	c.started = true // 设置为已启动
	c.wg.Add(1)      // 添加等待组
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(c.config.Interval) // 创建定时器
		defer ticker.Stop()                         // 停止定时器

		// 立即执行一次压缩
		if err := c.compressLogs(); err != nil {
			log.Printf("Error in initial compression: %v", err) // 打印初始压缩错误
		}

		for {
			select {
			case <-c.stopChan:
				log.Printf("Compressor received stop signal") // 打印停止信号
				return
			case <-ticker.C:
				log.Printf("Compressor ticker triggered") // 打印定时器触发
				if err := c.compressLogs(); err != nil {
					log.Printf("Error compressing logs: %v", err) // 打印压缩日志错误
				}
			}
		}
	}()

	log.Printf("Compressor started successfully") // 打印压缩器启动成功
}

// compressLogs 压缩日志文件
func (c *LogCompressor) compressLogs() error {
	log.Printf("Starting compression cycle with paths: %v", c.config.LogPaths) // 打印压缩路径

	for _, path := range c.config.LogPaths {
		// 跳过标准输出和标准错误
		if path == "stdout" || path == "stderr" {
			log.Printf("Skipping special path: %s", path) // 打印跳过特殊路径
			continue
		}

		compressedPath := path + ".gz" // 压缩路径

		// 检查是否已经存在压缩文件
		if _, err := os.Stat(compressedPath); err == nil {
			log.Printf("Compressed file already exists: %s", compressedPath) // 打印压缩文件已存在
			// 如果源文件仍然存在，尝试再次删除
			if _, err := os.Stat(path); err == nil && c.config.DeleteSource {
				if err := os.Remove(path); err != nil {
					log.Printf("Warning: Could not delete existing source file %s: %v", path, err) // 打印删除源文件失败
				} else {
					log.Printf("Successfully deleted existing source file: %s", path) // 打印删除源文件成功
				}
			}
			continue
		}

		// 读取源文件内容
		content, err := func() ([]byte, error) {
			srcFile, err := os.OpenFile(path, os.O_RDONLY, 0644) // 打开源文件
			if err != nil {
				return nil, fmt.Errorf("failed to open source file: %w", err) // 打印打开源文件失败
			}
			defer srcFile.Close() // 关闭源文件

			data, err := io.ReadAll(srcFile) // 读取源文件内容
			if err != nil {
				return nil, fmt.Errorf("failed to read source file: %w", err) // 打印读取源文件失败
			}

			log.Printf("Read %d bytes from source file: %s", len(data), path) // 打印读取源文件内容
			return data, nil
		}()

		if err != nil {
			if os.IsNotExist(err) {
				log.Printf("File does not exist: %s", path) // 打印文件不存在
				continue
			}
			log.Printf("Error reading source file: %v", err) // 打印读取源文件失败
			continue
		}

		// 如果文件内容为空，跳过
		if len(content) == 0 {
			log.Printf("Skipping empty file: %s", path) // 打印跳过空文件
			continue
		}

		// 创建压缩文件
		if err := func() error {
			gzFile, err := os.Create(compressedPath)
			if err != nil {
				return fmt.Errorf("failed to create compressed file: %w", err) // 打印创建压缩文件失败
			}
			defer gzFile.Close() // 关闭压缩文件

			gzWriter := gzip.NewWriter(gzFile) // 创建压缩写入器
			defer gzWriter.Close()             // 关闭压缩写入器

			written, err := gzWriter.Write(content) // 写入压缩数据
			if err != nil {
				return fmt.Errorf("failed to write compressed data: %w", err) // 打印写入压缩数据失败
			}
			log.Printf("Wrote %d bytes to compressed file", written) // 打印写入压缩数据

			if err := gzWriter.Close(); err != nil {
				return fmt.Errorf("failed to close gzip writer: %w", err) // 打印关闭压缩写入器失败
			}

			return nil
		}(); err != nil {
			log.Printf("Error compressing file: %v", err) // 打印压缩文件失败
			os.Remove(compressedPath)                     // 删除压缩文件
			return err                                    // 返回错误
		}

		log.Printf("Successfully compressed %s to %s", path, compressedPath) // 打印压缩文件成功

		// 如果需要删除源文件，尝试删除
		if c.config.DeleteSource {
			// 多次尝试删除源文件
			maxRetries := 3
			for i := 0; i < maxRetries; i++ {
				if err := os.Remove(path); err != nil {
					if i < maxRetries-1 {
						log.Printf("Retry %d: Could not delete source file %s: %v", i+1, path, err) // 打印删除源文件失败
						time.Sleep(time.Second)                                                     // 等待一秒后重试
						continue
					}
					log.Printf("Warning: Failed to delete source file after %d retries: %s", maxRetries, path) // 打印删除源文件失败
				} else {
					log.Printf("Successfully deleted source file: %s", path) // 打印删除源文件成功
					break
				}
			}
		}
	}

	return nil
}

// Stop 停止压缩器
func (c *LogCompressor) Stop() {
	c.mu.Lock()         // 锁定
	defer c.mu.Unlock() // 解锁

	if !c.started { // 如果未启动
		return
	}

	close(c.stopChan) // 关闭停止信号通道
	c.wg.Wait()       // 等待
	c.started = false // 设置为未启动
}
