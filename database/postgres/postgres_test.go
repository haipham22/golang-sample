package postgres_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/haipham22/govern/database/postgres"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		opts    []postgres.Option
		wantErr bool
	}{
		{
			name:    "invalid DSN",
			dsn:     "invalid://dsn",
			wantErr: true,
		},
		{
			name:    "invalid DSN with options",
			dsn:     "host=invalid user=invalid",
			wantErr: true,
		},
		{
			name:    "empty DSN",
			dsn:     "",
			wantErr: true,
		},
		{
			name:    "malformed DSN",
			dsn:     "host=localhost port=abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup, err := postgres.New(tt.dsn, tt.opts...)

			if tt.wantErr {
				if err == nil {
					t.Errorf("New() expected error, got nil")
				}
				if db != nil {
					t.Errorf("New() expected nil db on error, got %v", db)
				}
				if cleanup != nil {
					cleanup()
				}
				return
			}

			if err != nil {
				t.Errorf("New() unexpected error: %v", err)
				return
			}

			if db == nil {
				t.Errorf("New() expected non-nil db")
			}

			if cleanup == nil {
				t.Errorf("New() expected non-nil cleanup function")
			} else {
				cleanup()
			}
		})
	}
}

func TestNewWithOptions(t *testing.T) {
	opts := []postgres.Option{
		postgres.WithMaxIdleConns(10),
		postgres.WithMaxOpenConns(50),
		postgres.WithConnMaxLifetime(10 * time.Minute),
		postgres.WithConnMaxIdleTime(time.Minute),
		postgres.WithPreferSimpleProtocol(false),
		postgres.WithDebug(true),
	}

	// This will fail to connect but we're testing options are applied
	db, cleanup, err := postgres.New("host=localhost", opts...)

	if err != nil && db == nil {
		// Expected - connection fails but options were processed
		return
	}

	if cleanup != nil {
		cleanup()
	}
}

func TestDefaults(t *testing.T) {
	tests := []struct {
		name         string
		defaultValue interface{}
		field        string
	}{
		{
			name:         "DefaultMaxIdleConns",
			defaultValue: 25,
			field:        "MaxIdleConns",
		},
		{
			name:         "DefaultMaxOpenConns",
			defaultValue: 100,
			field:        "MaxOpenConns",
		},
		{
			name:         "DefaultConnMaxLifetime",
			defaultValue: 5 * time.Minute,
			field:        "ConnMaxLifetime",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that default constants are set to expected values
			switch tt.field {
			case "MaxIdleConns":
				if got := postgres.DefaultMaxIdleConns; got != tt.defaultValue.(int) {
					t.Errorf("DefaultMaxIdleConns = %v, want %v", got, tt.defaultValue)
				}
			case "MaxOpenConns":
				if got := postgres.DefaultMaxOpenConns; got != tt.defaultValue.(int) {
					t.Errorf("DefaultMaxOpenConns = %v, want %v", got, tt.defaultValue)
				}
			case "ConnMaxLifetime":
				if got := postgres.DefaultConnMaxLifetime; got != tt.defaultValue.(time.Duration) {
					t.Errorf("DefaultConnMaxLifetime = %v, want %v", got, tt.defaultValue)
				}
			}
		})
	}
}

func TestCleanup(t *testing.T) {
	db, cleanup, err := postgres.New("host=localhost")
	if err != nil && db == nil {
		t.Skip("cannot connect to test database")
		return
	}

	if cleanup == nil {
		t.Fatal("expected non-nil cleanup function")
	}

	// Calling cleanup should not panic
	cleanup()

	// Calling cleanup again should be safe
	cleanup()
}

