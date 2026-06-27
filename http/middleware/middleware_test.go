package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Test helpers

// bufferWriteSyncer is a zapcore.WriteSyncer that writes to a buffer.
type bufferWriteSyncer struct {
	*bytes.Buffer
}

func (b *bufferWriteSyncer) Sync() error {
	return nil
}

// createTestLogger creates a test logger that writes to a buffer.
func createTestLogger(t *testing.T) (*zap.SugaredLogger, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	syncer := &bufferWriteSyncer{buf}
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), syncer, zapcore.InfoLevel)
	logger := zap.New(core).Sugar()
	return logger, buf
}

// parseLogEntries parses JSON log entries from buffer.
func parseLogEntries(t *testing.T, buf *bytes.Buffer) []map[string]interface{} {
	t.Helper()
	var entries []map[string]interface{}
	decoder := json.NewDecoder(buf)
	for decoder.More() {
		var entry map[string]interface{}
		if err := decoder.Decode(&entry); err != nil {
			t.Fatalf("failed to decode log entry: %v", err)
		}
		entries = append(entries, entry)
	}
	return entries
}

func createTestRequest(t *testing.T, method, path string, headers map[string]string) *http.Request {
	t.Helper()
	req := httptest.NewRequest(method, path, http.NoBody)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req
}

func assertHeader(t *testing.T, rec *httptest.ResponseRecorder, key, value string) {
	t.Helper()
	got := rec.Header().Get(key)
	assert.Equal(t, value, got, "header %s", key)
}

// Test Logging Middleware

func TestLogging(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("response"))
	})

	tests := []struct {
		name           string
		request        *http.Request
		expectedStatus int
		handler        http.HandlerFunc
		validateLogs   func(*testing.T, []map[string]interface{})
	}{
		{
			name:           "logs incoming request",
			request:        createTestRequest(t, http.MethodGet, "/test", nil),
			expectedStatus: http.StatusOK,
			handler:        handler,
			validateLogs: func(t *testing.T, logs []map[string]interface{}) {
				t.Helper()
				assert.GreaterOrEqual(t, len(logs), 1, "should have request log")
				reqLog := logs[0]
				assert.Equal(t, "Incoming request", reqLog["msg"])
				assert.Equal(t, "/test", reqLog["path"])
				assert.Equal(t, "GET", reqLog["method"])
			},
		},
		{
			name:           "logs request with query params",
			request:        createTestRequest(t, http.MethodGet, "/test?foo=bar&baz=qux", nil),
			expectedStatus: http.StatusOK,
			handler:        handler,
			validateLogs: func(t *testing.T, logs []map[string]interface{}) {
				t.Helper()
				assert.GreaterOrEqual(t, len(logs), 1)
				assert.Equal(t, "foo=bar&baz=qux", logs[0]["query"])
			},
		},
		{
			name:           "logs response with status",
			request:        createTestRequest(t, http.MethodPost, "/api/users", nil),
			expectedStatus: http.StatusCreated,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
			}),
			validateLogs: func(t *testing.T, logs []map[string]interface{}) {
				t.Helper()
				// Find response log
				var respLog map[string]interface{}
				for _, log := range logs {
					if log["msg"] == "Request completed" {
						respLog = log
						break
					}
				}
				require.NotNil(t, respLog)
				assert.Equal(t, float64(201), respLog["status"])
				assert.Contains(t, respLog, "duration")
			},
		},
		{
			name: "logs with existing request ID",
			request: createTestRequest(t, http.MethodGet, "/test", map[string]string{
				"X-Request-ID": "test-request-123",
			}),
			expectedStatus: http.StatusOK,
			handler:        handler,
			validateLogs: func(t *testing.T, logs []map[string]interface{}) {
				t.Helper()
				assert.GreaterOrEqual(t, len(logs), 1)
				assert.Equal(t, "test-request-123", logs[0]["request_id"])
			},
		},
		{
			name:           "generates request ID when missing",
			request:        createTestRequest(t, http.MethodGet, "/test", nil),
			expectedStatus: http.StatusOK,
			handler:        handler,
			validateLogs: func(t *testing.T, logs []map[string]interface{}) {
				t.Helper()
				assert.GreaterOrEqual(t, len(logs), 1)
				reqID := logs[0]["request_id"]
				assert.NotEmpty(t, reqID)
				assert.NotEqual(t, "", reqID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, buf := createTestLogger(t)
			mw := Logging(logger)

			rec := httptest.NewRecorder()
			wrapped := mw(tt.handler)

			wrapped.ServeHTTP(rec, tt.request)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			tt.validateLogs(t, parseLogEntries(t, buf))
		})
	}
}

