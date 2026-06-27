package middleware

import "net/http"

// CORSConfig configures CORS middleware.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns default CORS configuration.
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           86400, // 24 hours
	}
}

// CORS returns CORS middleware.
func CORS(config *CORSConfig) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check allowed origins
			if origin != "" && isAllowedOrigin(origin, config.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if len(config.AllowedMethods) > 0 {
				methods := joinStrings(config.AllowedMethods, ", ")
				w.Header().Set("Access-Control-Allow-Methods", methods)
			}

			if len(config.AllowedHeaders) > 0 {
				headers := joinStrings(config.AllowedHeaders, ", ")
				w.Header().Set("Access-Control-Allow-Headers", headers)
			}

			if len(config.ExposedHeaders) > 0 {
				headers := joinStrings(config.ExposedHeaders, ", ")
				w.Header().Set("Access-Control-Expose-Headers", headers)
			}

			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", string(rune(config.MaxAge)))
			}

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isAllowedOrigin(origin string, allowed []string) bool {
	for _, a := range allowed {
		if a == "*" || a == origin {
			return true
		}
	}
	return false
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
