package provider

import (
	"gobase/pkg/logger/elk"
	"gobase/pkg/logger/logrus"
	"gobase/pkg/logger/types" // 只导入 types 包
)

// NewLogrusLogger 创建 logrus 日志记录器
func NewLogrusLogger(cfg types.Config) (types.Logger, error) { // 返回 types.Logger
	return logrus.NewAdapter(cfg)
}

// NewElkLogger 创建 ELK 日志记录器
func NewElkLogger(cfg types.Config) (types.Logger, error) { // 返回 types.Logger
	return elk.NewAdapter(cfg)
}