func TestLogging_ResponseWriter(t *testing.T) {
	tests := []struct {
		name                 string
		handler              http.HandlerFunc
		expectedBytes        int
		validateBytesWritten func(*testing.T, []map[string]interface{})
	}{
		{
			name: "tracks bytes written",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("hello world"))
			},
			expectedBytes: 11,
			validateBytesWritten: func(t *testing.T, logs []map[string]interface{}) {
				t.Helper()
				for _, log := range logs {
					if log["msg"] == "Request completed" {
						assert.Equal(t, float64(11), log["bytes_written"])
					}
				}
			},
		},
		{
			name: "handles multiple writes",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("first"))
				_, _ = w.Write([]byte(" "))
				_, _ = w.Write([]byte("second"))
			},
			expectedBytes: 12,
		},
		{
			name: "handles zero bytes",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			expectedBytes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, buf := createTestLogger(t)
			mw := Logging(logger)

			rec := httptest.NewRecorder()
			wrapped := mw(tt.handler)

			wrapped.ServeHTTP(rec, createTestRequest(t, http.MethodGet, "/", nil))

			assert.Equal(t, tt.expectedBytes, rec.Body.Len())
			if tt.validateBytesWritten != nil {
				tt.validateBytesWritten(t, parseLogEntries(t, buf))
			}
		})
	}
}

