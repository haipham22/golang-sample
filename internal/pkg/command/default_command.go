package command

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"os"
	"os/signal"
	config2 "paperback-vbook-converter/config"
	"syscall"
)

type DefaultCommand interface {
	RegisterNewCommandHandler(ctx context.Context)
	DefaultCommandHandler
}

type DefaultCommandHandler struct {
	logger *zap.SugaredLogger
	config config2.Config
}

func (c DefaultCommandHandler) Register(repository []string) ([]string, error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if len(repository) <= 0 {
		return nil, errors.New("empty repository url")
	}

	<-ctx.Done()
	return repository, nil
}
