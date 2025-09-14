package common

import (
	"os"
	"strings"
	"sync"
	"testing"

	"idm/inner/validator"

	"github.com/stretchr/testify/assert"
)

func writeDotEnvFile(path, content string) {
	_ = os.WriteFile(path, []byte(content), 0644)
}

func buildDotEnv(vars map[string]string) string {
	var b strings.Builder
	for k, v := range vars {
		b.WriteString(k)
		b.WriteString("=")
		b.WriteString(v)
		b.WriteString("\n")
	}
	return b.String()
}

// Глобальный мьютекс для безопасной работы с os.Environ
var envMu sync.Mutex

// снимок текущего окружения → map
func snapshotEnv() map[string]string {
	out := make(map[string]string, 64)
	for _, kv := range os.Environ() {
		i := strings.IndexByte(kv, '=')
		if i <= 0 {
			continue
		}
		out[kv[:i]] = kv[i+1:]
	}
	return out
}

// полное восстановление окружения к снимку
func restoreEnv(prev map[string]string) {
	cur := snapshotEnv()
	// unset всего, чего не было раньше
	for k := range cur {
		if _, ok := prev[k]; !ok {
			_ = os.Unsetenv(k)
		}
	}
	// выставить все прежние значения
	for k, v := range prev {
		_ = os.Setenv(k, v)
	}
}

// Выполнить fn внутри критической секции, с откатом окружения
func withCleanEnv(fn func()) {
	envMu.Lock()
	defer envMu.Unlock()
	prev := snapshotEnv()
	defer restoreEnv(prev)
	fn()
}

func Test_Config_Loading_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		dotEnvVars   map[string]string // что пишем в .env
		envOverlay   map[string]string // что симулируем в "окружении процесса"
		expectDriver string
		expectDSN    string
	}{
		{
			name:         "NoDotEnv_OnlyEnvVars",
			dotEnvVars:   nil, // .env отсутствует
			envOverlay:   map[string]string{"DB_DRIVER_NAME": "postgres", "DB_DSN": "test-from-env"},
			expectDriver: "postgres",
			expectDSN:    "test-from-env",
		},
		{
			name:         "DotEnvEmptyAndNoEnv",
			dotEnvVars:   map[string]string{}, // пустой .env
			envOverlay:   map[string]string{}, // и пустое окружение
			expectDriver: "",
			expectDSN:    "",
		},
		{
			name:         "DotEnvEmpty_EnvVarsSet",
			dotEnvVars:   map[string]string{},
			envOverlay:   map[string]string{"DB_DRIVER_NAME": "env-driver", "DB_DSN": "env-dsn"},
			expectDriver: "env-driver",
			expectDSN:    "env-dsn",
		},
		{
			name:         "DotEnvOnly",
			dotEnvVars:   map[string]string{"DB_DRIVER_NAME": "dotenv-driver", "DB_DSN": "dotenv-dsn"},
			envOverlay:   map[string]string{},
			expectDriver: "dotenv-driver",
			expectDSN:    "dotenv-dsn",
		},
		{
			name:         "DotEnvAndEnvConflict_ENV_Wins",
			dotEnvVars:   map[string]string{"DB_DRIVER_NAME": "dotenv-driver", "DB_DSN": "dotenv-dsn"},
			envOverlay:   map[string]string{"DB_DRIVER_NAME": "env-driver", "DB_DSN": "env-dsn"},
			expectDriver: "env-driver",
			expectDSN:    "env-dsn",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			envPath := t.TempDir() + "/.env"
			if tc.dotEnvVars != nil {
				writeDotEnvFile(envPath, buildDotEnv(tc.dotEnvVars))
			} else {
				// не создаём файл .env вообще
				envPath = t.TempDir() + "/noenv"
			}

			withCleanEnv(func() {
				// Накладываем «виртуальное окружение» поверх снимка,
				// чтобы проверить приоритет ENV над .env
				for k, v := range tc.envOverlay {
					_ = os.Setenv(k, v)
				}
				cfg := GetConfig(envPath)
				assert.Equal(t, tc.expectDriver, cfg.DbDriverName)
				assert.Equal(t, tc.expectDSN, cfg.Dsn)
			})
		})
	}
}

// Таблица: валидация TLS-полей (требуются оба: SSL_CERT и SSL_KEY)
func Test_Config_TLSValidation_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		dotEnvVars map[string]string // соберём минимальный валидный конфиг + вариации TLS
		wantError  bool
	}{
		{
			name: "missing_both",
			dotEnvVars: map[string]string{
				"DB_DRIVER_NAME": "driver", "DB_DSN": "dsn",
				"APP_NAME": "app", "APP_VERSION": "1.0",
				"KEYCLOAK_JWK_URL": "http://localhost:9990/realms/idm/protocol/openid-connect/certs",
				// TLS отсутствует
			},
			wantError: true,
		},
		{
			name: "only_cert",
			dotEnvVars: map[string]string{
				"DB_DRIVER_NAME": "driver", "DB_DSN": "dsn",
				"APP_NAME": "app", "APP_VERSION": "1.0",
				"KEYCLOAK_JWK_URL": "http://localhost:9990/realms/idm/protocol/openid-connect/certs",
				"SSL_CERT":         "/certs/ssl.crt",
			},
			wantError: true,
		},
		{
			name: "only_key",
			dotEnvVars: map[string]string{
				"DB_DRIVER_NAME": "driver", "DB_DSN": "dsn",
				"APP_NAME": "app", "APP_VERSION": "1.0",
				"KEYCLOAK_JWK_URL": "http://localhost:9990/realms/idm/protocol/openid-connect/certs",
				"SSL_KEY":          "/certs/ssl.key",
			},
			wantError: true,
		},
		{
			name: "both_present",
			dotEnvVars: map[string]string{
				"DB_DRIVER_NAME": "driver", "DB_DSN": "dsn",
				"APP_NAME": "app", "APP_VERSION": "1.0",
				"KEYCLOAK_JWK_URL": "http://localhost:9990/realms/idm/protocol/openid-connect/certs",
				"SSL_CERT":         "/certs/ssl.crt", "SSL_KEY": "/certs/ssl.key",
			},
			wantError: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			envPath := t.TempDir() + "/.env"
			writeDotEnvFile(envPath, buildDotEnv(tc.dotEnvVars))

			withCleanEnv(func() {
				cfg := GetConfig(envPath)
				v := validator.New()
				err := v.Validate(cfg)
				if tc.wantError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		})
	}
}
