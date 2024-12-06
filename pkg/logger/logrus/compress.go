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
	logrus.WithField("paths", c.config.LogPaths).Info("Starting compression cycle") // 打印压缩路径

	for _, path := range c.config.LogPaths {
		// 跳过标准输出和标准错误
		if path == "stdout" || path == "stderr" {
			logrus.WithField("path", path).Debug("Skipping special path") // 打印跳过特殊路径
			continue
		}

		compressedPath := path + ".gz" // 压缩路径

		// 检查是否已经存在压缩文件
		if _, err := os.Stat(compressedPath); err == nil {
			logrus.WithField("path", compressedPath).Debug("Compressed file already exists") // 打印压缩文件已存在
			// 如果源文件仍然存在，尝试再次删除
			if _, err := os.Stat(path); err == nil && c.config.DeleteSource {
				if err := os.Remove(path); err != nil {
					logrus.WithError(err).WithField("path", path).Warn("Could not delete existing source file") // 打印删除源文件失败
				} else {
					logrus.WithField("path", path).Info("Successfully deleted existing source file") // 打印删除源文件成功
				}
			}
			continue
		}

		// 读取源文件内容
		content, err := func() ([]byte, error) {
			srcFile, err := os.OpenFile(path, os.O_RDONLY, 0644) // 打开源文件
			if err != nil {
				return nil, errors.NewFileNotFoundError("failed to open source file", err) // 打开源文件失败
			}
			defer srcFile.Close() // 关闭源文件

			data, err := io.ReadAll(srcFile) // 读取源文件内容
			if err != nil {
				return nil, errors.NewFileDownloadError("failed to read source file", err) // 读取源文件失败
			}

			logrus.WithFields(logrus.Fields{
				"bytes": len(data),
				"path":  path,
			}).Debug("Read from source file") // 打印读取源文件内容
			return data, nil
		}()

		if err != nil {
			if os.IsNotExist(err) {
				logrus.WithField("path", path).Debug("File does not exist") // 打印文件不存在
				continue
			}
			logrus.WithError(err).WithField("path", path).Error("Error reading source file") // 打印读取源文件失败
			continue
		}

		// 如果文件内容为空，跳过
		if len(content) == 0 {
			logrus.WithField("path", path).Debug("Skipping empty file") // 打印跳过空文件
			continue
		}

		// 创建压缩文件
		if err := func() error {
			gzFile, err := os.Create(compressedPath)
			if err != nil {
				return errors.NewFileUploadError("failed to create compressed file", err)
			}
			defer gzFile.Close() // 关闭压缩文件

			gzWriter := gzip.NewWriter(gzFile) // 创建压缩写入器
			defer gzWriter.Close()             // 关闭压缩写入器

			written, err := gzWriter.Write(content) // 写入压缩数据
			if err != nil {
				return errors.NewFileUploadError("failed to write compressed data", err)
			}
			logrus.WithField("bytes", written).Debug("Wrote to compressed file") // 打印写入压缩数据

			if err := gzWriter.Close(); err != nil {
				return errors.NewFileUploadError("failed to close gzip writer", err)
			}

			return nil
		}(); err != nil {
			logrus.WithError(err).Error("Error compressing file") // 打印压缩文件失败
			os.Remove(compressedPath)                             // 删除压缩文件
			return err                                            // 返回错误
		}

		logrus.WithFields(logrus.Fields{
			"source": path,
			"target": compressedPath,
		}).Info("Successfully compressed file") // 打印压缩文件成功

		// 如果需要删除源文件，尝试删除
		if c.config.DeleteSource {
			// 多次尝试删除源文件
			maxRetries := 3
			for i := 0; i < maxRetries; i++ { // 多次尝试删除源文件
				if err := os.Remove(path); err != nil { // 删除源文件
					if i < maxRetries-1 { // 如果未达到最大重试次数
						logrus.WithError(err).WithFields(logrus.Fields{ // 打印删除源文件失败
							"path":  path,  // 路径
							"retry": i + 1, // 重试次数
						}).Warn("Could not delete source file") // 打印删除源文件失败
						time.Sleep(time.Second) // 等待一秒后重试
						continue
					}
					logrus.WithError(err).WithFields(logrus.Fields{ // 打印删除源文件失败
						"path":    path,       // 路径
						"retries": maxRetries, // 重试次数
					}).Warn("Failed to delete source file") // 打印删除源文件失败
				} else {
					logrus.WithField("path", path).Info("Successfully deleted source file") // 打印删除源文件成功
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
