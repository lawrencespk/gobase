package unit

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"gobase/pkg/logger/logrus"
)

func TestFileManager(t *testing.T) {
	// 创建临时目录用于测试
	tempDir, err := os.MkdirTemp("", "filemanager_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("Basic Write Operations", func(t *testing.T) {
		logFile := filepath.Join(tempDir, "test.log")
		opts := logrus.FileOptions{
			BufferSize:    1024,
			FlushInterval: time.Millisecond * 100,
			MaxOpenFiles:  10,
			DefaultPath:   logFile,
		}

		fm := logrus.NewFileManager(opts)
		defer fm.Close()

		testData := []byte("test log entry\n")
		n, err := fm.Write(testData)
		if err != nil {
			t.Errorf("Write failed: %v", err)
		}
		if n != len(testData) {
			t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
		}

		// 等待刷新到磁盘
		time.Sleep(opts.FlushInterval * 2)

		// 验证文件内容
		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Errorf("Failed to read log file: %v", err)
		}
		if string(content) != string(testData) {
			t.Errorf("Expected file content %q, got %q", testData, content)
		}
	})

	t.Run("Multiple Files", func(t *testing.T) {
		opts := logrus.FileOptions{
			BufferSize:    1024,
			FlushInterval: time.Millisecond * 100,
			MaxOpenFiles:  10,
			DefaultPath:   filepath.Join(tempDir, "default.log"),
		}

		fm := logrus.NewFileManager(opts)
		defer fm.Close()

		// 写入多个文件
		files := 5
		writes := 10
		var wg sync.WaitGroup

		for i := 0; i < files; i++ {
			wg.Add(1)
			go func(fileNum int) {
				defer wg.Done()
				filename := filepath.Join(tempDir, fmt.Sprintf("test%d.log", fileNum))
				for j := 0; j < writes; j++ {
					data := []byte(fmt.Sprintf("file%d-entry%d\n", fileNum, j))
					_, err := fm.WriteToFile(filename, data)
					if err != nil {
						t.Errorf("Write to file %d failed: %v", fileNum, err)
					}
				}
			}(i)
		}

		wg.Wait()
		time.Sleep(opts.FlushInterval * 2)

		// 验证所有文件内容
		for i := 0; i < files; i++ {
			filename := filepath.Join(tempDir, fmt.Sprintf("test%d.log", i))
			content, err := os.ReadFile(filename)
			if err != nil {
				t.Errorf("Failed to read file %d: %v", i, err)
				continue
			}

			expectedLines := writes
			actualLines := bytes.Count(content, []byte("\n"))
			if actualLines != expectedLines {
				t.Errorf("File %d: expected %d lines, got %d", i, expectedLines, actualLines)
			}
		}
	})

	t.Run("Max Open Files", func(t *testing.T) {
		opts := logrus.FileOptions{
			BufferSize:    1024,
			FlushInterval: time.Millisecond * 100,
			MaxOpenFiles:  2, // 设置较小的最大打开文件数
			DefaultPath:   filepath.Join(tempDir, "default.log"),
		}

		fm := logrus.NewFileManager(opts)
		defer fm.Close()

		// 尝试写入超过最大打开文件数的文件
		for i := 0; i < 5; i++ {
			filename := filepath.Join(tempDir, fmt.Sprintf("maxtest%d.log", i))
			data := []byte(fmt.Sprintf("test data for file %d\n", i))
			_, err := fm.WriteToFile(filename, data)
			if err != nil {
				t.Errorf("Write to file %d failed: %v", i, err)
			}
		}

		// 等待文件管理器进行清理
		time.Sleep(time.Second)

		// 验证实际打开的文件数不超过限制
		openFiles := 0
		filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				openFiles++
			}
			return nil
		})

		if openFiles < opts.MaxOpenFiles {
			t.Errorf("Expected at least %d open files, got %d", opts.MaxOpenFiles, openFiles)
		}
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		opts := logrus.FileOptions{
			BufferSize:    1024,
			FlushInterval: time.Millisecond * 100,
			MaxOpenFiles:  10,
			DefaultPath:   filepath.Join(tempDir, "concurrent.log"),
		}

		fm := logrus.NewFileManager(opts)
		defer fm.Close()

		var wg sync.WaitGroup
		workers := 10
		iterations := 100

		wg.Add(workers)
		for i := 0; i < workers; i++ {
			go func(workerID int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					data := []byte(fmt.Sprintf("worker%d-entry%d\n", workerID, j))
					_, err := fm.Write(data)
					if err != nil {
						t.Errorf("Worker %d write failed: %v", workerID, err)
					}
				}
			}(i)
		}

		wg.Wait()
		time.Sleep(opts.FlushInterval * 2)

		// 验证所有数据都被写入
		content, err := os.ReadFile(opts.DefaultPath)
		if err != nil {
			t.Errorf("Failed to read concurrent log file: %v", err)
		}

		expectedLines := workers * iterations
		actualLines := bytes.Count(content, []byte("\n"))
		if actualLines != expectedLines {
			t.Errorf("Expected %d lines, got %d", expectedLines, actualLines)
		}
	})
}

func BenchmarkFileManager(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "filemanager_bench_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	opts := logrus.FileOptions{
		BufferSize:    32 * 1024,
		FlushInterval: time.Millisecond * 100,
		MaxOpenFiles:  100,
		DefaultPath:   filepath.Join(tempDir, "bench.log"),
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