func TestLogging_ContextRequestID(t *testing.T) {
	logger, _ := createTestLogger(t)
	mw := Logging(logger)

	var capturedReqID interface{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReqID = r.Context().Value(RequestIDKeyVal)
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		request    *http.Request
		validateID func(*testing.T, interface{})
	}{
		{
			name: "sets request ID in context from header",
			request: createTestRequest(t, http.MethodGet, "/", map[string]string{
				"X-Request-ID": "custom-id-123",
			}),
			validateID: func(t *testing.T, reqID interface{}) {
				t.Helper()
				assert.Equal(t, "custom-id-123", reqID)
			},
		},
		{
			name:    "generates and sets request ID in context",
			request: createTestRequest(t, http.MethodGet, "/", nil),
			validateID: func(t *testing.T, reqID interface{}) {
				t.Helper()
				assert.NotNil(t, reqID)
				assert.NotEmpty(t, reqID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			wrapped := mw(handler)

			wrapped.ServeHTTP(rec, tt.request)

			tt.validateID(t, capturedReqID)
		})
	}
}

// Test ResponseWriter

func TestResponseWriter_WriteHeader(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedStatus int
	}{
		{
			name: "explicit status",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "default status when WriteHeader not called",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("data"))
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "status only written once",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.WriteHeader(http.StatusInternalServerError) // Should be ignored
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, _ := createTestLogger(t)
			mw := Logging(logger)

			rec := httptest.NewRecorder()
			wrapped := mw(tt.handler)

			wrapped.ServeHTTP(rec, createTestRequest(t, http.MethodGet, "/", nil))

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// Test Recovery Middleware

func TestRecovery(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		expectedStatus int
		expectedBody   string
		validatePanic  func(*testing.T, []map[string]interface{})
	}{
		{
			name: "recovers from panic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("test panic")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal Server Error",
			validatePanic: func(t *testing.T, logs []map[string]interface{}) {
				t.Helper()
				assert.GreaterOrEqual(t, len(logs), 1)
				assert.Equal(t, "Recovered from panic", logs[0]["msg"])
				assert.Equal(t, "test panic", logs[0]["error"])
				assert.Contains(t, logs[0], "stack")
			},
		},
		{
			name: "logs request details during panic",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("emergency")
			},
			expectedStatus: http.StatusInternalServerError,
			validatePanic: func(t *testing.T, logs []map[string]interface{}) {
				t.Helper()
				assert.GreaterOrEqual(t, len(logs), 1)
				ctx := logs[0]
				assert.Equal(t, "emergency", ctx["error"])
				assert.Contains(t, ctx, "method")
				assert.Contains(t, ctx, "path")
			},
		},
		{
			name: "normal requests pass through",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("ok"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
		{
			name: "recovers from different panic types",
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic(123) // int panic
			},
			expectedStatus: http.StatusInternalServerError,
			validatePanic: func(t *testing.T, logs []map[string]interface{}) {
				t.Helper()
				assert.GreaterOrEqual(t, len(logs), 1)
				assert.Equal(t, float64(123), logs[0]["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, buf := createTestLogger(t)
			mw := Recovery(logger)

			rec := httptest.NewRecorder()
			wrapped := mw(tt.handler)

			wrapped.ServeHTTP(rec, createTestRequest(t, http.MethodGet, "/test", nil))

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rec.Body.String(), tt.expectedBody)
			}
			if tt.validatePanic != nil {
				tt.validatePanic(t, parseLogEntries(t, buf))
			}
		})
	}
}

func TestRecovery_WithRequestID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	tests := []struct {
		name     string
		request  *http.Request
		validate func(*testing.T, []map[string]interface{})
	}{
		{
			name: "logs request ID from context",
			request: createTestRequest(t, http.MethodGet, "/", map[string]string{
				"X-Request-ID": "panic-123",
			}),
			validate: func(t *testing.T, logs []map[string]interface{}) {
				t.Helper()
				assert.GreaterOrEqual(t, len(logs), 1)
				assert.Equal(t, "panic-123", logs[0]["request_id"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, buf := createTestLogger(t)
			mw := Recovery(logger)

			// Set request ID in context like Logging middleware would
			ctx := context.WithValue(context.Background(), RequestIDKeyVal, "panic-123")
			req := tt.request.WithContext(ctx)

			rec := httptest.NewRecorder()
			wrapped := mw(handler)
			wrapped.ServeHTTP(rec, req)

			if tt.validate != nil {
				tt.validate(t, parseLogEntries(t, buf))
			}
		})
	}
}

// Test RequestID Middleware

func TestRequestID(t *testing.T) {
	mw := RequestID()

	tests := []struct {
		name         string
		request      *http.Request
		validateResp func(*testing.T, *httptest.ResponseRecorder)
		validateCtx  func(*testing.T, *http.Request)
	}{
		{
			name:    "adds request ID to response header",
			request: createTestRequest(t, http.MethodGet, "/", nil),
			validateResp: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				reqID := rec.Header().Get("X-Request-ID")
				assert.NotEmpty(t, reqID, "request ID should be set")
			},
		},
		{
			name: "uses existing request ID from header",
			request: createTestRequest(t, http.MethodGet, "/", map[string]string{
				"X-Request-ID": "existing-123",
			}),
			validateResp: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				reqID := rec.Header().Get("X-Request-ID")
				assert.Equal(t, "existing-123", reqID)
			},
		},
		{
			name:    "sets request ID in context",
			request: createTestRequest(t, http.MethodGet, "/", nil),
			validateCtx: func(t *testing.T, req *http.Request) {
				t.Helper()
				var capturedID interface{}
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					capturedID = r.Context().Value(RequestIDKeyVal)
					w.WriteHeader(http.StatusOK)
				})

				mw := RequestID()
				rec := httptest.NewRecorder()
				wrapped := mw(handler)
				wrapped.ServeHTTP(rec, req)

				assert.NotNil(t, capturedID, "request ID should be in context")
				assert.NotEmpty(t, capturedID, "request ID should not be empty")
			},
		},
		{
			name: "propagates existing request ID to context",
			request: createTestRequest(t, http.MethodGet, "/", map[string]string{
				"X-Request-ID": "propagate-456",
			}),
			validateCtx: func(t *testing.T, req *http.Request) {
				t.Helper()
				var capturedID interface{}
				handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					capturedID = r.Context().Value(RequestIDKeyVal)
					w.WriteHeader(http.StatusOK)
				})

				mw := RequestID()
				rec := httptest.NewRecorder()
				wrapped := mw(handler)
				wrapped.ServeHTTP(rec, req)

				assert.Equal(t, "propagate-456", capturedID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			rec := httptest.NewRecorder()
			wrapped := mw(handler)

			wrapped.ServeHTTP(rec, tt.request)

			if tt.validateResp != nil {
				tt.validateResp(t, rec)
			}
			if tt.validateCtx != nil {
				tt.validateCtx(t, tt.request)
			}
		})
	}
}

func TestRequestID_Consistency(t *testing.T) {
	mw := RequestID()

	var reqIDFromContext interface{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqIDFromContext = r.Context().Value(RequestIDKeyVal)
		w.WriteHeader(http.StatusOK)
	})

	req := createTestRequest(t, http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	wrapped := mw(handler)

	wrapped.ServeHTTP(rec, req)

	respReqID := rec.Header().Get("X-Request-ID")
	assert.NotEmpty(t, respReqID, "response should have request ID")
	assert.NotEmpty(t, reqIDFromContext, "context should have request ID")
	assert.Equal(t, reqIDFromContext, respReqID, "request ID should be consistent between context and response")
}

// Test CORS Middleware

