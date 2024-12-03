package logger

import (
	"time"
)

type FieldType uint8

const (
	StringType   FieldType = iota // 字符串类型
	IntType                       // 整数类型
	Int64Type                     // 64位整数类型
	Float64Type                   // 浮点数类型
	BoolType                      // 布尔类型
	TimeType                      // 时间类型
	DurationType                  // 持续时间类型
	ObjectType                    // 对象类型
	ErrorType                     // 错误类型
)

// Field 日志字段
type Field struct {
	Key   string      // 字段名
	Value interface{} // 字段值
	Type  FieldType   // 字段类型
}

// String 创建字符串字段
func String(key string, value string) Field {
	return Field{Key: key, Value: value, Type: StringType}
}

// Int 创建整数字段
func Int(key string, value int) Field {
	return Field{Key: key, Value: value, Type: IntType}
}

// Int64 创建64位整数字段
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value, Type: Int64Type}
}

// Float64 创建浮点数字段
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value, Type: Float64Type}
}

// Bool 创建布尔字段
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value, Type: BoolType}
}

// Error 创建错误字段
func Error(err error) Field {
	return Field{Key: "error", Value: err, Type: ErrorType}
}

// Time 创建时间类型字段
func Time(key string, value time.Time) Field {
	return Field{
		Key:   key,
		Value: value,
		Type:  TimeType,
	}
}

// Duration 创建持续时间类型字段
func Duration(key string, value time.Duration) Field {
	return Field{
		Key:   key,
		Value: value,
		Type:  DurationType,
	}
}

// Object 创建对象类型字段
func Object(key string, value interface{}) Field {
	return Field{
		Key:   key,
		Value: value,
		Type:  ObjectType,
	}
}
