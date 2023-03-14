package command

import (
	"context"
	"go.uber.org/zap"
	"os"
	"os/signal"
	config2 "paperback-vbook-converter/config"
	"syscall"
)

type Command func()

type CommandHandler interface {
	CreateNewCommandHandler(logger *zap.SugaredLogger, config config2.Config)
}

type DefaultCommandHandler struct {
	logger *zap.SugaredLogger
	config config2.Config
}

func (c DefaultCommandHandler) Run(repositories []string, command *CommandHandler) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	<-ctx.Done()
	return nil
}

func (c DefaultCommandHandler) execCmd(cmd Command) {

}
