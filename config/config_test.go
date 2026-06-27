package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Test helper types
type (
	serverConfig struct {
		Server struct {
			Host string `validate:"required"`
			Port int    `validate:"required,min=1,max=65535"`
		} `validate:"required"`
	}

	fullConfig struct {
		Server struct {
			Host string `validate:"required"`
			Port int    `validate:"required"`
		} `validate:"required"`
		Postgres struct {
			Host     string `validate:"required"`
			Port     int    `validate:"required"`
			Database string `validate:"required"`
		} `validate:"required"`
	}

	postgresConfig struct {
		Postgres struct {
			Host string `validate:"required"`
			Port int    `validate:"required"`
		} `validate:"required"`
	}

	testConfig struct {
		Test struct {
			Value string `validate:"required"`
		} `validate:"required"`
	}

	envConfig struct {
		ServerHost string `mapstructure:"server_host" validate:"required"`
		ServerPort int    `mapstructure:"server_port" validate:"required,min=1,max=65535"`
	}

	prefixConfig struct {
		Host string `mapstructure:"host" validate:"required"`
		Port int    `mapstructure:"port" validate:"required,min=1,max=65535"`
	}

	configWithDB struct {
		Server struct {
			Host string `validate:"required"`
			Port int    `validate:"required,min=1,max=65535"`
		} `validate:"required"`
		Database struct {
			Host string `validate:"required"`
			Port int    `validate:"required,min=1,max=65535"`
		} `validate:"required"`
	}
)

// Test helpers

func createTempFile(t *testing.T, filename, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, filename)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func setEnvVars(t *testing.T, vars map[string]string) func() {
	t.Helper()
	for k, v := range vars {
		_ = os.Setenv(k, v)
	}
	return func() {
		for k := range vars {
			_ = os.Unsetenv(k)
		}
	}
}

func TestLoad(t *testing.T) {
	yamlContent := `server:
  host: "0.0.0.0"
  port: 8080
`
	yamlPath := createTempFile(t, "config.yaml", yamlContent)

	cfg, err := Load[serverConfig](yamlPath)
	require.NoError(t, err)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
}

func TestLoad_ENVOverride(t *testing.T) {
	tests := []struct {
		name         string
		envVars      map[string]string
		yaml         string
		loadOptions  []Option
		expectedPort int
		wantErr      bool
		configType   string // "server" or "postgres"
	}{
		{
			name: "ENV override without prefix",
			envVars: map[string]string{
				"SERVER_PORT": "9090",
			},
			yaml: `server:
  port: 8080
`,
			loadOptions:  []Option{},
			expectedPort: 9090,
			wantErr:      false,
			configType:   "server",
		},
		{
			name: "ENV override with prefix",
			envVars: map[string]string{
				"APP_SERVER_PORT": "9090",
			},
			yaml: `server:
  port: 8080
`,
			loadOptions:  []Option{WithENVPrefix("APP")},
			expectedPort: 9090,
			wantErr:      false,
			configType:   "server",
		},
		{
			name: "ENV override nested config",
			envVars: map[string]string{
				"POSTGRES_HOST": "prod.db.local",
				"POSTGRES_PORT": "5433",
			},
			yaml: `postgres:
  host: "localhost"
  port: 5432
`,
			loadOptions:  []Option{},
			expectedPort: 5433,
			wantErr:      false,
			configType:   "postgres",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setEnvVars(t, tt.envVars)
			defer cleanup()

			yamlPath := createTempFile(t, "config.yaml", tt.yaml)

			switch tt.configType {
			case "postgres":
				cfg, err := LoadWithOptions[postgresConfig](yamlPath, tt.loadOptions...)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, tt.envVars["POSTGRES_HOST"], cfg.Postgres.Host, "ENV should override nested host")
				assert.Equal(t, tt.expectedPort, cfg.Postgres.Port, "ENV should override nested port")
			default: // server
				type serverPortConfig struct {
					Server struct {
						Port int `validate:"required,min=1,max=65535"`
					} `validate:"required"`
				}

				cfg, err := LoadWithOptions[serverPortConfig](yamlPath, tt.loadOptions...)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, tt.expectedPort, cfg.Server.Port)
			}
		})
	}
}