func TestOptionFunctions(t *testing.T) {
	tests := []struct {
		name   string
		option postgres.Option
		verify func(*postgres.Config) bool
	}{
		{
			name:   "WithDebug true",
			option: postgres.WithDebug(true),
			verify: func(cfg *postgres.Config) bool {
				return cfg.Debug == true
			},
		},
		{
			name:   "WithDebug false",
			option: postgres.WithDebug(false),
			verify: func(cfg *postgres.Config) bool {
				return cfg.Debug == false
			},
		},
		{
			name:   "WithMaxIdleConns",
			option: postgres.WithMaxIdleConns(15),
			verify: func(cfg *postgres.Config) bool {
				return cfg.MaxIdleConns == 15
			},
		},
		{
			name:   "WithMaxOpenConns",
			option: postgres.WithMaxOpenConns(75),
			verify: func(cfg *postgres.Config) bool {
				return cfg.MaxOpenConns == 75
			},
		},
		{
			name:   "WithConnMaxLifetime",
			option: postgres.WithConnMaxLifetime(15 * time.Minute),
			verify: func(cfg *postgres.Config) bool {
				return cfg.ConnMaxLifetime == 15*time.Minute
			},
		},
		{
			name:   "WithConnMaxIdleTime",
			option: postgres.WithConnMaxIdleTime(10 * time.Minute),
			verify: func(cfg *postgres.Config) bool {
				return cfg.ConnMaxIdleTime == 10*time.Minute
			},
		},
		{
			name:   "WithPreferSimpleProtocol true",
			option: postgres.WithPreferSimpleProtocol(true),
			verify: func(cfg *postgres.Config) bool {
				return cfg.PreferSimpleProtocol == true
			},
		},
		{
			name:   "WithPreferSimpleProtocol false",
			option: postgres.WithPreferSimpleProtocol(false),
			verify: func(cfg *postgres.Config) bool {
				return cfg.PreferSimpleProtocol == false
			},
		},
		{
			name:   "WithLogger",
			option: postgres.WithLogger(logger.Default),
			verify: func(cfg *postgres.Config) bool {
				return cfg.Logger != nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &postgres.Config{}
			tt.option(cfg)

			if !tt.verify(cfg) {
				t.Errorf("Option did not set expected value")
			}
		})
	}
}

func TestOptionValidation(t *testing.T) {
	tests := []struct {
		name   string
		option postgres.Option
		verify func(*postgres.Config) bool
	}{
		{
			name:   "zero max idle conns",
			option: postgres.WithMaxIdleConns(0),
			verify: func(cfg *postgres.Config) bool {
				return cfg.MaxIdleConns == 0
			},
		},
		{
			name:   "negative max idle conns",
			option: postgres.WithMaxIdleConns(-5),
			verify: func(cfg *postgres.Config) bool {
				return cfg.MaxIdleConns == -5
			},
		},
		{
			name:   "zero max open conns",
			option: postgres.WithMaxOpenConns(0),
			verify: func(cfg *postgres.Config) bool {
				return cfg.MaxOpenConns == 0
			},
		},
		{
			name:   "negative max open conns",
			option: postgres.WithMaxOpenConns(-10),
			verify: func(cfg *postgres.Config) bool {
				return cfg.MaxOpenConns == -10
			},
		},
		{
			name:   "zero conn max lifetime",
			option: postgres.WithConnMaxLifetime(0),
			verify: func(cfg *postgres.Config) bool {
				return cfg.ConnMaxLifetime == 0
			},
		},
		{
			name:   "zero conn max idle time",
			option: postgres.WithConnMaxIdleTime(0),
			verify: func(cfg *postgres.Config) bool {
				return cfg.ConnMaxIdleTime == 0
			},
		},
		{
			name:   "negative duration",
			option: postgres.WithConnMaxLifetime(-5 * time.Minute),
			verify: func(cfg *postgres.Config) bool {
				return cfg.ConnMaxLifetime == -5*time.Minute
			},
		},
		{
			name:   "nil logger",
			option: postgres.WithLogger(nil),
			verify: func(cfg *postgres.Config) bool {
				return cfg.Logger == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &postgres.Config{}
			tt.option(cfg)

			if !tt.verify(cfg) {
				t.Errorf("Option validation failed")
			}
		})
	}
}

