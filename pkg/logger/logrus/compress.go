package logrus

import (
	"compress/gzip"
	"io"
	"os"
	"sync"
	"time"

	"gobase/pkg/errors"

	"github.com/sirupsen/logrus"
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

	logrus.WithField("config", c.config).Info("Starting compressor") // 打印压缩配置

	c.started = true // 设置为已启动
	c.wg.Add(1)      // 添加等待组
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(c.config.Interval) // 创建定时器
		defer ticker.Stop()                         // 停止定时器

		// 立即执行一次压缩
		if err := c.compressLogs(); err != nil {
			logrus.WithError(err).Error("Error in initial compression") // 打印初始压缩错误
		}

		for {
			select {
			case <-c.stopChan:
				logrus.Info("Compressor received stop signal") // 打印停止信号
				return
			case <-ticker.C:
				logrus.Debug("Compressor ticker triggered") // 打印定时器触发
				if err := c.compressLogs(); err != nil {
					logrus.WithError(err).Error("Error compressing logs") // 打印压缩日志错误
				}
			}
		}
	}()

	logrus.Info("Compressor started successfully") // 打印压缩器启动成功
}

// compressLogs 压缩日志文件
func (c *LogCompressor) compressLogs() error {
	for _, path := range c.config.LogPaths {
		// 跳过标准输出和标准错误
		if path == "stdout" || path == "stderr" {
			continue
		}

		// 确保源文件已经完全写入
		if err := c.ensureFileReady(path); err != nil {
			logrus.WithError(err).WithField("path", path).Error("File not ready for compression")
			continue
		}

		compressedPath := path + ".gz"

		// 检查是否已经存在压缩文件
		if _, err := os.Stat(compressedPath); err == nil {
			continue
		}

		// 以只读方式打开源文件
		srcFile, err := os.OpenFile(path, os.O_RDONLY, 0644)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			logrus.WithError(err).WithField("path", path).Error("Error opening source file")
			continue
		}

		// 确保文件句柄被关闭
		if err := c.compressFile(srcFile, compressedPath); err != nil {
			logrus.WithError(err).WithField("path", path).Error("Compression failed")
			continue
		}

		// 如果需要删除源文件
		if c.config.DeleteSource {
			if err := c.deleteSourceWithRetry(path); err != nil {
				logrus.WithError(err).WithField("path", path).Error("Failed to delete source file")
			}
		}
	}
	return nil
}

// ensureFileReady 确保文件准备好被压缩
func (c *LogCompressor) ensureFileReady(path string) error {
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return errors.NewFileOperationError("failed to open file for compression", err)
	}
	defer file.Close()

	// 尝试读取文件状态
	_, err = file.Stat()
	if err != nil {
		return errors.NewFileOperationError("failed to get file stats", err)
	}

	// 简单检查文件是否可读
	buf := make([]byte, 1)
	_, err = file.Read(buf)
	if err != nil && err != io.EOF {
		return errors.NewFileOperationError("file is not readable", err)
	}

	return nil
}

// compressFile 压缩文件
func (c *LogCompressor) compressFile(srcFile *os.File, compressedPath string) error {
	defer srcFile.Close() // 确保源文件被关闭

	// 创建压缩文件
	gzFile, err := os.Create(compressedPath)
	if err != nil {
		return errors.NewFileOperationError("failed to create compressed file", err)
	}
	defer gzFile.Close()

	// 创建 gzip writer
	gzWriter := gzip.NewWriter(gzFile)
	defer gzWriter.Close()

	// 直接从源文件复制到压缩文件
	if _, err := io.Copy(gzWriter, srcFile); err != nil {
		os.Remove(compressedPath)
		return errors.NewFileOperationError("failed to compress file", err)
	}

	// 确保所有数据都被写入
	if err := gzWriter.Close(); err != nil {
		os.Remove(compressedPath)
		return errors.NewFileOperationError("failed to finalize compression", err)
	}

	return nil
}

// deleteSourceWithRetry 删除源文件并重试
func (c *LogCompressor) deleteSourceWithRetry(path string) error {
	// 多次尝试删除源文件
	maxRetries := 3
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := os.Remove(path); err != nil {
			lastErr = err
			if i < maxRetries-1 {
				logrus.WithError(err).WithFields(logrus.Fields{
					"path":  path,
					"retry": i + 1,
				}).Warn("Could not delete source file")
				time.Sleep(time.Second)
				continue
			}
			logrus.WithError(err).WithFields(logrus.Fields{
				"path":    path,
				"retries": maxRetries,
			}).Warn("Failed to delete source file")
		} else {
			logrus.WithField("path", path).Info("Successfully deleted source file")
			return nil
		}
	}
	return errors.NewFileOperationError("failed to delete source file after retries", lastErr)
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