func TestLoad_Validation(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing required field - port",
			yaml: `server:
  host: "localhost"
  # port is missing - required field
`,
			wantErr: true,
			errMsg:  "validate",
		},
		{
			name: "port above maximum value",
			yaml: `server:
  port: 99999
`,
			wantErr: true,
			errMsg:  "validate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yamlPath := createTempFile(t, "config.yaml", tt.yaml)

			_, err := Load[serverConfig](yamlPath)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestLoad_Errors(t *testing.T) {
	tests := []struct {
		name        string
		yamlPath    string
		yamlContent string
		wantErr     bool
		expectedErr string
	}{
		{
			name:        "file not found",
			yamlPath:    "/nonexistent/config.yaml",
			yamlContent: "",
			wantErr:     true,
			expectedErr: "read config",
		},
		{
			name: "invalid YAML syntax",
			yamlContent: `server:
  port: [invalid
`,
			wantErr:     true,
			expectedErr: "read config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type Config struct {
				Port int `validate:"required"`
			}

			yamlPath := tt.yamlPath
			if tt.yamlContent != "" {
				yamlPath = createTempFile(t, "config.yaml", tt.yamlContent)
			}

			_, err := Load[Config](yamlPath)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != "" {
					assert.Contains(t, err.Error(), tt.expectedErr)
				}
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestLoadWithLogger(t *testing.T) {
	yamlContent := `test:
  value: "debug"`
	yamlPath := createTempFile(t, "config.yaml", yamlContent)

	logger := zap.NewNop()
	cfg, err := LoadWithOptions[testConfig](yamlPath, WithLogger(logger))
	require.NoError(t, err)
	assert.Equal(t, "debug", cfg.Test.Value)
}

func TestLoadNestedConfig(t *testing.T) {
	yamlContent := `server:
  host: "localhost"
  port: 8080
postgres:
  host: "db.local"
  port: 5432
  database: "mydb"
`
	yamlPath := createTempFile(t, "config.yaml", yamlContent)

	cfg, err := Load[fullConfig](yamlPath)
	require.NoError(t, err)
	assert.Equal(t, "localhost", cfg.Server.Host)
	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "db.local", cfg.Postgres.Host)
	assert.Equal(t, 5432, cfg.Postgres.Port)
	assert.Equal(t, "mydb", cfg.Postgres.Database)
}

func TestLoadRace(t *testing.T) {
	yamlContent := `test:
  value: "debug"`
	yamlPath := createTempFile(t, "config.yaml", yamlContent)

	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = Load[testConfig](yamlPath)
		}
		close(done)
	}()

	for i := 0; i < 100; i++ {
		cfg, err := Load[testConfig](yamlPath)
		require.NoError(t, err)
		assert.Equal(t, "debug", cfg.Test.Value)
	}

	<-done
}

func ExampleLoad() {
	// Define your config structure
	type Config struct {
		Server struct {
			Host string `validate:"required"`
			Port int    `validate:"min=1,max=65535"`
		} `validate:"required"`
	}

	// Load from YAML file (with ENV override)
	cfg, err := Load[Config]("./config.yaml")
	if err != nil {
		panic(err)
	}

	// Use the config
	fmt.Println(cfg.Server.Host, cfg.Server.Port)
}

func TestLoadFromEnv(t *testing.T) {
	tests := []struct {
		name         string
		envContent   string
		loadOptions  []Option
		wantErr      bool
		errMsg       string
		expectedHost string
		expectedPort int
		configType   string // "env" or "prefix"
	}{
		{
			name: "successful load from .env",
			envContent: `SERVER_HOST=localhost
SERVER_PORT=8080
`,
			loadOptions:  []Option{},
			wantErr:      false,
			expectedHost: "localhost",
			expectedPort: 8080,
			configType:   "env",
		},
		{
			name: "validation fails for port above maximum",
			envContent: `SERVER_HOST=localhost
SERVER_PORT=99999
`,
			loadOptions: []Option{},
			wantErr:     true,
			errMsg:      "validate",
			configType:  "env",
		},
		{
			name: "load from .env with prefix",
			envContent: `HOST=localhost
PORT=8080
`,
			loadOptions:  []Option{WithENVPrefix("SERVER")},
			wantErr:      false,
			expectedHost: "localhost",
			expectedPort: 8080,
			configType:   "prefix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envPath := createTempFile(t, ".env", tt.envContent)

			switch tt.configType {
			case "prefix":
				cfg, err := LoadFromEnvWithOptions[prefixConfig](envPath, tt.loadOptions...)

				if tt.wantErr {
					assert.Error(t, err)
					if tt.errMsg != "" {
						assert.Contains(t, err.Error(), tt.errMsg)
					}
					return
				}

				require.NoError(t, err)
				assert.Equal(t, tt.expectedHost, cfg.Host)
				assert.Equal(t, tt.expectedPort, cfg.Port)
			default: // env
				cfg, err := LoadFromEnvWithOptions[envConfig](envPath, tt.loadOptions...)

				if tt.wantErr {
					assert.Error(t, err)
					if tt.errMsg != "" {
						assert.Contains(t, err.Error(), tt.errMsg)
					}
					return
				}

				require.NoError(t, err)
				assert.Equal(t, tt.expectedHost, cfg.ServerHost)
				assert.Equal(t, tt.expectedPort, cfg.ServerPort)
			}
		})
	}
}

func TestLoadWithEnvFile(t *testing.T) {
	envContent := `DATABASE_HOST=localhost
DATABASE_PORT=5432
`
	envPath := createTempFile(t, ".env", envContent)

	yamlContent := `server:
  host: "0.0.0.0"
  port: 8080
database:
  host: "default-db"
  port: 3306
`
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "config.yaml")
	require.NoError(t, os.WriteFile(yamlPath, []byte(yamlContent), 0o600))

	cfg, err := LoadWithOptions[configWithDB](yamlPath, WithEnvFile(envPath))
	require.NoError(t, err)
	assert.Equal(t, "0.0.0.0", cfg.Server.Host, "server config from YAML")
	assert.Equal(t, 8080, cfg.Server.Port, "server port from YAML")
	assert.Equal(t, "localhost", cfg.Database.Host, "database host from .env")
	assert.Equal(t, 5432, cfg.Database.Port, "database port from .env")
}
