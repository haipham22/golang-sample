package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricTypes(t *testing.T) {
	tests := []struct {
		name      string
		metric    string
		labels    []string
		setup     func(*Registry)
		assertion string
	}{
		{
			name:      "counter increments",
			metric:    "counter",
			labels:    []string{"label"},
			assertion: "test_counter",
			setup: func(r *Registry) {
				counter := NewCounter("test_counter", "Test counter", []string{"label"})
				r.MustRegister(counter.vec)
				counter.Inc("a")
				counter.Add(5, "b")
			},
		},
		{
			name:      "gauge operations",
			metric:    "gauge",
			labels:    []string{"label"},
			assertion: "test_gauge",
			setup: func(r *Registry) {
				gauge := NewGauge("test_gauge", "Test gauge", []string{"label"})
				r.MustRegister(gauge.vec)
				gauge.Set(10, "a")
				gauge.Inc("a")
				gauge.Dec("a")
				gauge.Add(2.5, "a")
			},
		},
		{
			name:      "histogram observations",
			metric:    "histogram",
			labels:    []string{"label"},
			assertion: "test_hist",
			setup: func(r *Registry) {
				hist := NewHistogram("test_hist", "Test histogram", []string{"label"}, []float64{1, 5, 10})
				r.MustRegister(hist.vec)
				hist.Observe(2, "a")
				hist.Observe(7, "a")
				hist.Observe(15, "a")
			},
		},
		{
			name:      "summary observations",
			metric:    "summary",
			labels:    []string{"label"},
			assertion: "test_summary",
			setup: func(r *Registry) {
				sum := NewSummary("test_summary", "Test summary", []string{"label"}, nil)
				r.MustRegister(sum.vec)
				sum.Observe(1, "a")
				sum.Observe(2, "a")
				sum.Observe(3, "a")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := New()
			tt.setup(registry)

			h := registry.Handler()
			req := httptest.NewRequest("GET", "/metrics", http.NoBody)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			assertEqual(t, http.StatusOK, rec.Code)
			assertContains(t, rec.Body.String(), tt.assertion)
		})
	}
}

func TestHandler(t *testing.T) {
	registry := New()
	counter := NewCounter("my_test_counter", "Test counter", nil)
	registry.MustRegister(counter.vec)

	// Increment the counter so it appears in metrics output
	counter.Inc()

	h := registry.Handler()
	requireNotNil(t, h)

	req := httptest.NewRequest("GET", "/metrics", http.NoBody)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assertEqual(t, http.StatusOK, rec.Code)
	body := rec.Body.String()
	assertContains(t, body, "my_test_counter")
}

func TestHandlerDefault(t *testing.T) {
	defaultRegistry = New()

	h := HandlerDefault()
	requireNotNil(t, h)

	req := httptest.NewRequest("GET", "/metrics", http.NoBody)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assertEqual(t, http.StatusOK, rec.Code)
}

func TestNewHistogramDefaultBuckets(t *testing.T) {
	hist := NewHistogram("test_hist_buckets", "Test histogram", nil, nil)
	requireNotNil(t, hist)
	requireNotNil(t, hist.vec)

	hist.Observe(0.5)
	hist.Observe(5)
	hist.Observe(50)
}

func TestNewSummaryDefaultObjectives(t *testing.T) {
	sum := NewSummary("test_summary_obj", "Test summary", nil, nil)
	requireNotNil(t, sum)
	requireNotNil(t, sum.vec)

	sum.Observe(1)
	sum.Observe(2)
	sum.Observe(3)
}

func TestMustRegister(t *testing.T) {
	registry := New()
	counter := NewCounter("reg_counter", "Registered counter", nil)
	assertNotPanics(t, func() { registry.MustRegister(counter.vec) })

	gauge := NewGauge("reg_gauge", "Registered gauge", nil)
	assertNotPanics(t, func() { registry.MustRegister(gauge.vec) })

	hist := NewHistogram("reg_hist", "Registered histogram", nil, nil)
	assertNotPanics(t, func() { registry.MustRegister(hist.vec) })

	sum := NewSummary("reg_summary", "Registered summary", nil, nil)
	assertNotPanics(t, func() { registry.MustRegister(sum.vec) })
}

func TestMetricVec(t *testing.T) {
	tests := []struct {
		name      string
		metric    string
		labels    []string
		assertion string
		setup     func(*Registry)
	}{
		{
			name:      "counter with multiple labels",
			metric:    "counter",
			labels:    []string{"method", "status"},
			assertion: "test_counter_total",
			setup: func(r *Registry) {
				counter := NewCounter("test_counter_total", "Test counter", []string{"method", "status"})
				r.MustRegister(counter.vec)
				counter.Inc("GET", "200")
				counter.Inc("POST", "201")
				counter.Inc("GET", "404")
			},
		},
		{
			name:      "gauge with multiple labels",
			metric:    "gauge",
			labels:    []string{"node"},
			assertion: "test_gauge_vec",
			setup: func(r *Registry) {
				gauge := NewGauge("test_gauge_vec", "Test gauge", []string{"node"})
				r.MustRegister(gauge.vec)
				gauge.Set(10, "node1")
				gauge.Set(20, "node2")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := New()
			tt.setup(registry)

			h := registry.Handler()
			req := httptest.NewRequest("GET", "/metrics", http.NoBody)
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			assertEqual(t, http.StatusOK, rec.Code)
			assertContains(t, rec.Body.String(), tt.assertion)
		})
	}
}