func TestCORS(t *testing.T) {
	tests := []struct {
		name            string
		config          *CORSConfig
		request         *http.Request
		validateHeaders func(*testing.T, *httptest.ResponseRecorder)
		expectedStatus  int
	}{
		{
			name:   "default configuration",
			config: DefaultCORSConfig(),
			request: createTestRequest(t, http.MethodGet, "/", map[string]string{
				"Origin": "https://example.com",
			}),
			validateHeaders: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				assertHeader(t, rec, "Access-Control-Allow-Origin", "https://example.com")
				assertHeader(t, rec, "Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				assertHeader(t, rec, "Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			},
		},
		{
			name: "wildcard origin",
			config: &CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET"},
			},
			request: createTestRequest(t, http.MethodGet, "/", map[string]string{
				"Origin": "https://any-origin.com",
			}),
			validateHeaders: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				assertHeader(t, rec, "Access-Control-Allow-Origin", "https://any-origin.com")
			},
		},
		{
			name: "specific allowed origins",
			config: &CORSConfig{
				AllowedOrigins: []string{"https://allowed.com", "https://also-allowed.com"},
				AllowedMethods: []string{"GET"},
			},
			request: createTestRequest(t, http.MethodGet, "/", map[string]string{
				"Origin": "https://allowed.com",
			}),
			validateHeaders: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				assertHeader(t, rec, "Access-Control-Allow-Origin", "https://allowed.com")
			},
		},
		{
			name: "disallowed origin",
			config: &CORSConfig{
				AllowedOrigins: []string{"https://allowed.com"},
				AllowedMethods: []string{"GET"},
			},
			request: createTestRequest(t, http.MethodGet, "/", map[string]string{
				"Origin": "https://not-allowed.com",
			}),
			validateHeaders: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
			},
		},
		{
			name: "preflight request",
			config: &CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET", "POST"},
				AllowedHeaders: []string{"Content-Type"},
				MaxAge:         3600,
			},
			request: createTestRequest(t, http.MethodOptions, "/", map[string]string{
				"Origin": "https://example.com",
			}),
			expectedStatus: http.StatusNoContent,
			validateHeaders: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				assertHeader(t, rec, "Access-Control-Allow-Methods", "GET, POST")
				assertHeader(t, rec, "Access-Control-Allow-Headers", "Content-Type")
			},
		},
		{
			name: "with credentials",
			config: &CORSConfig{
				AllowedOrigins:   []string{"https://example.com"},
				AllowCredentials: true,
				AllowedMethods:   []string{"GET"},
			},
			request: createTestRequest(t, http.MethodGet, "/", map[string]string{
				"Origin": "https://example.com",
			}),
			validateHeaders: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				assertHeader(t, rec, "Access-Control-Allow-Credentials", "true")
			},
		},
		{
			name: "exposed headers",
			config: &CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET"},
				ExposedHeaders: []string{"X-Custom-Header", "X-Another-Header"},
			},
			request: createTestRequest(t, http.MethodGet, "/", nil),
			validateHeaders: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				assertHeader(t, rec, "Access-Control-Expose-Headers", "X-Custom-Header, X-Another-Header")
			},
		},
		{
			name: "no origin header",
			config: &CORSConfig{
				AllowedOrigins: []string{"*"},
				AllowedMethods: []string{"GET"},
			},
			request: createTestRequest(t, http.MethodGet, "/", nil),
			validateHeaders: func(t *testing.T, rec *httptest.ResponseRecorder) {
				t.Helper()
				assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			mw := CORS(tt.config)
			rec := httptest.NewRecorder()
			wrapped := mw(handler)

			wrapped.ServeHTTP(rec, tt.request)

			if tt.expectedStatus != 0 {
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}
			if tt.validateHeaders != nil {
				tt.validateHeaders(t, rec)
			}
		})
	}
}

func TestCORS_DefaultConfig(t *testing.T) {
	config := DefaultCORSConfig()

	assert.NotNil(t, config)
	assert.Equal(t, []string{"*"}, config.AllowedOrigins)
	assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}, config.AllowedMethods)
	assert.Equal(t, []string{"Origin", "Content-Type", "Authorization"}, config.AllowedHeaders)
	assert.False(t, config.AllowCredentials)
	assert.Equal(t, 86400, config.MaxAge)
}

// Test CORS helper functions

