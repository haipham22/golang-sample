package command

import (
	"context"
	"os"
	"os/signal"
	"paperback-vbook-converter/internal/pkg/http"
	"syscall"

	"go.uber.org/zap"

	config2 "paperback-vbook-converter/config"
)

type ConvertAllCmd struct {
	logger *zap.SugaredLogger
	config config2.Config
}

func NewConvertAllCmd(logger *zap.SugaredLogger, config config2.Config) *ConvertAllCmd {
	return &ConvertAllCmd{
		logger: logger,
		config: config,
	}
}

func (c ConvertAllCmd) Run(repositories []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	<-ctx.Done()
	return nil
}

func (c ConvertAllCmd) execCmd(ctx context.Context, repository string, httpClient *http.ClientRequestHandler) {

}
