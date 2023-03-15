package http

import (
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"go.uber.org/zap"

	config2 "paperback-vbook-converter/config"
	"paperback-vbook-converter/internal/pkg/entity"
)

type ClientRequestHandler struct {
	logger     *zap.SugaredLogger
	config     *config2.Config
	httpClient *http.Client
}

func NewHttpClient(logger *zap.SugaredLogger, config *config2.Config) *ClientRequestHandler {
	return &ClientRequestHandler{
		logger:     logger,
		config:     config,
		httpClient: &http.Client{},
	}
}

func (h ClientRequestHandler) get(url string) ([]byte, error) {
	response, err := h.httpClient.Get(url)
	if err != nil {
		h.logger.Error("httpClient.Get: get %s has error", url, err.Error())
		return nil, err
	}

	h.logger.Infof("httpClient.Get: get %s", url)

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	bytes, err := io.ReadAll(response.Body)

	if err != nil {
		h.logger.Error("io.ReadAll", err.Error())
		return nil, err
	}

	return bytes, nil
}

func (h ClientRequestHandler) GetRepositoryInfo(url string) *entity.Repository {
	bytes, err := h.get(url)

	if err != nil {
		return nil
	}

	var repository *entity.Repository

	err = sonic.Unmarshal(bytes, &repository)
	if err != nil {
		return nil
	}

	var novelSource []entity.Source

	for _, source := range repository.Sources {
		if source.Type == entity.SourceTypeNovel {
			continue
		}
		novelSource = append(novelSource, source)
	}

	return &entity.Repository{
		Metadata: repository.Metadata,
		Sources:  novelSource,
	}
}
