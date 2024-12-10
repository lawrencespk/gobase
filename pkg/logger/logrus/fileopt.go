package logrus

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"gobase/pkg/errors"
	"gobase/pkg/errors/types"

	"github.com/sirupsen/logrus"
)

// FileOptions 文件操作配置
type FileOptions struct {
	BufferSize    int           // 写入缓冲区大小
	FlushInterval time.Duration // 刷新间隔
	MaxOpenFiles  int           // 最大打开文件数
	DefaultPath   string        // 默认日志文件路径
}

// FileManager 文件管理器
type FileManager struct {
	mu          sync.RWMutex           // 读写锁
	opts        FileOptions            // 配置
	currentFile *os.File               // 当前文件
	files       map[string]*FileHandle // 文件句柄映射
	pool        *BufferPool            // 缓冲池
	writes      *WritePool             // 写入器池
}

// FileHandle 文件句柄
type FileHandle struct {
	file    *os.File      // 文件
	buffer  *bytes.Buffer // 缓冲区
	lastUse time.Time     // 最后一次使用时间
	mu      sync.Mutex    // 互斥锁
}

// NewFileManager 创建文件管理器
func NewFileManager(opts FileOptions) *FileManager {
	if opts.BufferSize <= 0 {
		opts.BufferSize = 32 * 1024 // 32KB default
	}
	if opts.FlushInterval <= 0 {
		opts.FlushInterval = time.Second // 默认刷新间隔为1秒
	}
	if opts.MaxOpenFiles <= 0 {
		opts.MaxOpenFiles = 100 // 默认最大打开文件数为100
	}

	fm := &FileManager{
		opts:   opts,                          // 配置
		files:  make(map[string]*FileHandle),  // 文件句柄映射
		pool:   NewBufferPool(),               // 缓冲池
		writes: NewWritePool(opts.BufferSize), // 写入器池
	}

	go fm.flushLoop()   // 启动刷新循环
	go fm.cleanupLoop() // 启动清理循环

	return fm
}

// Write 写入文件
func (fm *FileManager) Write(p []byte) (n int, err error) {
	// 使用默认的日志文件路径
	filename := fm.opts.DefaultPath
	if filename == "" {
		filename = "app.log"
	}
	return fm.WriteToFile(filename, p)
}

// WriteToFile 写入指定文件
func (fm *FileManager) WriteToFile(filename string, p []byte) (n int, err error) {
	handle, err := fm.getHandle(filename)
	if err != nil {
		return 0, errors.NewFileOperationError("获取文件句柄失败", err)
	}

	handle.mu.Lock()         // 锁定
	defer handle.mu.Unlock() // 解锁

	// 写入缓冲区
	n, err = handle.buffer.Write(p)
	if err != nil {
		return n, errors.NewFileWriteError("写入缓冲区失败", err)
	}

	// 如果缓冲区超过阈值，立即刷新
	if handle.buffer.Len() >= fm.opts.BufferSize {
		if err := fm.flushHandle(handle); err != nil {
			return n, errors.NewFileFlushError("刷新缓冲区失败", err)
		}
	}

	handle.lastUse = time.Now()
	return n, nil
}

// flushHandle 刷新指定句柄的缓冲区
func (fm *FileManager) flushHandle(handle *FileHandle) error {
	if handle.buffer.Len() > 0 {
		_, err := handle.buffer.WriteTo(handle.file)
		if err != nil {
			return errors.NewFileWriteError("写入文件失败", err)
		}
		// 确保写入磁盘
		if err := handle.file.Sync(); err != nil {
			return errors.NewFileFlushError("同步文件失败", err)
		}
	}
	return nil
}

