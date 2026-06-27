package metrics

import "github.com/prometheus/client_golang/prometheus"

// Counter is a monotonic counter metric.
type Counter struct {
	vec *prometheus.CounterVec
}

// NewCounter creates a new counter metric.
func NewCounter(name, help string, labels []string) *Counter {
	return &Counter{
		vec: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: name,
			Help: help,
		}, labels),
	}
}

// Inc increments the counter by 1.
func (c *Counter) Inc(labels ...string) {
	c.vec.WithLabelValues(labels...).Inc()
}

// Add adds the given value to the counter.
func (c *Counter) Add(v float64, labels ...string) {
	c.vec.WithLabelValues(labels...).Add(v)
}

// MustRegister registers this counter with the default registry.
func (c *Counter) MustRegister() {
	MustRegisterDefault(c.vec)
}

// Gauge is a metric that can go up or down.
type Gauge struct {
	vec *prometheus.GaugeVec
}

// NewGauge creates a new gauge metric.
func NewGauge(name, help string, labels []string) *Gauge {
	return &Gauge{
		vec: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: name,
			Help: help,
		}, labels),
	}
}

// Set sets the gauge value.
func (g *Gauge) Set(v float64, labels ...string) {
	g.vec.WithLabelValues(labels...).Set(v)
}

// Inc increments the gauge by 1.
func (g *Gauge) Inc(labels ...string) {
	g.vec.WithLabelValues(labels...).Inc()
}

// Dec decrements the gauge by 1.
func (g *Gauge) Dec(labels ...string) {
	g.vec.WithLabelValues(labels...).Dec()
}

// Add adds the given value to the gauge.
func (g *Gauge) Add(v float64, labels ...string) {
	g.vec.WithLabelValues(labels...).Add(v)
}

// MustRegister registers this gauge with the default registry.
func (g *Gauge) MustRegister() {
	MustRegisterDefault(g.vec)
}

// Histogram is a metric that counts observations into configurable buckets.
type Histogram struct {
	vec *prometheus.HistogramVec
}

// NewHistogram creates a new histogram metric.
func NewHistogram(name, help string, labels []string, buckets []float64) *Histogram {
	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}
	return &Histogram{
		vec: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    name,
			Help:    help,
			Buckets: buckets,
		}, labels),
	}
}

// Observe records a value.
func (h *Histogram) Observe(v float64, labels ...string) {
	h.vec.WithLabelValues(labels...).Observe(v)
}

// MustRegister registers this histogram with the default registry.
func (h *Histogram) MustRegister() {
	MustRegisterDefault(h.vec)
}

// Summary is a metric that calculates configurable quantiles over a sliding window.
type Summary struct {
	vec *prometheus.SummaryVec
}

// NewSummary creates a new summary metric.
func NewSummary(name, help string, labels []string, objectives map[float64]float64) *Summary {
	if len(objectives) == 0 {
		objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	}
	return &Summary{
		vec: prometheus.NewSummaryVec(prometheus.SummaryOpts{
			Name:       name,
			Help:       help,
			Objectives: objectives,
		}, labels),
	}
}

// Observe records a value.
func (s *Summary) Observe(v float64, labels ...string) {
	s.vec.WithLabelValues(labels...).Observe(v)
}

// MustRegister registers this summary with the default registry.
func (s *Summary) MustRegister() {
	MustRegisterDefault(s.vec)
}
