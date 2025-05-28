package common

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	DbDriverName string `validate:"required"`
	Dsn          string `validate:"required"`
}

func GetConfig(envFile string) Config {
	_ = godotenv.Load(envFile)
	var cfg = Config{
		DbDriverName: os.Getenv("DB_DRIVER_NAME"),
		Dsn:          os.Getenv("DB_DSN"),
	}
	return cfg
}