// getHandle 获取文件句柄
func (fm *FileManager) getHandle(filename string) (*FileHandle, error) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// 检查是否已存在句柄
	if handle, exists := fm.files[filename]; exists {
		return handle, nil
	}

	// Windows 系统特殊处理
	fileMode := os.FileMode(0644)
	if runtime.GOOS == "windows" {
		fileMode = 0666
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filename), fileMode); err != nil {
		return nil, errors.NewFileOperationError("创建目录失败", err)
	}

	// 打开文件
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, fileMode)
	if err != nil {
		// Windows 权限处理
		if os.IsPermission(err) && runtime.GOOS == "windows" {
			cmd := exec.Command("icacls", filename, "/grant", "Everyone:F")
			if cmdErr := cmd.Run(); cmdErr == nil {
				file, err = os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, fileMode)
			}
		}
		if err != nil {
			return nil, errors.NewFileOpenError("打开文件失败", err)
		}
	}

	// 创建新的文件句柄
	handle := &FileHandle{
		file:    file,          // 文件
		buffer:  fm.pool.Get(), // 缓冲区
		lastUse: time.Now(),    // 最后一次使用时间
	}
	fm.files[filename] = handle

	return handle, nil
}

// flushLoop 定期刷新缓冲区
func (fm *FileManager) flushLoop() {
	ticker := time.NewTicker(fm.opts.FlushInterval)
	defer ticker.Stop() // 停止定时器

	for range ticker.C {
		fm.flushAll() // 刷新所有文件
	}
}

// flushAll 刷新所有文件
func (fm *FileManager) flushAll() {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	for _, handle := range fm.files {
		handle.mu.Lock()       // 锁定
		fm.flushHandle(handle) // 刷新文件
		handle.mu.Unlock()     // 解锁
	}
}

// closeIdleFiles 关闭空闲文件
func (fm *FileManager) closeIdleFiles() {
	now := time.Now()
	for filename, handle := range fm.files {
		if now.Sub(handle.lastUse) > time.Minute {
			handle.mu.Lock() // 锁定
			if _, err := handle.buffer.WriteTo(handle.file); err != nil {
				logrus.WithError(errors.NewFileWriteError("写入文件失败", err)).Error("关闭空闲文件时写入失败")
			}
			if err := handle.file.Close(); err != nil {
				logrus.WithError(errors.NewFileCloseError("关闭文件失败", err)).Error("关闭空闲文件失败")
			}
			fm.pool.Put(handle.buffer) // 释放缓冲区
			handle.mu.Unlock()         // 解锁
			delete(fm.files, filename) // 删除文件句柄
		}
	}
}

// cleanupLoop 定期清理空闲文件
func (fm *FileManager) cleanupLoop() {
	ticker := time.NewTicker(time.Minute) // 每分钟清理一次
	defer ticker.Stop()                   // 停止定时器

	for range ticker.C {
		fm.mu.Lock()        // 锁定
		fm.closeIdleFiles() // 关闭空闲文件
		fm.mu.Unlock()      // 解锁
	}
}

// Close 关闭文件管理器
func (fm *FileManager) Close() error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	var errs []error
	for _, handle := range fm.files {
		handle.mu.Lock() // 锁定
		if err := fm.flushHandle(handle); err != nil {
			errs = append(errs, errors.NewFileFlushError("关闭时刷新失败", err))
		}
		if err := handle.file.Close(); err != nil {
			errs = append(errs, errors.NewFileCloseError("关闭文件失败", err))
		}
		fm.pool.Put(handle.buffer) // 释放缓冲区
		handle.mu.Unlock()         // 解锁
	}

	fm.files = nil // 清空文件句柄映射

	if len(errs) > 0 {
		// 将 []error 转换为 []interface{}
		errDetails := make([]interface{}, len(errs))
		for i, err := range errs {
			errDetails[i] = err
		}
		baseErr := errors.NewFileOperationError("关闭文件管理器时发生错误", nil)
		if customErr, ok := baseErr.(types.Error); ok {
			return customErr.WithDetails(errDetails...)
		}
		return baseErr
	}
	return nil
}

// SetWriter 设置写入器
func (fm *FileManager) SetWriter(w io.Writer) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if file, ok := w.(*os.File); ok {
		fm.currentFile = file
		return nil
	}
	return errors.NewFileOperationError("writer must be *os.File", nil)
}