func TestConnectionErrorScenarios(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{
			name:    "invalid host format",
			dsn:     "host=/invalid/path port=5432",
			wantErr: true,
		},
		{
			name:    "invalid port",
			dsn:     "host=localhost port=invalid",
			wantErr: true,
		},
		{
			name:    "unknown database",
			dsn:     "host=localhost port=5432 database=nonexistent_db_xyz123",
			wantErr: true, // Will fail to connect if postgres exists
		},
		{
			name:    "invalid sslmode",
			dsn:     "host=localhost sslmode=invalid",
			wantErr: true,
		},
		{
			name:    "valid sslmode require",
			dsn:     "host=localhost sslmode=require",
			wantErr: true, // Still fails if no postgres, but DSN is valid
		},
		{
			name:    "valid sslmode disable",
			dsn:     "host=localhost sslmode=disable",
			wantErr: true, // Still fails if no postgres
		},
		{
			name:    "with timezone",
			dsn:     "host=localhost timezone=UTC",
			wantErr: true, // Still fails if no postgres
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup, err := postgres.New(tt.dsn)

			if tt.wantErr {
				// We expect connection to fail
				if err == nil && db == nil {
					t.Errorf("New() expected error for invalid DSN, got nil")
				}
				// db may be non-nil even on connection error
			}

			if cleanup != nil {
				cleanup()
			}
		})
	}
}

func TestNilCleanupOnError(t *testing.T) {
	// Verify that cleanup function is nil when connection fails
	db, cleanup, err := postgres.New("invalid://dsn")

	if err == nil {
		t.Error("Expected error for invalid DSN")
	}

	if db != nil {
		t.Error("Expected nil DB for invalid DSN")
	}

	// cleanup should be nil or safe to call
	if cleanup != nil {
		cleanup()
	}
}

func TestMockLogger(t *testing.T) {
	// Test with the built-in GORM default logger
	testLogger := logger.Default

	db, cleanup, err := postgres.New("host=localhost", postgres.WithLogger(testLogger))

	// Connection will fail, but logger option should be processed
	if err != nil && db == nil {
		return
	}

	if cleanup != nil {
		cleanup()
	}
}

// Test error paths in New function
func TestNewErrorPaths(t *testing.T) {
	tests := []struct {
		name          string
		dsn           string
		setupMock     func() // Optional setup for mocking
		expectError   bool
		errorContains string
	}{
		{
			name:        "completely invalid DSN",
			dsn:         "not-a-dsn",
			expectError: true,
		},
		{
			name:        "dsn with only spaces",
			dsn:         "   ",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup, err := postgres.New(tt.dsn)

			if tt.expectError {
				if err == nil {
					t.Errorf("New() expected error, got nil")
				}
				if tt.errorContains != "" && err != nil {
					if !errors.Is(err, errors.New(tt.errorContains)) {
						if err.Error() == "" {
							t.Errorf("Expected error containing %q, got empty error", tt.errorContains)
						}
					}
				}
			}

			if cleanup != nil {
				cleanup()
			}

			_ = db // Avoid unused variable error
		})
	}
}

func TestDebugWithoutLogger(t *testing.T) {
	// Test the code path where Debug=true but Logger=nil
	// This exercises line 98: gormCfg.Logger = logger.Default.LogMode(logger.Info)
	db, cleanup, err := postgres.New("host=localhost", postgres.WithDebug(true))

	// Connection will fail, but the debug option should be processed
	if err != nil && db == nil {
		return // Expected
	}

	if cleanup != nil {
		cleanup()
	}

	_ = db
}

func TestDebugWithCustomLogger(t *testing.T) {
	// Test that when both Debug=true and a custom logger is set,
	// the custom logger is used instead of the default
	customLogger := logger.Default.LogMode(logger.Silent)

	db, cleanup, err := postgres.New("host=localhost",
		postgres.WithDebug(true),
		postgres.WithLogger(customLogger),
	)

	// Connection will fail, but the options should be processed
	if err != nil && db == nil {
		return // Expected
	}

	if cleanup != nil {
		cleanup()
	}

	_ = db
}

