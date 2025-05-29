package database

import (
	"github.com/stretchr/testify/assert"
	"idm/inner/common"
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

func Test_NoDotEnv_OnlyEnvVars(t *testing.T) {
	cleanupEnv()
	removeDotEnv()

	_ = os.Setenv("DB_DRIVER_NAME", "postgres")
	_ = os.Setenv("DB_DSN", "test-from-env")

	cfg := common.GetConfig(".env")
	assert.Equal(t, "postgres", cfg.DbDriverName)
	assert.Equal(t, "test-from-env", cfg.Dsn)
}

func Test_DotEnvEmptyAndNoEnv(t *testing.T) {
	cleanupEnv()
	writeDotEnv("")
	defer removeDotEnv()

	cfg := common.GetConfig(".env")
	assert.Empty(t, cfg.DbDriverName)
	assert.Empty(t, cfg.Dsn)
}

func Test_DotEnvEmpty_EnvVarsSet(t *testing.T) {
	cleanupEnv()
	writeDotEnv("")
	defer removeDotEnv()
	defer cleanupEnv()

	_ = os.Setenv("DB_DRIVER_NAME", "env-driver")
	_ = os.Setenv("DB_DSN", "env-dsn")

	cfg := common.GetConfig(".env")
	assert.Equal(t, "env-driver", cfg.DbDriverName)
	assert.Equal(t, "env-dsn", cfg.Dsn)
}

func Test_DotEnvOnly(t *testing.T) {
	cleanupEnv()
	writeDotEnv(`DB_DRIVER_NAME=dotenv-driver
DB_DSN=dotenv-dsn`)
	defer removeDotEnv()

	cfg := common.GetConfig(".env")
	assert.Equal(t, "dotenv-driver", cfg.DbDriverName)
	assert.Equal(t, "dotenv-dsn", cfg.Dsn)
}

func Test_DotEnvAndEnvConflict(t *testing.T) {
	cleanupEnv()
	writeDotEnv(`DB_DRIVER_NAME=dotenv-driver
DB_DSN=dotenv-dsn`)
	defer removeDotEnv()

	_ = os.Setenv("DB_DRIVER_NAME", "env-driver")
	_ = os.Setenv("DB_DSN", "env-dsn")

	cfg := common.GetConfig(".env")
	assert.Equal(t, "env-driver", cfg.DbDriverName)
	assert.Equal(t, "env-dsn", cfg.Dsn)
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
	_ = ConnectDbWithCfg(cfg)
}

func Test_DBConnectionSucceeds(t *testing.T) {
	cleanupEnv()
	removeDotEnv()
	writeDotEnv(`DB_DRIVER_NAME=postgres
DB_DSN=host=127.0.0.1 port=5432 user=postgres password=postgres dbname=postgres sslmode=disable`)
	defer removeDotEnv()
	cfg := common.GetConfig(".env")
	db := ConnectDbWithCfg(cfg)
	assert.NotNil(t, db)
	_ = db.Ping()
}
