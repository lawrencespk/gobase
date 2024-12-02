package types

type Level uint32

const (
	DebugLevel Level = iota // 调试级别
	InfoLevel               // 信息级别
	WarnLevel               // 警告级别
	ErrorLevel              // 错误级别
	FatalLevel              // 严重级别
)

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG" // 调试级别
	case InfoLevel:
		return "INFO" // 信息级别
	case WarnLevel:
		return "WARN" // 警告级别
	case ErrorLevel:
		return "ERROR" // 错误级别
	case FatalLevel:
		return "FATAL" // 严重级别
	default:
		return "UNKNOWN" // 未知级别
	}
}

// 转换为logrus日志级别
func (l Level) ToLogrusLevel() uint32 {
	return uint32(l)
}
