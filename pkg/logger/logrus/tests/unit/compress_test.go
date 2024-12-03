package unit

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gobase/pkg/logger/logrus"
)

// TestNewLogCompressor 测试创建压缩器
func TestNewLogCompressor(t *testing.T) {
	config := logrus.CompressConfig{
		Enable:       true,
		Algorithm:    "gzip",
		Level:        gzip.BestCompression,
		DeleteSource: true,
		Interval:     time.Second,
	}

	compressor := logrus.NewLogCompressor(config)
	if compressor == nil {
		t.Error("NewLogCompressor should not return nil")
	}
}

// TestCompressService 测试压缩服务功能
func TestCompressService(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tmpDir, "test.log")
	content := "test log content"

	// 使用函数作用域确保文件被关闭
	func() {
		f, err := os.Create(testFile)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		defer f.Close()

		if _, err := f.WriteString(content); err != nil {
			t.Fatalf("Failed to write to test file: %v", err)
		}
	}()

	// 创建压缩器
	config := logrus.CompressConfig{
		Enable:       true,
		Algorithm:    "gzip",
		Level:        gzip.BestCompression,
		DeleteSource: true,
		Interval:     100 * time.Millisecond,
		LogPaths:     []string{testFile},
	}
	compressor := logrus.NewLogCompressor(config)

	// 启动压缩服务
	compressor.Start()

	// 等待足够时间让压缩发生
	time.Sleep(200 * time.Millisecond)

	// 停止服务
	compressor.Stop()

	// 验证压缩结果
	compressedFile := testFile + ".gz"

	// 验证压缩文件是否存在
	if _, err := os.Stat(compressedFile); os.IsNotExist(err) {
		t.Error("Compressed file should exist")
		return
	}

	// 验证源文件是否被删除
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("Source file should be deleted")
	}

	// 验证压缩文件内容
	func() {
		f, err := os.Open(compressedFile)
		if err != nil {
			t.Fatalf("Failed to open compressed file: %v", err)
		}
		defer f.Close()

		gr, err := gzip.NewReader(f)
		if err != nil {
			t.Fatalf("Failed to create gzip reader: %v", err)
		}
		defer gr.Close()

		decompressed, err := io.ReadAll(gr)
		if err != nil {
			t.Fatalf("Failed to read compressed content: %v", err)
		}

		if string(decompressed) != content {
			t.Error("Decompressed content does not match original")
		}
	}()
}

// TestDisabledCompressor 测试禁用压缩器
func TestDisabledCompressor(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tmpDir, "disabled_test.log")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 创建禁用的压缩器
	config := logrus.CompressConfig{
		Enable:       false,
		Algorithm:    "gzip",
		Level:        gzip.BestCompression,
		DeleteSource: true,
		Interval:     100 * time.Millisecond,
	}
	compressor := logrus.NewLogCompressor(config)

	// 启动服务
	compressor.Start()

	// 等待一段时间
	time.Sleep(200 * time.Millisecond)

	// 停止服务
	compressor.Stop()

	// 验证文件没有被压缩
	compressedFile := testFile + ".gz"
	if _, err := os.Stat(compressedFile); !os.IsNotExist(err) {
		t.Error("File should not be compressed when compressor is disabled")
	}

	// 验证源文件仍然存在
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Source file should still exist when compressor is disabled")
	}
}
