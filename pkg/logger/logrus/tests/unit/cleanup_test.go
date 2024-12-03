package unit

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gobase/pkg/logger/logrus"
)

// TestNewLogCleaner 测试创建清理器
func TestNewLogCleaner(t *testing.T) {
	config := logrus.CleanupConfig{
		Enable:     true,        // 启用清理
		MaxBackups: 3,           // 保留3个备份
		MaxAge:     7,           // 7天后过期
		Interval:   time.Second, // 1秒检查一次
	}

	cleaner := logrus.NewLogCleaner(config)
	if cleaner == nil {
		t.Error("NewLogCleaner should not return nil") // 清理器不应该返回nil
	}
}

// TestCleanupService 测试清理服务功能
func TestCleanupService(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建多个测试日志文件，模拟不同时间的日志
	files := []struct {
		name    string    // 文件名
		content string    // 文件内容
		modTime time.Time // 文件修改时间
	}{
		{
			name:    "recent1.log",  // 最近1天的日志
			content: "recent log 1", // 日志内容
			modTime: time.Now(),     // 当前时间
		},
		{
			name:    "recent2.log",                       // 最近2天的日志
			content: "recent log 2",                      // 日志内容
			modTime: time.Now().Add(-1 * 24 * time.Hour), // 1天前
		},
		{
			name:    "recent3.log",                       // 最近3天的日志
			content: "recent log 3",                      // 日志内容
			modTime: time.Now().Add(-2 * 24 * time.Hour), // 2天前
		},
		{
			name:    "old1.log",                          // 旧的日志
			content: "old log 1",                         // 日志内容
			modTime: time.Now().Add(-8 * 24 * time.Hour), // 8天前
		},
	}

	// 创建测试文件
	for _, f := range files {
		path := filepath.Join(tmpDir, f.name)              // 拼接文件路径
		err := os.WriteFile(path, []byte(f.content), 0644) // 写入文件
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", f.name, err)
		}
		// 设置文件修改时间
		err = os.Chtimes(path, f.modTime, f.modTime) // 设置文件修改时间
		if err != nil {
			t.Fatalf("Failed to set file time for %s: %v", f.name, err)
		}
	}

	// 创建清理器
	config := logrus.CleanupConfig{
		Enable:     true,                   // 启用清理
		MaxBackups: 2,                      // 只保留2个最新的文件
		MaxAge:     7,                      // 7天后过期
		Interval:   100 * time.Millisecond, // 100毫秒检查一次
		LogPaths:   []string{tmpDir},       // 添加日志路径
	}
	cleaner := logrus.NewLogCleaner(config) // 创建清理器

	// 启动清理服务
	cleaner.Start()

	// 等待足够时间让清理发生
	time.Sleep(200 * time.Millisecond)

	// 停止服务
	cleaner.Stop()

	// 验证结果
	remainingFiles, err := filepath.Glob(filepath.Join(tmpDir, "*.log"))
	if err != nil {
		t.Fatalf("Failed to list remaining files: %v", err) // 列出剩余文件失败
	}

	// 应该只剩下2个最新的文件
	if len(remainingFiles) != 2 {
		t.Errorf("Expected 2 files to remain, but got %d", len(remainingFiles)) // 期望剩下2个文件，但实际剩下%d个
	}

	// 验证是否保留了最新的两个文件
	for _, name := range []string{"recent1.log", "recent2.log"} {
		if _, err := os.Stat(filepath.Join(tmpDir, name)); os.IsNotExist(err) {
			t.Errorf("File %s should exist", name) // 文件%s应该存在
		}
	}

	// 验证旧文件是否被删除
	for _, name := range []string{"recent3.log", "old1.log"} {
		if _, err := os.Stat(filepath.Join(tmpDir, name)); !os.IsNotExist(err) {
			t.Errorf("File %s should be deleted", name) // 文件%s应该被删除
		}
	}
}

// TestDisabledCleaner 测试禁用清理器
func TestDisabledCleaner(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir() // 创建临时目录

	// 创建测试文件
	testFile := filepath.Join(tmpDir, "test.log")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err) // 创建测试文件失败
	}

	// 设置为一个旧时间
	oldTime := time.Now().Add(-10 * 24 * time.Hour)
	err = os.Chtimes(testFile, oldTime, oldTime)
	if err != nil {
		t.Fatalf("Failed to set file time: %v", err) // 设置文件时间失败
	}

	// 创建禁用的清理器
	config := logrus.CleanupConfig{
		Enable:     false,                  // 禁用清理
		MaxBackups: 1,                      // 保留1个备份
		MaxAge:     1,                      // 1天后过期
		Interval:   100 * time.Millisecond, // 100毫秒检查一次
	}
	cleaner := logrus.NewLogCleaner(config) // 创建清理器

	// 启动服务
	cleaner.Start()

	// 等待一段时间
	time.Sleep(200 * time.Millisecond)

	// 停止服务
	cleaner.Stop()

	// 验证文件仍然存在
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("File should still exist when cleaner is disabled") // 文件应该在清理器禁用时仍然存在
	}
}
