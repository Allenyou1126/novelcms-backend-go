package utils

import (
	"github.com/joho/godotenv"
	"os"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		// Do Nothing
	}
}

func GetConfig(key string) string {
	return os.Getenv(key)
}

func GetConfigWithDefault(key string, defaultValue string) string {
	ret, exist := os.LookupEnv(key)
	if !exist {
		return defaultValue
	}
	return ret
}
