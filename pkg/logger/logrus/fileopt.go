package logrus

import (
	"bytes"
	"os"
	"sync"
	"time"
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
	opts   FileOptions            // 配置
	files  map[string]*FileHandle // 文件句柄映射
	mu     sync.RWMutex           // 读写锁
	pool   *BufferPool            // 缓冲池
	writes *WritePool             // 写入器池
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
		return 0, err
	}

	handle.mu.Lock()         // 锁定
	defer handle.mu.Unlock() // 解锁

	// 写入缓冲区
	n, err = handle.buffer.Write(p)
	if err != nil {
		return n, err
	}

	// 如果缓冲区超过阈值，立即刷新
	if handle.buffer.Len() >= fm.opts.BufferSize {
		if err := fm.flushHandle(handle); err != nil {
			return n, err
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
			return err
		}
		// 确保写入磁盘
		if err := handle.file.Sync(); err != nil {
			return err
		}
	}
	return nil
}

// getHandle 获取文件句柄
func (fm *FileManager) getHandle(filename string) (*FileHandle, error) {
	fm.mu.RLock()                    // 锁定
	handle, ok := fm.files[filename] // 获取文件句柄
	fm.mu.RUnlock()                  // 解锁

	if ok {
		return handle, nil // 如果文件句柄存在，返回文件句柄
	}

	fm.mu.Lock()
	defer fm.mu.Unlock()

	// 双重检查
	if handle, ok = fm.files[filename]; ok {
		return handle, nil
	}

	// 检查是否超过最大打开文件数
	if len(fm.files) >= fm.opts.MaxOpenFiles {
		fm.closeIdleFiles() // 关闭空闲文件
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	handle = &FileHandle{
		file:    file,          // 文件
		buffer:  fm.pool.Get(), // 缓冲区
		lastUse: time.Now(),    // 最后一次使用时间
	}
	fm.files[filename] = handle // 添加文件句柄

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
			handle.mu.Lock()                   // 锁定
			handle.buffer.WriteTo(handle.file) // 写入文件
			handle.file.Close()                // 关闭文件
			fm.pool.Put(handle.buffer)         // 释放缓冲区
			handle.mu.Unlock()                 // 解锁
			delete(fm.files, filename)         // 删除文件句柄
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

	for _, handle := range fm.files {
		handle.mu.Lock()           // 锁定
		fm.flushHandle(handle)     // 刷新文件
		handle.file.Close()        // 关闭文件
		fm.pool.Put(handle.buffer) // 释放缓冲区
		handle.mu.Unlock()         // 解锁
	}

	fm.files = nil // 清空文件句柄映射
	return nil
}
