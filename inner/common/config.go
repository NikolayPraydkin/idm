package common

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DbDriverName   string `validate:"required"`
	Dsn            string `validate:"required"`
	AppName        string `validate:"required"`
	AppVersion     string `validate:"required"`
	LogLevel       string
	LogDevelopMode bool
	SslSert        string `validate:"required"`
	SslKey         string `validate:"required"`
	KeycloakJwkUrl string `validate:"required"`
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
		SslSert:        os.Getenv("SSL_SERT"),
		SslKey:         os.Getenv("SSL_KEY"),
		KeycloakJwkUrl: os.Getenv("KEYCLOAK_JWK_URL"),
	}
	return cfg
}
