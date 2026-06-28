# metrics

Prometheus metrics integration with registry and middleware.

## Overview

The `metrics` package provides Prometheus metrics integration with a default registry, custom registry support, and HTTP middleware for automatic request tracking.

## Key Types

### Registry

```go
type Registry struct {
    registry *prometheus.Registry
}
```

## Key Functions

### Registry Functions

```go
// Create new registry
func New() *Registry

// Get default global registry
func Default() *Registry

// Register metrics with custom registry
func (r *Registry) MustRegister(cs ...prometheus.Collector)

// Register metrics with default registry
func MustRegisterDefault(cs ...prometheus.Collector)

// Get HTTP handler for custom registry
func (r *Registry) Handler() http.Handler

// Get HTTP handler for default registry
func HandlerDefault() http.Handler
```

### Middleware Functions

```go
// Create metrics middleware
func Middleware(opts ...MiddlewareOption) func(http.Handler) http.Handler

// Middleware options
func WithRegistry(registry *Registry) MiddlewareOption
func WithSubsystem(subsystem string) MiddlewareOption
func WithDurationBuckets(buckets []float64) MiddlewareOption
```

## Usage

### Basic Metrics

```go
import "github.com/haipham22/govern/metrics"
import "github.com/prometheus/client_golang/prometheus"

// Create counter
var requestsTotal = prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "path", "status"},
)

// Register with default registry
metrics.MustRegisterDefault(requestsTotal)

// Increment metric
requestsTotal.WithLabelValues("GET", "/api/users", "200").Inc()
```

### Custom Registry

```go
// Create custom registry
registry := metrics.New()

// Create gauge
var activeConnections = prometheus.NewGauge(
    prometheus.GaugeOpts{
        Name: "db_active_connections",
        Help: "Number of active database connections",
    },
)

// Register with custom registry
registry.MustRegister(activeConnections)

// Set gauge value
activeConnections.Set(10)

// Serve metrics from custom registry
http.Handle("/metrics", registry.Handler())
```

### Histogram

```go
// Create histogram
var requestDuration = prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name:    "http_request_duration_seconds",
        Help:    "HTTP request duration in seconds",
        Buckets: []float64{0.1, 0.5, 1, 2.5, 5, 10},
    },
    []string{"method", "path"},
)

metrics.MustRegisterDefault(requestDuration)

// Record duration
start := time.Now()
// ... handle request ...
duration := time.Since(start).Seconds()
requestDuration.WithLabelValues("GET", "/api/users").Observe(duration)
```

### Summary

```go
// Create summary
var responseSize = prometheus.NewSummaryVec(
    prometheus.SummaryOpts{
        Name: "http_response_size_bytes",
        Help: "HTTP response size in bytes",
    },
    []string{"method", "path"},
)

metrics.MustRegisterDefault(responseSize)

// Record value
responseSize.WithLabelValues("GET", "/api/users").Observe(float64(len(data)))
```

## Middleware

### Basic Middleware

```go
import "github.com/haipham22/govern/metrics"

// Apply middleware to HTTP handler
var myHandler http.Handler = // ...

handler = metrics.Middleware()(myHandler)
http.Handle("/api", handler)

// Metrics endpoint
http.Handle("/metrics", metrics.HandlerDefault())
```

### With Custom Options

```go
// Custom registry
registry := metrics.New()

// Middleware with options
handler = metrics.Middleware(
    metrics.WithRegistry(registry),
    metrics.WithSubsystem("myapp"),
    metrics.WithDurationBuckets([]float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}),
)(myHandler)

http.Handle("/api", handler)
http.Handle("/metrics", registry.Handler())
```

### With Echo Framework

```go
import "github.com/haipham22/govern/metrics"
import "github.com/labstack/echo/v4"

e := echo.New()

// Wrap middleware
e.Use(echo.WrapMiddleware(metrics.Middleware()))

// Metrics endpoint
e.GET("/metrics", echo.WrapHandler(metrics.HandlerDefault()))
```

### Metrics Collected by Middleware

The middleware automatically tracks:

```go
// Counter: http_requests_total
// Labels: method, path, status

// Histogram: http_request_duration_seconds
// Labels: method, path
```

## Metrics Format

### Prometheus Text Format

```
# HELP http_requests_total Total number of HTTP requests
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/api/users",status="200"} 1234

# HELP http_request_duration_seconds HTTP request duration in seconds
# TYPE http_request_duration_seconds histogram
http_request_duration_seconds_bucket{method="GET",path="/api/users",le="0.1"} 100
http_request_duration_seconds_bucket{method="GET",path="/api/users",le="0.5"} 150
http_request_duration_seconds_bucket{method="GET",path="/api/users",le="1"} 180
http_request_duration_seconds_bucket{method="GET",path="/api/users",le="+Inf"} 200
http_request_duration_seconds_sum{method="GET",path="/api/users"} 15.5
http_request_duration_seconds_count{method="GET",path="/api/users"} 200
```

## Common Metrics Patterns

### Database Operations

```go
var (
    dbQueryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "db_query_duration_seconds",
            Help: "Database query duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"query", "status"},
    )

    dbConnectionsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "db_connections_total",
            Help: "Total number of database connections",
        },
        []string{"state"}, // open, closed
    )
)

metrics.MustRegisterDefault(dbQueryDuration, dbConnectionsTotal)
```

### Business Metrics

```go
var (
    ordersTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "orders_total",
            Help: "Total number of orders",
        },
        []string{"status"}, // created, paid, shipped, cancelled
    )

    orderAmount = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "order_amount_dollars",
            Help: "Order amount in dollars",
            Buckets: []float64{10, 25, 50, 100, 250, 500, 1000},
        },
        []string{"currency"},
    )
)

metrics.MustRegisterDefault(ordersTotal, orderAmount)
```

### Cache Metrics

```go
var (
    cacheHitsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_hits_total",
            Help: "Total number of cache hits",
        },
        []string{"cache"},
    )

    cacheMissesTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "cache_misses_total",
            Help: "Total number of cache misses",
        },
        []string{"cache"},
    )
)

metrics.MustRegisterDefault(cacheHitsTotal, cacheMissesTotal)
```

## Best Practices

1. **Use meaningful names** - Follow Prometheus naming conventions
2. **Add relevant labels** - Include dimensions for filtering
3. **Choose appropriate metric types** - Counter, Gauge, Histogram, Summary
4. **Set proper buckets** - For histograms, use relevant percentiles
5. **Document metrics** - Add helpful descriptions
6. **Avoid high cardinality** - Don't use labels with many unique values

## References

- [metrics/registry.go](../../metrics/registry.go) - Registry implementation
- [metrics/middleware.go](../../metrics/middleware.go) - Middleware implementation
- [metrics/metrics-types.go](../../metrics/metrics-types.go) - Type definitions
