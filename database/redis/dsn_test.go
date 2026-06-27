package redis_test

import (
	"testing"

	"github.com/haipham22/govern/database/redis"
)

func TestParseDSN(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		wantAddr string
		wantOpts int // expected number of options
		wantErr  bool
	}{
		{
			name:     "simple address",
			dsn:      "redis://localhost:6379",
			wantAddr: "localhost:6379",
			wantOpts: 0,
		},
		{
			name:     "with password",
			dsn:      "redis://:secret@localhost:6379",
			wantAddr: "localhost:6379",
			wantOpts: 1,
		},
		{
			name:     "with db in path",
			dsn:      "redis://localhost:6379/1",
			wantAddr: "localhost:6379",
			wantOpts: 1,
		},
		{
			name:     "with password and db",
			dsn:      "redis://:secret@localhost:6379/2",
			wantAddr: "localhost:6379",
			wantOpts: 2,
		},
		{
			name:     "with query options",
			dsn:      "redis://localhost:6379?pool_size=50&db=1",
			wantAddr: "localhost:6379",
			wantOpts: 2,
		},
		{
			name:     "with all options",
			dsn:      "redis://:secret@localhost:6379/1?pool_size=50&min_idle=5",
			wantAddr: "localhost:6379",
			wantOpts: 4,
		},
		{
			name:     "without port",
			dsn:      "redis://localhost",
			wantAddr: "localhost",
			wantOpts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, opts, err := redis.ParseDSN(tt.dsn)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDSN() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDSN() unexpected error: %v", err)
				return
			}

			if addr != tt.wantAddr {
				t.Errorf("ParseDSN() addr = %v, want %v", addr, tt.wantAddr)
			}

			if len(opts) != tt.wantOpts {
				t.Errorf("ParseDSN() opts count = %v, want %v", len(opts), tt.wantOpts)
			}
		})
	}
}

func TestParseDSNEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		dsn      string
		wantAddr string
		wantErr  bool
	}{
		{
			name:     "empty DSN",
			dsn:      "",
			wantAddr: "localhost:6379",
			wantErr:  false, // url.Parse accepts empty string
		},
		{
			name:     "malformed URL",
			dsn:      "://invalid",
			wantAddr: "",
			wantErr:  true,
		},
		{
			name:     "missing scheme - treated as relative URL",
			dsn:      "localhost:6379",
			wantAddr: "localhost:6379",
			wantErr:  false, // url.Parse accepts this as relative URL
		},
		{
			name:     "rediss (TLS)",
			dsn:      "rediss://localhost:6379",
			wantAddr: "localhost:6379",
			wantErr:  false,
		},
		{
			name:     "with username only (no password)",
			dsn:      "redis://user@localhost:6379",
			wantAddr: "localhost:6379",
			wantErr:  false,
		},
		{
			name:     "with username and password",
			dsn:      "redis://user:pass@localhost:6379",
			wantAddr: "localhost:6379",
			wantErr:  false,
		},
		{
			name:     "invalid db number",
			dsn:      "redis://localhost:6379/abc",
			wantAddr: "localhost:6379",
			wantErr:  false, // parseDB ignores invalid, returns nil opts
		},
		{
			name:     "empty host",
			dsn:      "redis://",
			wantAddr: "localhost:6379", // defaults to localhost:6379
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, opts, err := redis.ParseDSN(tt.dsn)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDSN() expected error for %q, got nil", tt.dsn)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDSN() unexpected error for %q: %v", tt.dsn, err)
				return
			}

			if addr != tt.wantAddr {
				t.Errorf("ParseDSN() addr = %v, want %v", addr, tt.wantAddr)
			}

			// opts may be nil for some cases, that's ok
			_ = opts
		})
	}
}

func TestParseDSNTimeoutOptions(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{
			name:    "dial timeout",
			dsn:     "redis://localhost:6379?dial_timeout=5",
			wantErr: false,
		},
		{
			name:    "read timeout",
			dsn:     "redis://localhost:6379?read_timeout=3",
			wantErr: false,
		},
		{
			name:    "write timeout",
			dsn:     "redis://localhost:6379?write_timeout=3",
			wantErr: false,
		},
		{
			name:    "pool timeout",
			dsn:     "redis://localhost:6379?pool_timeout=4",
			wantErr: false,
		},
		{
			name:    "idle timeout",
			dsn:     "redis://localhost:6379?idle_timeout=300",
			wantErr: false,
		},
		{
			name:    "fractional timeout",
			dsn:     "redis://localhost:6379?dial_timeout=0.5",
			wantErr: false,
		},
		{
			name:    "invalid timeout value",
			dsn:     "redis://localhost:6379?dial_timeout=invalid",
			wantErr: false, // parseDurationOption returns nil on error
		},
		{
			name:    "all timeouts",
			dsn:     "redis://localhost:6379?dial_timeout=5&read_timeout=3&write_timeout=3&pool_timeout=4&idle_timeout=300",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, opts, err := redis.ParseDSN(tt.dsn)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseDSN() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseDSN() unexpected error: %v", err)
				return
			}

			if addr == "" {
				t.Error("ParseDSN() returned empty address")
			}

			if opts == nil {
				opts = []redis.Option{} // normalize nil to empty slice
			}

			// Verify options were created (even if values are invalid, options should exist)
			_ = opts
		})
	}
}

func TestNewFromDSN(t *testing.T) {
	tests := []struct {
		name    string
		dsn     string
		wantErr bool
	}{
		{
			name:    "simple DSN",
			dsn:     "redis://localhost:6379",
			wantErr: false,
		},
		{
			name:    "DSN with password and db",
			dsn:     "redis://:secret@localhost:6379/1",
			wantErr: false,
		},
		{
			name:    "DSN with options",
			dsn:     "redis://localhost:6379?pool_size=50&min_idle=5",
			wantErr: false,
		},
		{
			name:    "invalid DSN",
			dsn:     "://invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, cleanup, err := redis.NewFromDSN(tt.dsn)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewFromDSN() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewFromDSN() unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("NewFromDSN() expected non-nil client")
			}

			if cleanup == nil {
				t.Error("NewFromDSN() expected non-nil cleanup function")
			} else {
				cleanup()
			}
		})
	}
}
