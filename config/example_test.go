package config_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/haipham22/govern/config"
)

// Example_basic demonstrates basic usage of the config package.
func Example_basic() {
	// Create a temporary config file for this example
	tmpDir := os.TempDir()
	yamlPath := filepath.Join(tmpDir, "config-example.yaml")
	yamlContent := `server:
  host: "0.0.0.0"
  port: 8080

postgres:
  host: "localhost"
  port: 5432
  database: "myapp"
  user: "dbuser"
`
	_ = os.WriteFile(yamlPath, []byte(yamlContent), 0o600)
	defer func() { _ = os.Remove(yamlPath) }()

	// Define your config structure
	type Config struct {
		Server struct {
			Host string `validate:"required"`
			Port int    `validate:"min=1,max=65535"`
		} `validate:"required"`
		Postgres struct {
			Host     string `validate:"required"`
			Port     int    `validate:"min=1,max=65535"`
			Database string `validate:"required"`
			User     string `validate:"required"`
		} `validate:"required"`
	}

	// Load config from YAML file
	cfg, err := config.Load[Config](yamlPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Use the config directly
	fmt.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Postgres: %s@%s:%d/%s",
		cfg.Postgres.User,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
	)

	// Output:
	// Server: 0.0.0.0:8080
	// Postgres: dbuser@localhost:5432/myapp
}

// Example_envOverride demonstrates ENV variable override.
func Example_envOverride() {
	// Set ENV variables to override YAML values
	_ = os.Setenv("SERVER_PORT", "9090")
	_ = os.Setenv("POSTGRES_HOST", "prod.db.local")
	defer func() { _ = os.Unsetenv("SERVER_PORT") }()
	defer func() { _ = os.Unsetenv("POSTGRES_HOST") }()

	// Create a temporary config file
	tmpDir := os.TempDir()
	yamlPath := filepath.Join(tmpDir, "config-env.yaml")
	yamlContent := `server:
  host: "0.0.0.0"
  port: 8080

postgres:
  host: "localhost"
  port: 5432
`
	_ = os.WriteFile(yamlPath, []byte(yamlContent), 0o600)
	defer func() { _ = os.Remove(yamlPath) }()

	type Config struct {
		Server struct {
			Host string `validate:"required"`
			Port int    `validate:"required"`
		} `validate:"required"`
		Postgres struct {
			Host string `validate:"required"`
			Port int    `validate:"required"`
		} `validate:"required"`
	}

	// Load config - ENV vars will override YAML values
	cfg, err := config.Load[Config](yamlPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// SERVER_PORT and POSTGRES_HOST override the YAML values
	fmt.Printf("Server Port: %d (overridden by ENV)\n", cfg.Server.Port)
	fmt.Printf("Postgres Host: %s (overridden by ENV)\n", cfg.Postgres.Host)

	// Output:
	// Server Port: 9090 (overridden by ENV)
	// Postgres Host: prod.db.local (overridden by ENV)
}

// Example_withENVPrefix demonstrates using ENV prefix.
func Example_withENVPrefix() {
	// Set ENV variables with prefix
	_ = os.Setenv("APP_SERVER_PORT", "9090")
	defer func() { _ = os.Unsetenv("APP_SERVER_PORT") }()

	// Create a temporary config file
	tmpDir := os.TempDir()
	yamlPath := filepath.Join(tmpDir, "config-prefix.yaml")
	yamlContent := `server:
  port: 8080
`
	_ = os.WriteFile(yamlPath, []byte(yamlContent), 0o600)
	defer func() { _ = os.Remove(yamlPath) }()

	type Config struct {
		Server struct {
			Port int `validate:"required"`
		} `validate:"required"`
	}

	// Load config with ENV prefix
	cfg, err := config.LoadWithOptions[Config](
		yamlPath,
		config.WithENVPrefix("APP"),
	)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// APP_SERVER_PORT overrides server.port
	fmt.Printf("Port: %d\n", cfg.Server.Port)

	// Output:
	// Port: 9090
}

// Example_validation demonstrates struct validation.
func Example_validation() {
	// Create a config file with invalid port
	tmpDir := os.TempDir()
	yamlPath := filepath.Join(tmpDir, "config-validation.yaml")
	yamlContent := `server:
  port: 99999
`
	_ = os.WriteFile(yamlPath, []byte(yamlContent), 0o600)
	defer func() { _ = os.Remove(yamlPath) }()

	type Config struct {
		Server struct {
			Port int `validate:"required,min=1,max=65535"`
		} `validate:"required"`
	}

	// Validation will fail because port > 65535
	_, err := config.Load[Config](yamlPath)
	if err != nil {
		// Print just the error prefix, not the full message
		fmt.Println("Validation error: validate")
	}

	// Output: Validation error: validate
}

// Example_envOnly demonstrates loading config from .env file only (no YAML).
func Example_envOnly() {
	// Create a temporary .env file for this example
	tmpDir := os.TempDir()
	envPath := filepath.Join(tmpDir, ".env-example")
	envContent := `SERVER_HOST=0.0.0.0
SERVER_PORT=8080
DATABASE_HOST=localhost
DATABASE_PORT=5432
`
	_ = os.WriteFile(envPath, []byte(envContent), 0o600)
	defer func() { _ = os.Remove(envPath) }()

	// Define your config structure
	// Note: Use mapstructure tags for .env files
	type Config struct {
		ServerHost   string `mapstructure:"server_host" validate:"required"`
		ServerPort   int    `mapstructure:"server_port" validate:"required,min=1,max=65535"`
		DatabaseHost string `mapstructure:"database_host" validate:"required"`
		DatabasePort int    `mapstructure:"database_port" validate:"required,min=1,max=65535"`
	}

	// Load config from .env file
	cfg, err := config.LoadFromEnv[Config](envPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Use the config
	fmt.Println(cfg.ServerHost, cfg.ServerPort)
	// Output: 0.0.0.0 8080
}

// Example_envFile demonstrates loading YAML config with .env file override.
func Example_envFile() {
	// Create a temporary .env file
	tmpDir := os.TempDir()
	envPath := filepath.Join(tmpDir, ".env-override")
	envContent := `DATABASE_HOST=production-db
DATABASE_PORT=5432
`
	_ = os.WriteFile(envPath, []byte(envContent), 0o600)
	defer func() { _ = os.Remove(envPath) }()

	// Create a temporary YAML config file
	yamlPath := filepath.Join(tmpDir, "config-override.yaml")
	yamlContent := `server:
  host: "localhost"
  port: 8080

database:
  host: "localhost"
  port: 3306
`
	_ = os.WriteFile(yamlPath, []byte(yamlContent), 0o600)
	defer func() { _ = os.Remove(yamlPath) }()

	// Define your config structure
	type Config struct {
		Server struct {
			Host string `validate:"required"`
			Port int    `validate:"required,min=1,max=65535"`
		} `validate:"required"`
		Database struct {
			Host string `validate:"required"`
			Port int    `validate:"required,min=1,max=65535"`
		} `validate:"required"`
	}

	// Load YAML config with .env file override
	cfg, err := config.LoadWithOptions[Config](yamlPath, config.WithEnvFile(envPath))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// .env values override YAML values
	fmt.Println(cfg.Database.Host)
	// Output: production-db
}
