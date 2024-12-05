package benchmark

import (
	"bytes"
	"testing"

	"github.com/sirupsen/logrus"
)

func BenchmarkLogrusFormatter(b *testing.B) {
	// 创建一个新的 Logrus logger
	logger := logrus.New()

	// 设置不同的格式化器
	formatter := &logrus.JSONFormatter{}
	logger.SetFormatter(formatter)

	// 定义不同的日志级别
	levels := []logrus.Level{
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
	}

	// 为每个日志级别运行基准测试
	for _, level := range levels {
		b.Run(level.String(), func(b *testing.B) {
			// 设置日志级别
			logger.SetLevel(level)

			// 创建一个缓冲区来捕获日志输出
			var buf bytes.Buffer
			logger.SetOutput(&buf)

			// 重置计时器以排除设置时间
			b.ResetTimer()

			// 运行基准测试
			for i := 0; i < b.N; i++ {
				logger.WithFields(logrus.Fields{
					"key1": "value1",
					"key2": "value2",
				}).Log(level, "This is a benchmark test log message")
			}
		})
	}
}
