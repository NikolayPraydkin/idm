package common

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	DbDriverName   string `validate:"required"`
	Dsn            string `validate:"required"`
	AppName        string `validate:"required"`
	AppVersion     string `validate:"required"`
	LogLevel       string
	LogDevelopMode bool
}

func GetConfig(envFile string) Config {
	_ = godotenv.Load(envFile)
	var cfg = Config{
		DbDriverName:   os.Getenv("DB_DRIVER_NAME"),
		Dsn:            os.Getenv("DB_DSN"),
		AppName:        os.Getenv("APP_NAME"),
		AppVersion:     os.Getenv("APP_VERSION"),
		LogLevel:       os.Getenv("LOG_LEVEL"),
		LogDevelopMode: os.Getenv("LOG_DEVELOP_MODE") == "true",
	}
	return cfg
}
