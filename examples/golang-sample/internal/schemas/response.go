package schemas

import "time"

type Response[T any] struct {
	Data       T           `json:"data"`
	Timestamp  int64       `json:"timestamp"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

func NewResponse[T any](data T) Response[T] {
	return Response[T]{
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
}

type Pagination struct {
	Page    uint32 `json:"page"`
	PerPage uint32 `json:"per_page"`
	Total   uint32 `json:"total"`
}

func NewPaginationResponse[T any](data T, currentPage, perPage, total uint32) Response[T] {
	return Response[T]{
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
		Pagination: &Pagination{
			Page:    currentPage,
			PerPage: perPage,
			Total:   total,
		},
	}
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
