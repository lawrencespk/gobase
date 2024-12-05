package elk

// Logger 定义日志接口
type Logger interface {
	Error(err error)
	Info(msg string)
	Debug(msg string)
}
