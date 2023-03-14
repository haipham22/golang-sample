package command

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	config2 "paperback-vbook-converter/config"
)

type CheckForUpdateSource struct {
	logger *zap.SugaredLogger
	config *config2.Config
}

func NewCheckForUpdateCmd(logger *zap.SugaredLogger, config *config2.Config) *CheckForUpdateSource {
	return &CheckForUpdateSource{
		logger: logger,
		config: config,
	}
}

func (c CheckForUpdateSource) Run(repositories []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	c.logger.Info("Start check for update....")

	c.logger.Debug(repositories)

	//if repositoriesUrl := c.config.RepositoryUrl; repositoriesUrl == "" {
	//	c.logger.Warn("Not found any repository, stop....")
	//}
	<-ctx.Done()
	return nil
}
