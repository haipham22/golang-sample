package redis

import (
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ParseDSN parses a Redis DSN string into options.
// Supported formats:
//   - redis://[:password@]host[:port][/db][?options]
//   - rediss://[:password@]host[:port][/db][?options] (TLS)
//
// Query options:
//   - db: database number (overrides path)
//   - pool_size: connection pool size
//   - min_idle: minimum idle connections
//   - max_retries: maximum retry attempts
//   - dial_timeout: dial timeout in seconds
//   - read_timeout: read timeout in seconds
//   - write_timeout: write timeout in seconds
//   - pool_timeout: pool timeout in seconds
//   - idle_timeout: idle timeout in seconds
func ParseDSN(dsn string) (addr string, opts []Option, err error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", nil, err
	}

	// Extract address
	addr = u.Host
	if addr == "" {
		addr = "localhost:6379"
	}

	opts = parseAuth(u)
	opts = append(opts, parseDB(u)...)
	opts = append(opts, parseQueryOptions(u)...)

	return addr, opts, nil
}

func parseAuth(u *url.URL) []Option {
	if u.User == nil {
		return nil
	}
	password, hasPassword := u.User.Password()
	if !hasPassword {
		return nil
	}
	return []Option{WithPassword(password)}
}

func parseDB(u *url.URL) []Option {
	if u.Path == "" || u.Path == "/" {
		return nil
	}
	dbStr := strings.TrimPrefix(u.Path, "/")
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		return nil
	}
	return []Option{WithDB(db)}
}

func parseQueryOptions(u *url.URL) []Option {
	var opts []Option
	for key, values := range u.Query() {
		if len(values) == 0 {
			continue
		}
		if opt := parseIntOption(key, values[0]); opt != nil {
			opts = append(opts, opt)
		} else if opt := parseDurationOption(key, values[0]); opt != nil {
			opts = append(opts, opt)
		}
	}
	return opts
}

func parseIntOption(key, value string) Option {
	switch key {
	case "db":
		if n, err := strconv.Atoi(value); err == nil {
			return WithDB(n)
		}
	case "pool_size":
		if n, err := strconv.Atoi(value); err == nil {
			return WithPoolSize(n)
		}
	case "min_idle":
		if n, err := strconv.Atoi(value); err == nil {
			return WithMinIdleConns(n)
		}
	case "max_retries":
		if n, err := strconv.Atoi(value); err == nil {
			return WithMaxRetries(n)
		}
	}
	return nil
}

func parseDurationOption(key, value string) Option {
	secs, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil
	}
	d := secsToDuration(secs)

	switch key {
	case "dial_timeout":
		return WithDialTimeout(d)
	case "read_timeout":
		return WithReadTimeout(d)
	case "write_timeout":
		return WithWriteTimeout(d)
	case "pool_timeout":
		return WithPoolTimeout(d)
	case "idle_timeout":
		return WithConnMaxIdleTime(d)
	}
	return nil
}

func secsToDuration(secs float64) time.Duration {
	return time.Duration(secs * float64(time.Second))
}
