# govern/metrics

Import: `github.com/haipham22/govern/metrics`

Prometheus metrics with HTTP middleware for automatic request tracking.

## Use When

- App exposes Prometheus metrics.

## Counter / Gauge / Histogram

```go
import "github.com/haipham22/govern/metrics"

requests := metrics.NewCounter("http_requests_total", "Total HTTP requests", []string{"method", "status"})
requests.MustRegister()
requests.Inc("GET", "200")
requests.Add(1, "POST", "201")
```

## HTTP Middleware (auto-tracking)

```go
handler := metrics.HTTPMiddleware(yourHandler, "service")
http.Handle("/metrics", metrics.HandlerDefault())
http.ListenAndServe(":8080", handler)
```

Middleware auto-records:

- `http_requests_total` — counter by method, status
- `http_request_duration_seconds` — histogram
- `http_response_size_bytes` — histogram (exponential buckets)

## Types

| Type | Use |
|---|---|
| Counter | Monotonically increasing |
| Gauge | Up/down value (set, inc, dec, add) |
| Histogram | Bucket distribution |
| Summary | Quantile over sliding window |

## Rules

- ✅ Register metrics centrally in bootstrap, not per-request.
- ✅ Use labels for dimensions (method, status, route).
- ✅ Expose `/metrics` endpoint.
- ✅ Prefer middleware over manual instrumentation for HTTP.
- ❌ Do not create metrics dynamically (label cardinality explosion).
- ❌ Do not put high-cardinality values (user ID, email) in labels.

## Avoid

- Manual Prometheus instrumentation when middleware covers it.
- Unbounded label cardinality.

## Reference

Source: [`metrics/`](../../../../../../../metrics/). Uses `github.com/prometheus/client_golang`.
