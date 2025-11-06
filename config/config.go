package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config 应用程序配置
type Config struct {
	ServerName    string
	ServerVersion string
	Port          string
	Host          string
	GinMode       string
}

// Load 加载配置
func Load() *Config {
	// 尝试加载 .env 文件，如果不存在也不报错
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}

	return &Config{
		ServerName:    getEnv("SERVER_NAME", "Demo 🚀"),
		ServerVersion: getEnv("SERVER_VERSION", "1.0.0"),
		Port:          getEnv("PORT", "8080"),
		Host:          getEnv("HOST", "0.0.0.0"),
		GinMode:       getEnv("GIN_MODE", "release"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}