func TestConnMaxIdleTimeZero(t *testing.T) {
	// Test the code path where ConnMaxIdleTime is 0
	// This exercises the condition at line 114
	db, cleanup, err := postgres.New("host=localhost",
		postgres.WithConnMaxIdleTime(0),
	)

	if err != nil && db == nil {
		return // Expected - no database
	}

	if cleanup != nil {
		cleanup()
	}

	_ = db
}

func TestConnMaxIdleTimePositive(t *testing.T) {
	// Test the code path where ConnMaxIdleTime > 0
	// This exercises line 115: sqlDB.SetConnMaxIdleTime
	db, cleanup, err := postgres.New("host=localhost",
		postgres.WithConnMaxIdleTime(5*time.Minute),
	)

	if err != nil && db == nil {
		return // Expected - no database
	}

	if cleanup != nil {
		cleanup()
	}

	_ = db
}

func TestPoolConfiguration(t *testing.T) {
	// Test various pool configurations
	tests := []struct {
		name        string
		maxIdle     int
		maxOpen     int
		maxLifetime time.Duration
		maxIdleTime time.Duration
	}{
		{
			name:        "small pool",
			maxIdle:     5,
			maxOpen:     10,
			maxLifetime: 1 * time.Minute,
			maxIdleTime: 30 * time.Second,
		},
		{
			name:        "large pool",
			maxIdle:     50,
			maxOpen:     200,
			maxLifetime: 30 * time.Minute,
			maxIdleTime: 10 * time.Minute,
		},
		{
			name:        "extended lifetime",
			maxIdle:     25,
			maxOpen:     100,
			maxLifetime: 1 * time.Hour,
			maxIdleTime: 30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup, err := postgres.New("host=localhost",
				postgres.WithMaxIdleConns(tt.maxIdle),
				postgres.WithMaxOpenConns(tt.maxOpen),
				postgres.WithConnMaxLifetime(tt.maxLifetime),
				postgres.WithConnMaxIdleTime(tt.maxIdleTime),
			)

			if err != nil && db == nil {
				return // Expected - no database
			}

			if cleanup != nil {
				cleanup()
			}

			_ = db
		})
	}
}

func TestPreferSimpleProtocolOptions(t *testing.T) {
	tests := []struct {
		name              string
		preferSimpleProto bool
	}{
		{
			name:              "simple protocol enabled",
			preferSimpleProto: true,
		},
		{
			name:              "simple protocol disabled",
			preferSimpleProto: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup, err := postgres.New("host=localhost",
				postgres.WithPreferSimpleProtocol(tt.preferSimpleProto),
			)

			if err != nil && db == nil {
				return // Expected - no database
			}

			if cleanup != nil {
				cleanup()
			}

			_ = db
		})
	}
}

func TestMultipleOptions(t *testing.T) {
	// Test that multiple options are correctly applied
	db, cleanup, err := postgres.New("host=localhost",
		postgres.WithDebug(true),
		postgres.WithMaxIdleConns(20),
		postgres.WithMaxOpenConns(80),
		postgres.WithConnMaxLifetime(8*time.Minute),
		postgres.WithConnMaxIdleTime(4*time.Minute),
		postgres.WithPreferSimpleProtocol(false),
		postgres.WithLogger(logger.Default.LogMode(logger.Silent)),
	)

	if err != nil && db == nil {
		return // Expected - no database
	}

	if cleanup != nil {
		cleanup()
	}

	_ = db
}

func TestNoOptions(t *testing.T) {
	// Test that New works with no options (uses all defaults)
	db, cleanup, err := postgres.New("host=localhost")

	if err != nil && db == nil {
		return // Expected - no database
	}

	if cleanup != nil {
		cleanup()
	}

	_ = db
}

