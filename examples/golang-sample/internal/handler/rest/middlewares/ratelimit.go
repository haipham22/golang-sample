package middlewares

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	// RequestsPerMinute is the maximum number of requests allowed per minute
	RequestsPerMinute int
	// WindowSize is the sliding window size in seconds
	WindowSize int
}

// DefaultRateLimiterConfig returns default configuration for rate limiting
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RequestsPerMinute: 10,
		WindowSize:        60, // 1 minute sliding window
	}
}

// ipLimiter tracks requests for a single IP address
type ipLimiter struct {
	mu       sync.Mutex
	requests []time.Time
	window   time.Duration
	limit    int
}

// newIPLimiter creates a new IP rate limiter
func newIPLimiter(requestsPerMinute int, windowSize int) *ipLimiter {
	return &ipLimiter{
		requests: make([]time.Time, 0, requestsPerMinute),
		window:   time.Duration(windowSize) * time.Second,
		limit:    requestsPerMinute,
	}
}

// allow checks if a request should be allowed
func (il *ipLimiter) allow() bool {
	il.mu.Lock()
	defer il.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-il.window)

	// Remove requests outside the window
	validIdx := len(il.requests) // Default to all requests being expired
	for i, reqTime := range il.requests {
		if reqTime.After(cutoff) {
			validIdx = i
			break
		}
	}
	il.requests = il.requests[validIdx:]

	// Check if limit exceeded
	if len(il.requests) >= il.limit {
		return false
	}

	// Add current request
	il.requests = append(il.requests, now)
	return true
}

// RateLimit creates a rate limiting middleware with default config (10 req/min)
// The cleanup goroutine will be cancelled when the provided context is done
func RateLimit(ctx context.Context) echo.MiddlewareFunc {
	return RateLimitWithConfig(ctx, DefaultRateLimiterConfig())
}

// RateLimitWithConfig creates a rate limiting middleware with custom config
// The cleanup goroutine will be cancelled when the provided context is done
func RateLimitWithConfig(ctx context.Context, config RateLimiterConfig) echo.MiddlewareFunc {
	// Map to store limiters per IP
	limiters := make(map[string]*ipLimiter)
	var mu sync.Mutex

	// Cleanup goroutine to remove stale limiters
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mu.Lock()
				now := time.Now()
				for ip, limiter := range limiters {
					limiter.mu.Lock()
					// Remove if no recent requests (older than window + cleanup margin)
					cutoff := now.Add(-limiter.window - time.Minute)
					stale := true
					for _, reqTime := range limiter.requests {
						if reqTime.After(cutoff) {
							stale = false
							break
						}
					}
					limiter.mu.Unlock()
					if stale {
						delete(limiters, ip)
					}
				}
				mu.Unlock()
			case <-ctx.Done():
				// Context cancelled, stop cleanup goroutine
				return
			}
		}
	}()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get client IP (use Echo's IPExtractor if configured)
			ip := c.RealIP()
			if ip == "" {
				// Fallback to RemoteAddr
				ip = c.Request().RemoteAddr
			}

			// Get or create limiter for this IP
			mu.Lock()
			limiter, exists := limiters[ip]
			if !exists {
				limiter = newIPLimiter(config.RequestsPerMinute, config.WindowSize)
				limiters[ip] = limiter
			}
			mu.Unlock()

			// Check if request is allowed
			if !limiter.allow() {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "Too many requests",
					"msg":   "Rate limit exceeded. Please try again later.",
				})
			}

			return next(c)
		}
	}
}
