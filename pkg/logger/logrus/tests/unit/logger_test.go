package unit

import (
	"errors"
	"testing"

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
}
