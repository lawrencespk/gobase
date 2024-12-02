package unit

import (
	"bytes"
	"sync"
	"testing"

	"gobase/pkg/logger/logrus"
)

// 测试BufferPool
func TestBufferPool(t *testing.T) {
	t.Run("Basic Pool Operations", func(t *testing.T) {
		pool := logrus.NewBufferPool()

		// 获取缓冲区
		buf := pool.Get()
		if buf == nil {
			t.Fatal("Expected non-nil buffer from pool")
		}

		// 写入数据
		testData := []byte("test data")
		buf.Write(testData)

		// 验证数据
		if !bytes.Equal(buf.Bytes(), testData) {
			t.Errorf("Expected %v, got %v", testData, buf.Bytes())
		}

		// 放回池中
		pool.Put(buf)

		// 再次获取，应该被重置
		buf = pool.Get()
		if buf.Len() != 0 {
			t.Errorf("Expected empty buffer, got length %d", buf.Len())
		}
	})

	t.Run("Concurrent Access", func(t *testing.T) {
		pool := logrus.NewBufferPool()
		var wg sync.WaitGroup
		workers := 100
		iterations := 1000

		wg.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					buf := pool.Get()
					buf.WriteString("test")
					pool.Put(buf)
				}
			}()
		}
		wg.Wait()
	})
}

// 测试WritePool
func TestWritePool(t *testing.T) {
	t.Run("Basic WritePool Operations", func(t *testing.T) {
		size := 1024
		pool := logrus.NewWritePool(size)

		// 获取缓冲区
		buf := pool.Get()
		if cap(buf) != size {
			t.Errorf("Expected capacity %d, got %d", size, cap(buf))
		}

		// 写入数据
		data := []byte("test data")
		buf = append(buf, data...)

		// 验证数据
		if !bytes.Equal(buf, data) {
			t.Errorf("Expected %v, got %v", data, buf)
		}

		// 放回池中
		pool.Put(buf)

		// 再次获取，应该被重置
		buf = pool.Get()
		if len(buf) != 0 {
			t.Errorf("Expected empty buffer, got length %d", len(buf))
		}
	})

	t.Run("Concurrent WritePool Access", func(t *testing.T) {
		pool := logrus.NewWritePool(1024)
		var wg sync.WaitGroup
		workers := 100
		iterations := 1000

		wg.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					buf := pool.Get()
					buf = append(buf, []byte("test")...)
					pool.Put(buf)
				}
			}()
		}
		wg.Wait()
	})
}

// 测试BufferPool的性能
func BenchmarkBufferPool(b *testing.B) {
	pool := logrus.NewBufferPool()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get()
			buf.WriteString("benchmark test data")
			pool.Put(buf)
		}
	})
}

// 测试WritePool的性能
func BenchmarkWritePool(b *testing.B) {
	pool := logrus.NewWritePool(1024)
	data := []byte("benchmark test data")
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := pool.Get()
			buf = append(buf, data...)
			pool.Put(buf)
		}
	})
}
