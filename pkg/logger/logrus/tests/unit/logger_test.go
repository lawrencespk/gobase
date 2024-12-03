package unit

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"gobase/pkg/logger"
)

// TestStringField 测试字符串字段创建
func TestStringField(t *testing.T) {
	key := "test_key"
	value := "test_value"
	field := logger.String(key, value)

	if field.Key != key {
		t.Errorf("期望 key 为 %s, 实际得到 %s", key, field.Key)
	}
	if field.Value != value {
		t.Errorf("期望 value 为 %s, 实际得到 %v", value, field.Value)
	}
	if field.Type != logger.StringType {
		t.Errorf("期望 type 为 StringType, 实际得到 %v", field.Type)
	}
}

// TestIntField 测试整数字段创建
func TestIntField(t *testing.T) {
	key := "test_int"
	value := 123
	field := logger.Int(key, value)

	if field.Key != key {
		t.Errorf("期望 key 为 %s, 实际得到 %s", key, field.Key)
	}
	if field.Value != value {
		t.Errorf("期望 value 为 %d, 实际得到 %v", value, field.Value)
	}
	if field.Type != logger.IntType {
		t.Errorf("期望 type 为 IntType, 实际得到 %v", field.Type)
	}
}

// TestInt64Field 测试64位整数字段创建
func TestInt64Field(t *testing.T) {
	key := "test_int64"
	value := int64(9223372036854775807)
	field := logger.Int64(key, value)

	if field.Key != key {
		t.Errorf("期望 key 为 %s, 实际得到 %s", key, field.Key)
	}
	if field.Value != value {
		t.Errorf("期望 value 为 %d, 实际得到 %v", value, field.Value)
	}
	if field.Type != logger.Int64Type {
		t.Errorf("期望 type 为 Int64Type, 实际得到 %v", field.Type)
	}
}

// TestFloat64Field 测试浮点数字段创建
func TestFloat64Field(t *testing.T) {
	key := "test_float64"
	value := 3.14159
	field := logger.Float64(key, value)

	if field.Key != key {
		t.Errorf("期望 key 为 %s, 实际得到 %s", key, field.Key)
	}
	if field.Value != value {
		t.Errorf("期望 value 为 %f, 实际得到 %v", value, field.Value)
	}
	if field.Type != logger.Float64Type {
		t.Errorf("期望 type 为 Float64Type, 实际得到 %v", field.Type)
	}
}

// TestBoolField 测试布尔字段创建
func TestBoolField(t *testing.T) {
	key := "test_bool"
	value := true
	field := logger.Bool(key, value)

	if field.Key != key {
		t.Errorf("期望 key 为 %s, 实际得到 %s", key, field.Key)
	}
	if field.Value != value {
		t.Errorf("期望 value 为 %t, 实际得到 %v", value, field.Value)
	}
	if field.Type != logger.BoolType {
		t.Errorf("期望 type 为 BoolType, 实际得到 %v", field.Type)
	}
}

// TestErrorField 测试错误字段创建
func TestErrorField(t *testing.T) {
	err := errors.New("test error")
	field := logger.Error(err)

	if field.Key != "error" {
		t.Errorf("期望 key 为 'error', 实际得到 %s", field.Key)
	}
	if field.Value != err {
		t.Errorf("期望 value 为 %v, 实际得到 %v", err, field.Value)
	}
	if field.Type != logger.ErrorType {
		t.Errorf("期望 type 为 ErrorType, 实际得到 %v", field.Type)
	}
}

// TestFieldsWithNilValues 测试空值处理
func TestFieldsWithNilValues(t *testing.T) {
	t.Run("Nil Error Field", func(t *testing.T) {
		field := logger.Error(nil)
		if field.Value != nil {
			t.Error("期望错误字段的空值为 nil")
		}
	})

	t.Run("Empty String Field", func(t *testing.T) {
		field := logger.String("empty", "")
		if field.Value != "" {
			t.Error("期望空字符串字段的值为空字符串")
		}
	})
}

// TestFieldValueConversion 测试字段值类型转换
func TestFieldValueConversion(t *testing.T) {
	t.Run("Int to Int64", func(t *testing.T) {
		field := logger.Int("key", 123)
		if _, ok := field.Value.(int); !ok {
			t.Error("期望 int 类型值保持不变")
		}
	})

	t.Run("Float32 to Float64", func(t *testing.T) {
		value := float32(3.14)
		field := logger.Float64("key", float64(value))
		if _, ok := field.Value.(float64); !ok {
			t.Error("期望 float32 被转换为 float64")
		}
	})
}

// TestInvalidFieldKeys 测试无效的字段键
func TestInvalidFieldKeys(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"Empty Key", ""},
		{"Very Long Key", string(make([]byte, 1024))},
		{"Special Characters", "test@#$%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := logger.String(tt.key, "value")
			if field.Key != tt.key {
				t.Errorf("期望 key 为 %s, 实际得到 %s", tt.key, field.Key)
			}
		})
	}
}

