package benchmark

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gobase/pkg/logger/logrus"
)

// BenchmarkAsyncWriter 测试异步写入器的性能
func BenchmarkAsyncWriter(b *testing.B) {
	mock := NewMockWriter()
	config := logrus.AsyncConfig{
		Enable:        true,
		BufferSize:    1024 * 1024, // 1MB 缓冲区
		FlushInterval: time.Millisecond * 100,
		BlockOnFull:   true,
		FlushOnExit:   true,
	}

	writer := logrus.NewAsyncWriter(mock, config)
	defer writer.Stop()

	data := make([]byte, 1024) // 1KB 数据
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			writer.Write(data)
		}
	})
}

// BenchmarkLogCompressor 测试日志压缩器的性能
func BenchmarkLogCompressor(b *testing.B) {
	// 创建临时测试目录
	testDir, err := os.MkdirTemp("", "compressor_benchmark_*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	// 创建测试文件
	testFile := filepath.Join(testDir, "test.log")
	if err := os.WriteFile(testFile, []byte("test log content"), 0644); err != nil {
		b.Fatal(err)
	}

	config := logrus.CompressConfig{
		Enable:       true,
		Algorithm:    "gzip",
		Level:        5,
		DeleteSource: false,
		Interval:     time.Second,
		LogPaths:     []string{testFile},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			compressor := logrus.NewLogCompressor(config)
			compressor.Start()
			time.Sleep(100 * time.Millisecond)
			compressor.Stop()
		}
	})
}

// BenchmarkLogCleaner 测试日志清理器的性能
func BenchmarkLogCleaner(b *testing.B) {
	// 创建临时测试目录
	testDir, err := os.MkdirTemp("", "cleaner_benchmark_*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	// 创建测试文件
	testFile := filepath.Join(testDir, "test.log")
	if err := os.WriteFile(testFile, []byte("test log content"), 0644); err != nil {
		b.Fatal(err)
	}

	// 设置文件修改时间为过去
	pastTime := time.Now().Add(-24 * time.Hour)
	if err := os.Chtimes(testFile, pastTime, pastTime); err != nil {
		b.Fatal(err)
	}

	config := logrus.CleanupConfig{
		Enable:   true,
		MaxAge:   86400, // 24小时，以秒为单位
		Interval: time.Second,
		LogPaths: []string{testDir},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cleaner := logrus.NewLogCleaner(config)
			cleaner.Start()
			time.Sleep(100 * time.Millisecond)
			cleaner.Stop()
		}
	})
}

// BenchmarkFileManager 测试文件管理器的性能
func BenchmarkFileManager(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "filemanager_benchmark_*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	opts := logrus.FileOptions{
		BufferSize:    32 * 1024,
		FlushInterval: time.Millisecond * 100,
		MaxOpenFiles:  100,
		DefaultPath:   filepath.Join(tmpDir, "bench.log"),
	}

	fm := logrus.NewFileManager(opts)
	defer fm.Close()

	data := []byte("benchmark test data\n")

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			fm.Write(data)
		}
	})
}
