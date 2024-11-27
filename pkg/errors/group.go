package errors

import (
	"context"
	"fmt"
	"sync"
)

// Group 错误组
type Group struct {
	errors []error
	mutex  sync.Mutex
}

// NewErrorGroup 创建错误组
func NewErrorGroup() *Group {
	return &Group{
		errors: make([]error, 0),
	}
}

// Add 添加错误到组
func (g *Group) Add(err error) {
	if err == nil {
		return
	}
	g.mutex.Lock()
	g.errors = append(g.errors, err)
	g.mutex.Unlock()
}

// HasErrors 检查是否有错误
func (g *Group) HasErrors() bool {
	return len(g.errors) > 0
}

// GetErrors 获取所有错误
func (g *Group) GetErrors() []error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	// 返回错误切片的副本
	result := make([]error, len(g.errors))
	copy(result, g.errors)
	return result
}

// Error 实现 error 接口
func (g *Group) Error() string {
	if len(g.errors) == 0 {
		return ""
	}
	if len(g.errors) == 1 {
		return g.errors[0].Error()
	}
	return fmt.Sprintf("多个错误发生 (%d): %v", len(g.errors), g.errors[0])
}

// Go 并发执行任务并收集错误
func (g *Group) Go(ctx context.Context, fns ...func(ctx context.Context) error) {
	var wg sync.WaitGroup
	wg.Add(len(fns))

	for _, fn := range fns {
		fn := fn // 创建副本以避免闭包问题
		go func() {
			defer wg.Done()
			if err := fn(ctx); err != nil {
				g.Add(err)
			}
		}()
	}

	wg.Wait()
}

// Clear 清除所有错误
func (g *Group) Clear() {
	g.mutex.Lock()
	g.errors = g.errors[:0]
	g.mutex.Unlock()
}

// First 获取第一个错误
func (g *Group) First() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	if len(g.errors) > 0 {
		return g.errors[0]
	}
	return nil
}

// Len 获取错误数量
func (g *Group) Len() int {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return len(g.errors)
}