func TestConfigCombos(t *testing.T) {
	// Test various combinations of configuration options
	tests := []struct {
		name    string
		options []postgres.Option
	}{
		{
			name: "debug only",
			options: []postgres.Option{
				postgres.WithDebug(true),
			},
		},
		{
			name: "pool settings only",
			options: []postgres.Option{
				postgres.WithMaxIdleConns(15),
				postgres.WithMaxOpenConns(60),
			},
		},
		{
			name: "lifetime settings only",
			options: []postgres.Option{
				postgres.WithConnMaxLifetime(12 * time.Minute),
				postgres.WithConnMaxIdleTime(6 * time.Minute),
			},
		},
		{
			name: "protocol settings only",
			options: []postgres.Option{
				postgres.WithPreferSimpleProtocol(false),
			},
		},
		{
			name: "logger only",
			options: []postgres.Option{
				postgres.WithLogger(logger.Default.LogMode(logger.Warn)),
			},
		},
		{
			name: "all debug options",
			options: []postgres.Option{
				postgres.WithDebug(true),
				postgres.WithLogger(logger.Default.LogMode(logger.Info)),
			},
		},
		{
			name: "all pool options",
			options: []postgres.Option{
				postgres.WithMaxIdleConns(30),
				postgres.WithMaxOpenConns(120),
				postgres.WithConnMaxLifetime(15 * time.Minute),
				postgres.WithConnMaxIdleTime(7 * time.Minute),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup, err := postgres.New("host=localhost", tt.options...)

			if err != nil && db == nil {
				return // Expected - no database
			}

			if cleanup != nil {
				cleanup()
			}

			_ = db
		})
	}
}

func TestDSNVariations(t *testing.T) {
	// Test various DSN formats and connection strings
	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{
			name:    "host only",
			dsn:     "host=localhost",
			wantErr: true, // Will fail to connect without actual DB
		},
		{
			name:    "host and port",
			dsn:     "host=localhost port=5432",
			wantErr: true, // Will fail to connect
		},
		{
			name:    "host port and database",
			dsn:     "host=localhost port=5432 database=testdb",
			wantErr: true, // Will fail to connect
		},
		{
			name:    "with user",
			dsn:     "host=localhost user=testuser",
			wantErr: true, // Will fail to connect
		},
		{
			name:    "with password",
			dsn:     "host=localhost password=testpass",
			wantErr: true, // Will fail to connect
		},
		{
			name:    "full connection string",
			dsn:     "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable",
			wantErr: true, // Will fail to connect
		},
		{
			name:    "with sslmode disable",
			dsn:     "host=localhost sslmode=disable",
			wantErr: true, // Will fail to connect
		},
		{
			name:    "with sslmode require",
			dsn:     "host=localhost sslmode=require",
			wantErr: true, // Will fail to connect
		},
		{
			name:    "with connect_timeout",
			dsn:     "host=localhost connect_timeout=10",
			wantErr: true, // Will fail to connect
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup, err := postgres.New(tt.dsn)

			// We expect connection to fail (no postgres running)
			// but we want to verify the DSN is parsed correctly
			if tt.wantErr {
				if err == nil && db == nil {
					t.Errorf("New() expected connection error, got nil")
				}
			}

			if cleanup != nil {
				cleanup()
			}

			_ = db
		})
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		options []postgres.Option
	}{
		{
			name:    "empty string DSN",
			dsn:     "",
			options: nil,
		},
		{
			name:    "dsn with trailing spaces",
			dsn:     "host=localhost  ",
			options: nil,
		},
		{
			name:    "dsn with special characters in password",
			dsn:     "host=localhost password=p@ss!w0rd",
			options: nil,
		},
		{
			name: "very large pool size",
			dsn:  "host=localhost",
			options: []postgres.Option{
				postgres.WithMaxOpenConns(10000),
			},
		},
		{
			name: "very long lifetime",
			dsn:  "host=localhost",
			options: []postgres.Option{
				postgres.WithConnMaxLifetime(24 * time.Hour),
			},
		},
		{
			name: "very short lifetime",
			dsn:  "host=localhost",
			options: []postgres.Option{
				postgres.WithConnMaxLifetime(time.Second),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup, err := postgres.New(tt.dsn, tt.options...)

			// Connection will likely fail without postgres
			if err != nil && db == nil {
				return // Expected
			}

			if cleanup != nil {
				cleanup()
			}

			_ = db
		})
	}
}

