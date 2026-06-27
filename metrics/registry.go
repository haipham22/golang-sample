package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Registry wraps prometheus.Registry with convenience methods.
type Registry struct {
	registry *prometheus.Registry
}

// New creates a new metrics registry.
func New() *Registry {
	return &Registry{
		registry: prometheus.NewRegistry(),
	}
}

// Default returns the default global registry.
func Default() *Registry {
	return defaultRegistry
}

var defaultRegistry = New()

// MustRegister registers metrics with the registry, panicing on error.
func (r *Registry) MustRegister(cs ...prometheus.Collector) {
	r.registry.MustRegister(cs...)
}

// MustRegisterDefault registers metrics with the default registry.
func MustRegisterDefault(cs ...prometheus.Collector) {
	defaultRegistry.MustRegister(cs...)
}

// Handler returns an http.Handler for serving metrics.
func (r *Registry) Handler() http.Handler {
	return promhttp.HandlerFor(r.registry, promhttp.HandlerOpts{})
}

// HandlerDefault returns an http.Handler for the default registry.
func HandlerDefault() http.Handler {
	return defaultRegistry.Handler()
}
