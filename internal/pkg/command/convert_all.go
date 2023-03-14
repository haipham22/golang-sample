package command

import (
	"context"
	"go.uber.org/zap"
	"os/signal"
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
	signal.NotifyContext(context.Background())
	return nil
}
