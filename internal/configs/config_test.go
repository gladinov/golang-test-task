//go:build unit

package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPostgresHost_GetStringHost(t *testing.T) {
	cases := []struct {
		name       string
		cfg        PostgresHost
		wantErr    bool
		wantSubstr string
	}{
		{
			name: "all fields filled",
			cfg: PostgresHost{
				Host: "localhost", User: "user", Password: "pass", Dbname: "db", Port: "5432", SslMode: "disable",
			},
			wantErr:    false,
			wantSubstr: "host=localhost user=user password=pass dbname=db port=5432 sslmode=disable",
		},
		{
			name:       "empty host",
			cfg:        PostgresHost{User: "user", Password: "pass", Dbname: "db", Port: "5432"},
			wantErr:    true,
			wantSubstr: "empty host",
		},
		{
			name:       "empty user",
			cfg:        PostgresHost{Host: "localhost", Password: "pass", Dbname: "db", Port: "5432"},
			wantErr:    true,
			wantSubstr: "empty user",
		},
		{
			name: "empty dbname",
			cfg: PostgresHost{
				Host: "localhost", User: "user", Password: "pass", Port: "5432", SslMode: "disable",
			},
			wantErr:    true,
			wantSubstr: "empty dbname",
		},
		{
			name: "empty password",
			cfg: PostgresHost{
				Host: "localhost", User: "user", Dbname: "db", Port: "5432", SslMode: "disable",
			},
			wantErr:    true,
			wantSubstr: "empty password",
		},
		{
			name: "empty port",
			cfg: PostgresHost{
				Host: "localhost", User: "user", Password: "pass", Dbname: "db", SslMode: "disable",
			},
			wantErr:    true,
			wantSubstr: "empty port",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.cfg.GetStringHost()
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.wantSubstr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantSubstr, got)
			}
		})
	}
}

func TestInjectEnvs(t *testing.T) {
	cases := []struct {
		name        string
		rootPath    string
		configPath  string
		wantErr     bool
		errContains string
		wantRoot    string
		wantConfig  string
	}{
		{
			name:       "both envs set",
			rootPath:   "/root",
			configPath: "config.yml",
			wantErr:    false,
			wantRoot:   "/root",
			wantConfig: "config.yml",
		},
		{
			name:        "ROOT_PATH missing",
			rootPath:    "",
			configPath:  "config.yml",
			wantErr:     true,
			errContains: "ROOT_PATH",
		},
		{
			name:        "CONFIG_PATH missing",
			rootPath:    "/root",
			configPath:  "",
			wantErr:     true,
			errContains: "CONFIG_PATH",
		},
		{
			name:        "both missing",
			rootPath:    "",
			configPath:  "",
			wantErr:     true,
			errContains: "ROOT_PATH",
		},
	}

	for _, tc := range cases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ROOT_PATH", tc.rootPath)
			t.Setenv("CONFIG_PATH", tc.configPath)

			envs, err := InjectEnvs()
			if tc.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.wantRoot, envs.RootPath)
				require.Equal(t, tc.wantConfig, envs.ConfigPath)
			}
		})
	}
}