// TestConcurrentFieldCreation 测试并发字段创建
func TestConcurrentFieldCreation(t *testing.T) {
	const goroutines = 100
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				_ = logger.String(key, "value")
				_ = logger.Int(key, j)
				_ = logger.Bool(key, true)
			}
		}(i)
	}

	wg.Wait()
}

// TestFieldValueOverflow 测试字段值溢出
func TestFieldValueOverflow(t *testing.T) {
	t.Run("Max Int64", func(t *testing.T) {
		value := int64(9223372036854775807)
		field := logger.Int64("key", value)
		if field.Value != value {
			t.Error("期望正确处理最大 int64 值")
		}
	})

	t.Run("Min Int64", func(t *testing.T) {
		value := int64(-9223372036854775808)
		field := logger.Int64("key", value)
		if field.Value != value {
			t.Error("期望正确处理最小 int64 值")
		}
	})
}

// TestFieldMemoryUsage 测试字段内存使用
func TestFieldMemoryUsage(t *testing.T) {
	t.Run("Large String", func(t *testing.T) {
		largeString := string(make([]byte, 1<<20)) // 1MB string
		field := logger.String("key", largeString)
		if field.Value != largeString {
			t.Error("期望正确处理大字符串")
		}
	})
}

// TestTimeField 测试时间字段创建
func TestTimeField(t *testing.T) {
	now := time.Now()
	field := logger.Time("timestamp", now)

	if field.Key != "timestamp" {
		t.Errorf("期望 key 为 'timestamp', 实际得到 %s", field.Key)
	}
	if field.Value != now {
		t.Errorf("期望 value 为 %v, 实际得到 %v", now, field.Value)
	}
	if field.Type != logger.TimeType {
		t.Errorf("期望 type 为 TimeType, 实际得到 %v", field.Type)
	}
}

// TestDurationField 测试持续时间字段创建
func TestDurationField(t *testing.T) {
	duration := 5 * time.Second
	field := logger.Duration("elapsed", duration)

	if field.Key != "elapsed" {
		t.Errorf("期望 key 为 'elapsed', 实际得到 %s", field.Key)
	}
	if field.Value != duration {
		t.Errorf("期望 value 为 %v, 实际得到 %v", duration, field.Value)
	}
	if field.Type != logger.DurationType {
		t.Errorf("期望 type 为 DurationType, 实际得到 %v", field.Type)
	}
}

// TestObjectField 测试对象字段创建
func TestObjectField(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}
	user := User{Name: "test", Age: 18}
	field := logger.Object("user", user)

	if field.Key != "user" {
		t.Errorf("期望 key 为 'user', 实际得到 %s", field.Key)
	}
	if v, ok := field.Value.(User); !ok || v != user {
		t.Errorf("期望 value 为 %v, 实际得到 %v", user, field.Value)
	}
	if field.Type != logger.ObjectType {
		t.Errorf("期望 type 为 ObjectType, 实际得到 %v", field.Type)
	}
}

// TestFieldValueValidation 测试字段值验证
func TestFieldValueValidation(t *testing.T) {
	t.Run("Invalid Time Value", func(t *testing.T) {
		var zeroTime time.Time
		field := logger.Time("zero_time", zeroTime)
		if !zeroTime.Equal(field.Value.(time.Time)) {
			t.Error("期望零时间值被正确处理")
		}
	})

	t.Run("Negative Duration", func(t *testing.T) {
		duration := -5 * time.Second
		field := logger.Duration("negative", duration)
		if field.Value != duration {
			t.Error("期望负持续时间被正确处理")
		}
	})
}

// TestFieldCopy 测试字段复制
func TestFieldCopy(t *testing.T) {
	original := logger.String("key", "value")
	copied := original

	// 修改复制后的字段
	copied.Value = "new_value"

	// 验证原字段未被修改
	if original.Value == copied.Value {
		t.Error("期望字段复制是值复制而不是引用复制")
	}
}

// TestFieldEdgeCases 测试更多边界情况
func TestFieldEdgeCases(t *testing.T) {
	t.Run("Unicode Key", func(t *testing.T) {
		key := "测试键"
		field := logger.String(key, "value")
		if field.Key != key {
			t.Error("期望正确处理 Unicode 键名")
		}
	})

	t.Run("Zero Values", func(t *testing.T) {
		intField := logger.Int("zero_int", 0)
		if intField.Value != 0 {
			t.Error("期望正确处理零值整数")
		}

		floatField := logger.Float64("zero_float", 0.0)
		if floatField.Value != 0.0 {
			t.Error("期望正确处理零值浮点数")
		}
	})
}

// BenchmarkFieldCreation 性能测试
func BenchmarkFieldCreation(b *testing.B) {
	b.Run("String Field", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.String("key", "value")
		}
	})

	b.Run("Int Field", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			logger.Int("key", 123)
		}
	})

	b.Run("Error Field", func(b *testing.B) {
		err := errors.New("test error")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error(err)
		}
	})

	b.Run("Concurrent String Fields", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logger.String("key", "value")
			}
		})
	})

	b.Run("Large Value", func(b *testing.B) {
		largeString := string(make([]byte, 1024)) // 1KB string
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.String("key", largeString)
		}
	})
}
