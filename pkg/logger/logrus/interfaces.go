package logrus

// Writer 定义写入接口
type Writer interface {
	Write(p []byte) (n int, err error)
}
