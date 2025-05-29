package tests

import (
	"github.com/stretchr/testify/assert"
	"idm/inner/common"
	"idm/inner/database"
	"os"
	"testing"
)

func cleanupEnv() {
	_ = os.Unsetenv("DB_DRIVER_NAME")
	_ = os.Unsetenv("DB_DSN")
}

func writeDotEnv(content string) {
	_ = os.WriteFile(".env", []byte(content), 0644)
}

func removeDotEnv() {
	_ = os.Remove(".env")
}

func Test_DBConnectionFails(t *testing.T) {
	cfg := common.Config{
		DbDriverName: "postgres",
		Dsn:          "host=127.0.0.1 port=9999 user=fail password=fail dbname=fail sslmode=disable",
	}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic from MustConnect, got none")
		}
	}()
	_ = database.ConnectDbWithCfg(cfg)
}

func Test_DBConnectionSucceeds(t *testing.T) {
	cleanupEnv()
	removeDotEnv()
	writeDotEnv(`DB_DRIVER_NAME=postgres
DB_DSN=host=127.0.0.1 port=5432 user=postgres password=postgres dbname=postgres sslmode=disable`)
	defer removeDotEnv()
	cfg := common.GetConfig(".env")
	db := database.ConnectDbWithCfg(cfg)
	assert.NotNil(t, db)
	_ = db.Ping()
}
