package errors

type ErrResponse struct {
	Timestamp int64    `json:"timestamp"`
	Message   string   `json:"message"`
	ErrorCode string   `json:"error_code"`
	Errors    []string `json:"errors,omitempty"`
	Path      string   `json:"path"`
	RequestID *string  `json:"request_id,omitempty"`
}
