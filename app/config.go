package app

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config is a application configuration structure
type AppConfig struct {
	Database   DatabaseConfig
	ConfigFile string

}

var Database *DatabaseConfig

func Setup() {

	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("Error loading .env file:", err)
	}

	Http := &AppConfig{
		Database: DatabaseConfig{
			Driver:   os.Getenv("DB_DRIVER"),
			Host:     os.Getenv("DB_HOST"),
			Username: os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASSWORD"),
			DBName:   os.Getenv("DB_NAME"),
			Port:     getEnvAsInt("DB_PORT", 4000),
			Debug:    os.Getenv("DB_DEBUG") == "true",
		},
	}

	Http.Database.Setup()
	Database = &Http.Database
	InitRedis()

}

func Config(key string) string {
	return os.Getenv(key)
}

// Helper convert env -> int
func getEnvAsInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}
