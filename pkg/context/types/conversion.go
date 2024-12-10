package types

import "time"

// GetIntValue 获取整数值
func GetIntValue(ctx Context, key string) (int, bool) {
	val, ok := ctx.GetValue(key).(int)
	return val, ok
}

// GetFloat64Value 获取浮点数值
func GetFloat64Value(ctx Context, key string) (float64, bool) {
	val, ok := ctx.GetValue(key).(float64)
	return val, ok
}

// GetBoolValue 获取布尔值
func GetBoolValue(ctx Context, key string) (bool, bool) {
	val, ok := ctx.GetValue(key).(bool)
	return val, ok
}

// GetTimeValue 获取时间值
func GetTimeValue(ctx Context, key string) (time.Time, bool) {
	val, ok := ctx.GetValue(key).(time.Time)
	return val, ok
}