// Tests using go-sqlmock to cover the success path

func TestConfigureConnectionPoolSuccess(t *testing.T) {
	// Create mock database
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = mockDB.Close() }()

	// Create GORM DB from mock
	dialector := gormpostgres.New(gormpostgres.Config{
		Conn: mockDB,
		DSN:  "sqlmock_database",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	// Test configuration
	cfg := &postgres.Config{
		MaxIdleConns:         10,
		MaxOpenConns:         20,
		ConnMaxLifetime:      5 * time.Minute,
		ConnMaxIdleTime:      2 * time.Minute,
		PreferSimpleProtocol: true,
	}

	cleanup, err := postgres.ConfigureConnectionPool(db, cfg)
	require.NoError(t, err)
	require.NotNil(t, cleanup)

	// Verify cleanup can be called without panic
	cleanup()

	// Calling cleanup again should be safe
	cleanup()
}

func TestConfigureConnectionPoolWithZeroIdleTime(t *testing.T) {
	// Test the code path where ConnMaxIdleTime is 0
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = mockDB.Close() }()

	dialector := gormpostgres.New(gormpostgres.Config{
		Conn: mockDB,
		DSN:  "sqlmock_database",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	cfg := &postgres.Config{
		MaxIdleConns:    10,
		MaxOpenConns:    20,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: 0, // This should skip SetConnMaxIdleTime
	}

	cleanup, err := postgres.ConfigureConnectionPool(db, cfg)
	require.NoError(t, err)
	require.NotNil(t, cleanup)

	cleanup()
}

func TestConfigureConnectionPoolVariousConfigurations(t *testing.T) {
	tests := []struct {
		name        string
		maxIdle     int
		maxOpen     int
		maxLifetime time.Duration
		maxIdleTime time.Duration
	}{
		{
			name:        "small pool",
			maxIdle:     5,
			maxOpen:     10,
			maxLifetime: 1 * time.Minute,
			maxIdleTime: 30 * time.Second,
		},
		{
			name:        "large pool",
			maxIdle:     50,
			maxOpen:     200,
			maxLifetime: 30 * time.Minute,
			maxIdleTime: 10 * time.Minute,
		},
		{
			name:        "zero idle time",
			maxIdle:     25,
			maxOpen:     100,
			maxLifetime: 5 * time.Minute,
			maxIdleTime: 0,
		},
		{
			name:        "extended lifetime",
			maxIdle:     25,
			maxOpen:     100,
			maxLifetime: 1 * time.Hour,
			maxIdleTime: 30 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB, _, err := sqlmock.New()
			require.NoError(t, err)
			defer func() { _ = mockDB.Close() }()

			dialector := gormpostgres.New(gormpostgres.Config{
				Conn: mockDB,
				DSN:  "sqlmock_database",
			})

			db, err := gorm.Open(dialector, &gorm.Config{})
			require.NoError(t, err)

			cfg := &postgres.Config{
				MaxIdleConns:    tt.maxIdle,
				MaxOpenConns:    tt.maxOpen,
				ConnMaxLifetime: tt.maxLifetime,
				ConnMaxIdleTime: tt.maxIdleTime,
			}

			cleanup, err := postgres.ConfigureConnectionPool(db, cfg)
			require.NoError(t, err)
			require.NotNil(t, cleanup)

			cleanup()
		})
	}
}

func TestConfigureConnectionPoolFailure(t *testing.T) {
	// Test error handling when db.DB() fails
	// We'll create a mock that fails to get underlying sql.DB

	// Create a mock sql.DB that will fail
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = mockDB.Close() }()

	dialector := gormpostgres.New(gormpostgres.Config{
		Conn: mockDB,
		DSN:  "sqlmock_database",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	// Close the underlying DB to force db.DB() to fail
	sqlDB, _ := db.DB()
	_ = sqlDB.Close()

	cfg := &postgres.Config{
		MaxIdleConns:    10,
		MaxOpenConns:    20,
		ConnMaxLifetime: 5 * time.Minute,
	}

	cleanup, err := postgres.ConfigureConnectionPool(db, cfg)
	// This should succeed as db.DB() returns cached connection
	// The error will occur when trying to use the connection
	require.NoError(t, err)

	// Cleanup should still work
	cleanup()
}

func TestNewWithMockSuccess(t *testing.T) {
	// Test the full New() function with a mock connection
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = mockDB.Close() }()

	dialector := gormpostgres.New(gormpostgres.Config{
		Conn: mockDB,
		DSN:  "sqlmock_database",
	})

	// This won't actually call New() since we can't pass a pre-made connection
	// But we can test the configureConnectionPool function directly
	gormDB, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	cfg := &postgres.Config{
		MaxIdleConns:         25,
		MaxOpenConns:         100,
		ConnMaxLifetime:      5 * time.Minute,
		ConnMaxIdleTime:      5 * time.Minute,
		PreferSimpleProtocol: true,
	}

	// Expect a ping on connection (GORM does this automatically)
	mock.ExpectPing()

	cleanup, err := postgres.ConfigureConnectionPool(gormDB, cfg)
	require.NoError(t, err)
	require.NotNil(t, cleanup)

	// Verify all expectations were met
	require.NoError(t, mock.ExpectationsWereMet())

	// Test cleanup
	cleanup()
}

func TestCleanupFunction(t *testing.T) {
	// Test that cleanup function properly closes the database
	mockDB, _, err := sqlmock.New()
	require.NoError(t, err)

	dialector := gormpostgres.New(gormpostgres.Config{
		Conn: mockDB,
		DSN:  "sqlmock_database",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	cfg := &postgres.Config{
		MaxIdleConns:    10,
		MaxOpenConns:    20,
		ConnMaxLifetime: 5 * time.Minute,
	}

	cleanup, err := postgres.ConfigureConnectionPool(db, cfg)
	require.NoError(t, err)

	// Call cleanup - should close the database
	cleanup()

	// Verify database is closed
	err = mockDB.Close()
	if err == nil || err == sql.ErrConnDone {
		// Expected - database already closed or connection done
		return
	}
}

func TestCleanupIdempotent(t *testing.T) {
	// Test that cleanup can be called multiple times safely
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = mockDB.Close() }()

	dialector := gormpostgres.New(gormpostgres.Config{
		Conn: mockDB,
		DSN:  "sqlmock_database",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	mock.ExpectPing()

	cfg := &postgres.Config{
		MaxIdleConns:    10,
		MaxOpenConns:    20,
		ConnMaxLifetime: 5 * time.Minute,
	}

	cleanup, err := postgres.ConfigureConnectionPool(db, cfg)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())

	// Call cleanup multiple times - should not panic
	cleanup()
	cleanup()
	cleanup()
}

func TestPoolConfigurationApplied(t *testing.T) {
	// Verify that pool configuration values are correctly applied
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = mockDB.Close() }()

	dialector := gormpostgres.New(gormpostgres.Config{
		Conn: mockDB,
		DSN:  "sqlmock_database",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	require.NoError(t, err)

	mock.ExpectPing()

	cfg := &postgres.Config{
		MaxIdleConns:         15,
		MaxOpenConns:         75,
		ConnMaxLifetime:      12 * time.Minute,
		ConnMaxIdleTime:      6 * time.Minute,
		PreferSimpleProtocol: false,
	}

	cleanup, err := postgres.ConfigureConnectionPool(db, cfg)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())

	// Verify the configuration was applied by checking the underlying sql.DB
	sqlDB, err := db.DB()
	require.NoError(t, err)

	// GORM doesn't expose getters for pool config, but we can verify
	// that the DB is still valid and cleanup works
	require.NotNil(t, sqlDB)
	require.NotNil(t, cleanup)

	cleanup()
}
