package common

import (
	"github.com/stretchr/testify/assert"
	"idm/inner/validator"
	"os"
	"testing"
)

func cleanupEnv() {
	_ = os.Unsetenv("DB_DRIVER_NAME")
	_ = os.Unsetenv("DB_DSN")
	_ = os.Unsetenv("APP_NAME")
	_ = os.Unsetenv("APP_VERSION")
	_ = os.Unsetenv("LOG_LEVEL")
	_ = os.Unsetenv("LOG_DEVELOP_MODE")
	_ = os.Unsetenv("SSL_SERT")
	_ = os.Unsetenv("SSL_KEY")
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

	cfg := GetConfig(".env")
	assert.Equal(t, "postgres", cfg.DbDriverName)
	assert.Equal(t, "test-from-env", cfg.Dsn)
}

func Test_DotEnvEmptyAndNoEnv(t *testing.T) {
	cleanupEnv()
	writeDotEnv("")
	defer removeDotEnv()

	cfg := GetConfig(".env")
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

	cfg := GetConfig(".env")
	assert.Equal(t, "env-driver", cfg.DbDriverName)
	assert.Equal(t, "env-dsn", cfg.Dsn)
}

func Test_DotEnvOnly(t *testing.T) {
	cleanupEnv()
	writeDotEnv(`DB_DRIVER_NAME=dotenv-driver
DB_DSN=dotenv-dsn`)
	defer removeDotEnv()

	cfg := GetConfig(".env")
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

	cfg := GetConfig(".env")
	assert.Equal(t, "env-driver", cfg.DbDriverName)
	assert.Equal(t, "env-dsn", cfg.Dsn)
}

func Test_Validation_TLS_MissingBoth_ShouldFail(t *testing.T) {
	cleanupEnv()
	writeDotEnv("")
	defer removeDotEnv()
	_ = os.Setenv("DB_DRIVER_NAME", "env-driver")
	_ = os.Setenv("DB_DSN", "env-dsn")
	_ = os.Setenv("APP_NAME", "idm-test")
	_ = os.Setenv("APP_VERSION", "1.0")

	cfg := GetConfig(".env")
	v := validator.New()
	err := v.Validate(cfg)
	assert.Error(t, err, "validation should fail when both SSL_SERT and SSL_KEY are missing")
}

func Test_Validation_TLS_OnlyCert_ShouldFail(t *testing.T) {
	cleanupEnv()
	writeDotEnv("")
	defer removeDotEnv()

	_ = os.Setenv("DB_DRIVER_NAME", "env-driver")
	_ = os.Setenv("DB_DSN", "env-dsn")
	_ = os.Setenv("APP_NAME", "idm-test")
	_ = os.Setenv("APP_VERSION", "1.0")
	_ = os.Setenv("SSL_SERT", "/certs/ssl.crt")

	cfg := GetConfig(".env")
	v := validator.New()
	err := v.Validate(cfg)
	assert.Error(t, err, "validation should fail when SSL_KEY is missing")
}

func Test_Validation_TLS_OnlyKey_ShouldFail(t *testing.T) {
	cleanupEnv()
	writeDotEnv("")
	defer removeDotEnv()

	_ = os.Setenv("DB_DRIVER_NAME", "env-driver")
	_ = os.Setenv("DB_DSN", "env-dsn")
	_ = os.Setenv("APP_NAME", "idm-test")
	_ = os.Setenv("APP_VERSION", "1.0")
	_ = os.Setenv("SSL_KEY", "/certs/ssl.key")

	cfg := GetConfig(".env")
	v := validator.New()
	err := v.Validate(cfg)
	assert.Error(t, err, "validation should fail when SSL_SERT is missing")
}

func Test_Validation_TLS_BothPresent_ShouldPass(t *testing.T) {
	cleanupEnv()
	writeDotEnv("")
	defer removeDotEnv()

	_ = os.Setenv("DB_DRIVER_NAME", "env-driver")
	_ = os.Setenv("DB_DSN", "env-dsn")
	_ = os.Setenv("APP_NAME", "idm-test")
	_ = os.Setenv("APP_VERSION", "1.0")
	_ = os.Setenv("SSL_SERT", "1.0")
	_ = os.Setenv("SSL_KEY", "1.0")

	cfg := GetConfig(".env")
	v := validator.New()
	err := v.Validate(cfg)
	assert.NoError(t, err, "validation should pass when both SSL_SERT and SSL_KEY are present")
}
