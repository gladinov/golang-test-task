//go:build integration

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInitConfig_Success_Integration(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("ROOT_PATH", tmpDir)
	t.Setenv("CONFIG_PATH", "config.yaml")
	t.Setenv("ENV", "local")
	t.Setenv("POSTGRES_HOST", "localhost")
	t.Setenv("POSTGRES_USER", "user")
	t.Setenv("POSTGRES_PASSWORD", "parol")
	t.Setenv("POSTGRES_DB", "service")
	t.Setenv("POSTGRES_SERVICE_PORT", "5433")
	t.Setenv("PGUSER", "user")
	t.Setenv("TEST_APP_PORT", "8080")

	// Создаем временный YAML-файл
	configPath := filepath.Join(tmpDir, "config.yaml")
	yamlContent := `
postgresHost:
  sslmode: "disable"
clients:
  testAppClient:
    testAppHost: "0.0.0.0"
`
	require.NoError(t, os.WriteFile(configPath, []byte(yamlContent), 0o644))

	want := Config{
		Env:        "local",
		RootPath:   tmpDir,
		ConfigPath: "config.yaml",
		PostgresHost: PostgresHost{
			Host:     "localhost",
			User:     "user",
			Password: "parol",
			Dbname:   "service",
			Port:     "5433",
			PgUser:   "user",
			SslMode:  "disable",
		},
		Clients: Clients{
			TestAppClient: TestAppClient{
				Host: "0.0.0.0",
				Port: "8080",
			},
		},
	}

	cfg, err := InitConfig()
	require.NoError(t, err)
	require.Equal(t, want, cfg)
}

func TestInitConfig_InjectEnvsErr_RootPath_Integration(t *testing.T) {
	t.Setenv("ROOT_PATH", "")
	t.Setenv("CONFIG_PATH", "config.yaml")

	_, err := InitConfig()

	require.Error(t, err)
	require.ErrorContains(t, err, "InjectEnvs failed")
	require.ErrorContains(t, err, "ROOT_PATH environment variable is required")
}

func TestInitConfig_InjectEnvsErr_ConfigPath_Integration(t *testing.T) {
	t.Setenv("ROOT_PATH", "some_path")
	t.Setenv("CONFIG_PATH", "")

	_, err := InitConfig()

	require.Error(t, err)
	require.ErrorContains(t, err, "InjectEnvs failed")
	require.ErrorContains(t, err, "CONFIG_PATH environment variable is required")
}

func TestInitConfig_ConfigFileNotFound_Integration(t *testing.T) {
	t.Setenv("ROOT_PATH", "some_path")
	t.Setenv("CONFIG_PATH", "config.yaml")

	wantErrContain := "cannot read config"

	_, err := InitConfig()
	require.Error(t, err)
	require.ErrorContains(t, err, wantErrContain)
}
