package main

import (
	"fmt"
	"log"
	"time"

	"gobase/pkg/config"
	"gobase/pkg/config/types"
)

// AppConfig 应用配置结构
type AppConfig struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
	Database struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		DBName   string `json:"dbname"`
	} `json:"database"`
	Redis struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`
}

func main() {
	// 创建配置管理器
	manager, err := config.NewManager(
		// 基础配置
		types.WithConfigFile("config/config.yaml"),
		types.WithConfigType("yaml"),
		types.WithEnvironment("development"),
		types.WithEnableEnv(true),
		types.WithEnvPrefix("APP"),

		// Nacos配置（可选）
		types.WithNacosEndpoint("localhost:8848"),
		types.WithNacosNamespace("public"),
		types.WithNacosGroup("DEFAULT_GROUP"),
		types.WithNacosDataID("application.yaml"),
		types.WithNacosTimeout(5*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create config manager: %v", err)
	}
	defer manager.Close()

	// 监听配置变化
	err = manager.Watch("app.server.port", func(key string, value interface{}) {
		fmt.Printf("Config changed: %s = %v\n", key, value)
	})
	if err != nil {
		log.Printf("Failed to watch config: %v", err)
	}

	// 获取单个配置
	host := manager.GetString("app.server.host")
	port := manager.GetInt("app.server.port")
	fmt.Printf("Server: %s:%d\n", host, port)

	// 获取数据库超时配置
	timeout := manager.GetDuration("app.database.timeout")
	fmt.Printf("Database timeout: %v\n", timeout)

	// 获取完整配置结构
	var appConfig AppConfig
	if err := manager.Parse("app", &appConfig); err != nil {
		log.Printf("Failed to parse config: %v", err)
	} else {
		fmt.Printf("Database config: %+v\n", appConfig.Database)
	}
}
