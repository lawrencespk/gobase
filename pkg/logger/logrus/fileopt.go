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
		opts.FlushInterval = time.Second
	}
	if opts.MaxOpenFiles <= 0 {
		opts.MaxOpenFiles = 100
	}

	fm := &FileManager{
		opts:   opts,                          // 配置
		files:  make(map[string]*FileHandle),  // 文件句柄映射
		pool:   NewBufferPool(),               // 缓冲池
		writes: NewWritePool(opts.BufferSize), // 写入器池
	}

	go fm.flushLoop()
	go fm.cleanupLoop()

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

	handle.mu.Lock()
	defer handle.mu.Unlock()

	handle.lastUse = time.Now()
	return handle.buffer.Write(p)
}

// getHandle 获取文件句柄
func (fm *FileManager) getHandle(filename string) (*FileHandle, error) {
	fm.mu.RLock()
	handle, ok := fm.files[filename]
	fm.mu.RUnlock()

	if ok {
		return handle, nil
	}

	fm.mu.Lock()
	defer fm.mu.Unlock()

	// 双重检查
	if handle, ok = fm.files[filename]; ok {
		return handle, nil
	}

	// 检查是否超过最大打开文件数
	if len(fm.files) >= fm.opts.MaxOpenFiles {
		fm.closeIdleFiles()
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
	fm.files[filename] = handle

	return handle, nil
}

// flushLoop 定期刷新缓冲区
func (fm *FileManager) flushLoop() {
	ticker := time.NewTicker(fm.opts.FlushInterval)
	defer ticker.Stop()

	for range ticker.C {
		fm.flushAll()
	}
}

// flushAll 刷新所有文件
func (fm *FileManager) flushAll() {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	for _, handle := range fm.files {
		handle.mu.Lock()
		if handle.buffer.Len() > 0 {
			handle.buffer.WriteTo(handle.file)
		}
		handle.mu.Unlock()
	}
}

// closeIdleFiles 关闭空闲文件
func (fm *FileManager) closeIdleFiles() {
	now := time.Now()
	for filename, handle := range fm.files {
		if now.Sub(handle.lastUse) > time.Minute {
			handle.mu.Lock()
			handle.buffer.WriteTo(handle.file)
			handle.file.Close()
			fm.pool.Put(handle.buffer)
			handle.mu.Unlock()
			delete(fm.files, filename)
		}
	}
}

// cleanupLoop 定期清理空闲文件
func (fm *FileManager) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		fm.mu.Lock()
		fm.closeIdleFiles()
		fm.mu.Unlock()
	}
}

// Close 关闭文件管理器
func (fm *FileManager) Close() error {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	for _, handle := range fm.files {
		handle.mu.Lock()
		handle.buffer.WriteTo(handle.file)
		handle.file.Close()
		fm.pool.Put(handle.buffer)
		handle.mu.Unlock()
	}

	fm.files = nil
	return nil
}
