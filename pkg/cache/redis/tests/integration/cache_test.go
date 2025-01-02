package integration

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"gobase/pkg/cache/redis"
	redisClient "gobase/pkg/client/redis"
	"gobase/pkg/client/redis/tests/testutils"
	"gobase/pkg/errors"
	"gobase/pkg/errors/codes"

	"github.com/stretchr/testify/suite"
)

type RedisCacheTestSuite struct {
	suite.Suite
	cache  *redis.Cache
	client redisClient.Client
	ctx    context.Context
}

func (s *RedisCacheTestSuite) SetupSuite() {
	// 启动Redis容器
	addr, err := testutils.StartRedisSingleContainer()
	s.Require().NoError(err)

	// 创建Redis客户端
	s.client, err = redisClient.NewClient(
		redisClient.WithAddresses([]string{addr}),
		redisClient.WithDialTimeout(time.Second*5),
	)
	s.Require().NoError(err)

	// 创建缓存实例
	s.cache, err = redis.NewCache(redis.Options{
		Client: s.client,
	})
	s.Require().NoError(err)

	s.ctx = context.Background()
}

func (s *RedisCacheTestSuite) TearDownSuite() {
	s.client.Close()
	testutils.CleanupRedisContainers()
}

func (s *RedisCacheTestSuite) TestBasicOperations() {
	type TestStruct struct {
		Name  string
		Value int
	}

	testData := TestStruct{
		Name:  "test",
		Value: 123,
	}

	// 测试Set
	err := s.cache.Set(s.ctx, "test_key", testData, time.Minute)
	s.Require().NoError(err)

	// 测试Get
	value, err := s.cache.Get(s.ctx, "test_key")
	s.Require().NoError(err)

	// 将 interface{} 转换为 TestStruct
	result, ok := value.(map[string]interface{})
	s.Require().True(ok, "value should be a map")

	// 验证字段值
	s.Equal(testData.Name, result["Name"])
	s.Equal(float64(testData.Value), result["Value"]) // JSON 数字会被解析为 float64

	// 测试Delete
	err = s.cache.Delete(s.ctx, "test_key")
	s.Require().NoError(err)

	// 验证删除后无法获取
	_, err = s.cache.Get(s.ctx, "test_key")
	s.Require().Error(err)
	s.True(errors.HasErrorCode(err, codes.RedisKeyNotFoundError))
}

// TestConcurrentOperations 测试并发操作
func (s *RedisCacheTestSuite) TestConcurrentOperations() {
	const (
		goroutines = 10
		operations = 100
	)

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < operations; j++ {
				key := fmt.Sprintf("concurrent_key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)

				// 测试Set
				err := s.cache.Set(s.ctx, key, value, time.Minute)
				s.Require().NoError(err)

				// 测试Get
				result, err := s.cache.Get(s.ctx, key)
				s.Require().NoError(err)
				s.Equal(value, result)

				// 测试Delete
				err = s.cache.Delete(s.ctx, key)
				s.Require().NoError(err)
			}
		}(i)
	}
	wg.Wait()
}

// TestExpiration 测试过期机制
func (s *RedisCacheTestSuite) TestExpiration() {
	key := "expiring_key"
	value := "expiring_value"

	// 设置短期过期时间
	err := s.cache.Set(s.ctx, key, value, time.Second)
	s.Require().NoError(err)

	// 立即获取应该成功
	result, err := s.cache.Get(s.ctx, key)
	s.Require().NoError(err)
	s.Equal(value, result)

	// 等待过期
	time.Sleep(time.Second * 2)

	// 获取应该失败
	_, err = s.cache.Get(s.ctx, key)
	s.Require().Error(err)
	s.True(errors.HasErrorCode(err, codes.RedisKeyNotFoundError))
}

// TestClear 测试清空缓存
func (s *RedisCacheTestSuite) TestClear() {
	// 设置多个键值对
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("clear_test_key_%d", i)
		value := fmt.Sprintf("value_%d", i)
		err := s.cache.Set(s.ctx, key, value, time.Minute)
		s.Require().NoError(err)
	}

	// 清空缓存
	err := s.cache.Clear(s.ctx)
	s.Require().NoError(err)

	// 验证所有键都已被删除
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("clear_test_key_%d", i)
		_, err := s.cache.Get(s.ctx, key)
		s.Require().Error(err)
		s.True(errors.HasErrorCode(err, codes.RedisKeyNotFoundError))
	}
}

// TestInvalidInputs 测试无效输入
func (s *RedisCacheTestSuite) TestInvalidInputs() {
	// 测试空键
	err := s.cache.Set(s.ctx, "", "value", time.Minute)
	s.Require().Error(err)

	// 测试nil值
	err = s.cache.Set(s.ctx, "nil_key", nil, time.Minute)
	s.Require().Error(err)

	// 测试负过期时间
	err = s.cache.Set(s.ctx, "negative_ttl", "value", -time.Second)
	s.Require().Error(err)
}

// TestComplexDataTypes 测试复杂数据类型
func (s *RedisCacheTestSuite) TestComplexDataTypes() {
	type ComplexStruct struct {
		IntArray   []int
		StringMap  map[string]string
		NestedData struct {
			Field1 int
			Field2 string
		}
	}

	testData := ComplexStruct{
		IntArray:  []int{1, 2, 3},
		StringMap: map[string]string{"key1": "value1", "key2": "value2"},
		NestedData: struct {
			Field1 int
			Field2 string
		}{
			Field1: 42,
			Field2: "nested",
		},
	}

	// 测试Set
	err := s.cache.Set(s.ctx, "complex_key", testData, time.Minute)
	s.Require().NoError(err)

	// 测试Get
	value, err := s.cache.Get(s.ctx, "complex_key")
	s.Require().NoError(err)

	// 验证复杂数据结构
	result, ok := value.(map[string]interface{})
	s.Require().True(ok)

	// 验证数组
	intArray, ok := result["IntArray"].([]interface{})
	s.Require().True(ok)
	s.Equal(3, len(intArray))

	// 验证Map
	stringMap, ok := result["StringMap"].(map[string]interface{})
	s.Require().True(ok)
	s.Equal(2, len(stringMap))

	// 验证嵌套结构
	nestedData, ok := result["NestedData"].(map[string]interface{})
	s.Require().True(ok)
	s.Equal(float64(42), nestedData["Field1"])
	s.Equal("nested", nestedData["Field2"])
}

func TestRedisCache(t *testing.T) {
	suite.Run(t, new(RedisCacheTestSuite))
}
