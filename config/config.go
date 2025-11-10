package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server configuration
	Port        string
	Environment string

	// Database configuration
	DatabasePath string

	// IoT Platform configuration
	IotApiBaseURL  string
	IotAppKey      string
	IotAppSecret   string
	IotDeviceCode  string

	// Proxy configuration (optional)
	HttpProxy  string
	HttpsProxy string
}

var AppConfig *Config

func LoadConfig() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	AppConfig = &Config{
		Port:        getEnv("PORT", "3000"),
		Environment: getEnv("NODE_ENV", "development"),

		DatabasePath: getEnv("DATABASE_PATH", "./database/device_monitor.db"),

		IotApiBaseURL:  getEnv("IOT_API_BASE_URL", "https://iot.know-act.com"),
		IotAppKey:      getEnv("IOT_APP_KEY", ""),
		IotAppSecret:   getEnv("IOT_APP_SECRET", ""),
		IotDeviceCode:  getEnv("IOT_DEVICE_CODE", ""),

		HttpProxy:  getEnv("HTTP_PROXY", ""),
		HttpsProxy: getEnv("HTTPS_PROXY", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func IsProduction() bool {
	return AppConfig.Environment == "production"
}

func IsDevelopment() bool {
	return AppConfig.Environment == "development"
}