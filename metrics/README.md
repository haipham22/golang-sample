# metrics

Prometheus metrics integration with HTTP middleware.

## Features

- Counter, Gauge, Histogram, Summary metrics
- Default global registry
- HTTP middleware for automatic request tracking
- Label support for multi-dimensional metrics

## Usage

```go
import "github.com/haipham22/govern/metrics"

// Register metrics
requestsTotal := metrics.NewCounter("http_requests_total", "Total HTTP requests", []string{"method", "status"})
requestsTotal.MustRegister()

// Record metrics
requestsTotal.Inc("GET", "200")
requestsTotal.Add(1, "POST", "201")

// HTTP middleware
handler := metrics.HTTPMiddleware(yourHandler, "service")
http.ListenAndServe(":8080", handler)

// Serve metrics endpoint
http.Handle("/metrics", metrics.HandlerDefault())
```

## Metrics Types

- **Counter**: Monotonically increasing value
- **Gauge**: Up/down value (set, inc, dec, add)
- **Histogram**: Configurable bucket distribution
- **Summary**: Quantile calculation over sliding window

## Middleware

The HTTP middleware automatically records:
- `http_requests_total` - counter by method, status
- `http_request_duration_seconds` - histogram
- `http_response_size_bytes` - histogram with exponential buckets
