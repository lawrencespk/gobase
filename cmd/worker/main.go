package main

import (
	"gobase/pkg/config"
	"log"
)

func main() {
	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// ... 其他初始化代码 ...
}
