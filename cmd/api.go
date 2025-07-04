package cmd

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"golang-sample/internal/api"
	"golang-sample/pkg/config"
)

// apiServerCmd represents the binance command
var apiServerCmd = &cobra.Command{
	Use:   "api",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

This is sample command.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		log := zap.S()

		shutdownTime, err := cmd.Flags().GetInt64("shutdown_time")
		if err != nil {
			shutdownTime = 10
		}

		port, err := cmd.Flags().GetInt64("port")
		if err != nil {
			log.Fatal("Get port config error")
		}

		apiHandler, cleanup, err := api.InitApp(config.ENV.APP.DEBUG, config.ENV.Postgres.DSN, log)
		if err != nil {
			cleanup()
			log.Fatalf("Could not initialize api handler: %v", err)
		}
		defer cleanup()

		serverFunc, err := apiHandler.CreateServer(port)
		go func() {
			log.Info("Server started on port :", port)
			if err = serverFunc.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("shutting down the server. Err: %v", err)
			}
		}()

		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(shutdownTime)*time.Second)
		defer cancel()
		if err := serverFunc.Shutdown(shutdownCtx); err != nil {
			log.Fatal(err)
		}
		log.Info("Server gracefully stopped")
	},
}

func init() {
	rootCmd.AddCommand(apiServerCmd)

	apiServerCmd.Flags().Int64("port", 8080, "API port listening")
	apiServerCmd.Flags().Int64("shutdown_time", 10, "Gracefully shutdown time")
}