func TestIsAllowedOrigin(t *testing.T) {
	tests := []struct {
		name     string
		origin   string
		allowed  []string
		expected bool
	}{
		{
			name:     "wildcard allows all",
			origin:   "https://any-origin.com",
			allowed:  []string{"*"},
			expected: true,
		},
		{
			name:     "exact match",
			origin:   "https://example.com",
			allowed:  []string{"https://example.com", "https://another.com"},
			expected: true,
		},
		{
			name:     "no match",
			origin:   "https://not-in-list.com",
			allowed:  []string{"https://example.com"},
			expected: false,
		},
		{
			name:     "empty allowed list",
			origin:   "https://example.com",
			allowed:  []string{},
			expected: false,
		},
		{
			name:     "case sensitive",
			origin:   "https://Example.com",
			allowed:  []string{"https://example.com"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAllowedOrigin(tt.origin, tt.allowed)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		sep      string
		expected string
	}{
		{
			name:     "single string",
			input:    []string{"one"},
			sep:      ", ",
			expected: "one",
		},
		{
			name:     "multiple strings",
			input:    []string{"one", "two", "three"},
			sep:      ", ",
			expected: "one, two, three",
		},
		{
			name:     "empty list",
			input:    []string{},
			sep:      ", ",
			expected: "",
		},
		{
			name:     "custom separator",
			input:    []string{"a", "b", "c"},
			sep:      "|",
			expected: "a|b|c",
		},
		{
			name:     "empty separator",
			input:    []string{"x", "y"},
			sep:      "",
			expected: "xy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinStrings(tt.input, tt.sep)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test middleware chaining

func TestMiddlewareChain(t *testing.T) {
	logger, _ := createTestLogger(t)
	logging := Logging(logger)
	recovery := Recovery(logger)
	requestID := RequestID()

	calls := []string{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, "handler")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Chain: requestID -> logging -> recovery -> handler
	wrapped := requestID(logging(recovery(handler)))

	req := createTestRequest(t, http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
	assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
}

// Test edge cases and security scenarios

func TestSecurityScenarios(t *testing.T) {
	logger, _ := createTestLogger(t)
	logging := Logging(logger)
	recovery := Recovery(logger)

	t.Run("XSS in path - logging middleware", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		maliciousPath := "/path<script>alert('xss')</script>"
		req := createTestRequest(t, http.MethodGet, maliciousPath, nil)
		rec := httptest.NewRecorder()
		wrapped := logging(handler)

		wrapped.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("header injection - request ID", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := createTestRequest(t, http.MethodGet, "/", map[string]string{
			"X-Request-ID": "id\r\nInjected-Header: malicious",
		})
		rec := httptest.NewRecorder()
		wrapped := RequestID()(handler)

		wrapped.ServeHTTP(rec, req)

		// Note: The middleware uses the value as-is. In production, applications should
		// validate and sanitize headers before using them. This test documents the behavior.
		respReqID := rec.Header().Get("X-Request-ID")
		assert.Contains(t, respReqID, "id") // At least contains the ID part
	})

	t.Run("panic in goroutine - recovery middleware", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			go func() {
				defer func() {
					_ = recover() // Prevent test crash
				}()
				panic("goroutine panic") // This won't be caught by middleware
			}()
			w.WriteHeader(http.StatusOK)
		})

		req := createTestRequest(t, http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		wrapped := recovery(handler)

		wrapped.ServeHTTP(rec, req)

		// Main handler should succeed (goroutine panic is separate)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestMiddleware_NilLogger(t *testing.T) {
	t.Run("logging middleware with nil logger", func(t *testing.T) {
		// Note: This test documents that nil logger causes panic during request handling
		// In production, ensure logger is properly initialized
		assert.Panics(t, func() {
			mw := Logging(nil)
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := createTestRequest(t, http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			wrapped := mw(handler)

			wrapped.ServeHTTP(rec, req)
		})
	})
}

// Test response writer edge cases

func TestResponseWriter_EdgeCases(t *testing.T) {
	logger, _ := createTestLogger(t)
	mw := Logging(logger)

	t.Run("write before write header", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("early write"))
		})

		rec := httptest.NewRecorder()
		wrapped := mw(handler)
		wrapped.ServeHTTP(rec, createTestRequest(t, http.MethodGet, "/", nil))

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "early write", rec.Body.String())
	})

	t.Run("multiple write header calls", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.WriteHeader(http.StatusAccepted) // Should be ignored
		})

		rec := httptest.NewRecorder()
		wrapped := mw(handler)
		wrapped.ServeHTTP(rec, createTestRequest(t, http.MethodGet, "/", nil))

		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("empty write", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte{})
		})

		rec := httptest.NewRecorder()
		wrapped := mw(handler)
		wrapped.ServeHTTP(rec, createTestRequest(t, http.MethodGet, "/", nil))

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, 0, rec.Body.Len())
	})
}
