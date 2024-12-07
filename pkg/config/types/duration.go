package types

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration 包装 time.Duration 以支持 JSON 序列化/反序列化
type Duration struct {
	time.Duration
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid duration")
	}

	return nil
}

// MarshalJSON 实现 json.Marshaler 接口
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}
