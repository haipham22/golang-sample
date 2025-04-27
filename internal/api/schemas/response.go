package schemas

type Response struct {
	Data       interface{} `json:"data"`
	Timestamp  int64       `json:"timestamp"`
	Pagination *Pagination `json:"pagination,omitempty"`
	StatusCode int         `json:"status_code"`
}

type Pagination struct {
	Page    uint32 `json:"page"`
	PerPage uint32 `json:"per_page"`
	Total   uint32 `json:"total"`
}

type ErrResponseBody struct {
	Timestamp int64          `json:"timestamp"`
	Msg       string         `json:"msg"`
	ErrorCode int            `json:"error_code"`
	Errors    []*ErrorDetail `json:"errors"`
	Path      string         `json:"path"`
}

type ErrorDetail struct {
	Msg       string                 `json:"msg"`
	MsgValues map[string]interface{} `json:"msg_values"`
	ErrorCode int                    `json:"error_code"`
	Property  string                 `json:"property"`
	Detail    string                 `json:"detail"`
}
