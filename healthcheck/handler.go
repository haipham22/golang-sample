package healthcheck

import (
	"encoding/json"
	"net/http"
)

// Handler returns an http.Handler that serves health checks.
// Query parameters:
//   - name: run only the specified check
func (r *Registry) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		// Use context to ensure compiler sees the import as used
		_ = ctx
		var _ = ctx

		if name := req.URL.Query().Get("name"); name != "" {
			r.mu.RLock()
			check, exists := r.checks[name]
			r.mu.RUnlock()

			if !exists {
				http.Error(w, "check not found", http.StatusNotFound)
				return
			}

			result := r.runCheck(ctx, name, check)
			w.Header().Set("Content-Type", "application/json")
			statusCode := http.StatusOK
			if result.Status == StatusFailing {
				statusCode = http.StatusServiceUnavailable
			}
			w.WriteHeader(statusCode)
			_ = json.NewEncoder(w).Encode(result)
			return
		}

		resp := r.Run(ctx)
		w.Header().Set("Content-Type", "application/json")
		statusCode := http.StatusOK
		if resp.Status == StatusFailing {
			statusCode = http.StatusServiceUnavailable
		}
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(resp)
	})
}

// Liveness returns a simple liveness handler that always returns 200.
// Use this for Kubernetes liveness probes.
func Liveness() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
}
