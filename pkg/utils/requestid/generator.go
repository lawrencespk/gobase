package requestid

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Generator 定义请求ID生成器接口
type Generator interface {
	Generate() string
}

// Options 生成器配置选项
type Options struct {
	// 生成器类型: uuid, snowflake, custom
	Type string
	// 自定义前缀
	Prefix string
	// 是否启用对象池
	EnablePool bool
	// Snowflake相关配置
	WorkerID     int64
	DatacenterID int64
}

// DefaultOptions 默认配置
func DefaultOptions() *Options {
	return &Options{
		Type:         "uuid",
		Prefix:       "",
		EnablePool:   true,
		WorkerID:     1,
		DatacenterID: 1,
	}
}

// UUIDGenerator UUID格式生成器
type UUIDGenerator struct {
	prefix string
	pool   *sync.Pool
}

// NewUUIDGenerator 创建UUID生成器
func NewUUIDGenerator(opts *Options) *UUIDGenerator {
	g := &UUIDGenerator{
		prefix: opts.Prefix,
	}

	if opts.EnablePool {
		g.pool = &sync.Pool{
			New: func() interface{} {
				return new([36]byte)
			},
		}
	}

	return g
}

// Generate 生成UUID格式请求ID
func (g *UUIDGenerator) Generate() string {
	id := uuid.New()

	if g.pool != nil {
		buf := g.pool.Get().(*[36]byte)
		defer g.pool.Put(buf)

		uuidStr := id.String()
		copy(buf[:], uuidStr)

		if g.prefix != "" {
			return g.prefix + "-" + string(buf[:len(uuidStr)])
		}
		return string(buf[:len(uuidStr)])
	}

	if g.prefix != "" {
		return g.prefix + "-" + id.String()
	}
	return id.String()
}

// SnowflakeGenerator 雪花算法生成器
type SnowflakeGenerator struct {
	mutex         sync.Mutex
	prefix        string
	lastTimestamp int64
	workerID      int64
	datacenterID  int64
	sequence      int64
}

// SetLastTimestamp 仅用于测试
func (g *SnowflakeGenerator) SetLastTimestamp(timestamp int64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.lastTimestamp = timestamp
}

// ResetLastTimestamp 重置最后时间戳（仅用于测试）
func (g *SnowflakeGenerator) ResetLastTimestamp() {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.lastTimestamp = time.Now().UnixNano() / 1000000
}

// GetLastTimestamp 仅用于测试
func (g *SnowflakeGenerator) GetLastTimestamp() int64 {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return g.lastTimestamp
}

// NewSnowflakeGenerator 创建Snowflake生成器
func NewSnowflakeGenerator(opts *Options) *SnowflakeGenerator {
	return &SnowflakeGenerator{
		prefix:       opts.Prefix,
		workerID:     opts.WorkerID,
		datacenterID: opts.DatacenterID,
	}
}

// Generate 生成Snowflake格式请求ID
func (g *SnowflakeGenerator) Generate() string {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	timestamp := time.Now().UnixNano() / 1000000 // 毫秒

	if timestamp < g.lastTimestamp {
		// 时钟回拨,抛出异常
		panic(fmt.Sprintf("Clock moved backwards. Refusing to generate id for %d milliseconds", g.lastTimestamp-timestamp))
	}

	if timestamp == g.lastTimestamp {
		g.sequence = (g.sequence + 1) & 4095 // sequence最大4095
		if g.sequence == 0 {
			// sequence溢出,等待下一毫秒
			timestamp = g.waitNextMillis(g.lastTimestamp)
		}
	} else {
		g.sequence = 0
	}

	g.lastTimestamp = timestamp

	id := ((timestamp - 1288834974657) << 22) | // timestamp减去开始时间戳后左移22位
		(g.datacenterID << 17) | // datacenterID左移17位
		(g.workerID << 12) | // workerID左移12位
		g.sequence // sequence

	if g.prefix != "" {
		return fmt.Sprintf("%s-%d", g.prefix, id)
	}
	return fmt.Sprintf("%d", id)
}

// waitNextMillis 等待下一毫秒
func (g *SnowflakeGenerator) waitNextMillis(lastTimestamp int64) int64 {
	timestamp := time.Now().UnixNano() / 1000000
	for timestamp <= lastTimestamp {
		timestamp = time.Now().UnixNano() / 1000000
	}
	return timestamp
}

// CustomGenerator 自定义格式生成器
type CustomGenerator struct {
	prefix    string
	generator func() string
}

// NewCustomGenerator 创建自定义生成器
func NewCustomGenerator(prefix string, generator func() string) *CustomGenerator {
	return &CustomGenerator{
		prefix:    prefix,
		generator: generator,
	}
}

// Generate 生成自定义格式请求ID
func (g *CustomGenerator) Generate() string {
	id := g.generator()
	if g.prefix != "" {
		return g.prefix + "-" + id
	}
	return id
}

// NewGenerator 创建请求ID生成器
func NewGenerator(opts *Options) Generator {
	if opts == nil {
		opts = DefaultOptions()
	}

	switch opts.Type {
	case "uuid":
		return NewUUIDGenerator(opts)
	case "snowflake":
		return NewSnowflakeGenerator(opts)
	default:
		return NewUUIDGenerator(opts)
	}
}

// ValidateRequestID 验证请求ID格式是否正确
func ValidateRequestID(requestID string) bool {
	// TODO: 实现验证逻辑
	return true
}

// ParseRequestID 解析请求ID，返回其组成部分
func ParseRequestID(requestID string) (timestamp string, uuid string, ok bool) {
	// TODO: 实现解析逻辑
	return "", "", true
}
