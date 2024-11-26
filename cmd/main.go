package main

import (
	"gobase/pkg/logger/bootstrap"
	"gobase/pkg/logger/types"
)

func main() {
	// 初始化日志系统
	bootstrap.Initialize()

	// 获取默认日志器
	logger, err := bootstrap.GetLogger("default")
	if err != nil {
		panic(err)
	}

	// 使用日志器
	logger.WithFields(types.Fields{
		"user": "john",
	}).Info("User logged in")

	// 获取 ELK 日志器
	elkLogger, err := bootstrap.GetLogger("elk")
	if err != nil {
		panic(err)
	}

	elkLogger.WithFields(types.Fields{
		"component": "database",
	}).Error("Database connection error")
}
