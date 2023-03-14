package http

import (
	"go.uber.org/zap"
	"net/http"
	config2 "paperback-vbook-converter/config"
)

type ClientRequestHandler struct {
	logger     *zap.SugaredLogger
	config     config2.Config
	httpClient *http.Client
}

func NewHttpClient(logger *zap.SugaredLogger, config config2.Config) *ClientRequestHandler {
	return &ClientRequestHandler{
		logger:     logger,
		config:     config,
		httpClient: &http.Client{},
	}
}

func (c ClientRequestHandler) Get(url string) error {
	return nil
}
