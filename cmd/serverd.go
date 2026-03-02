package cmd

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	govern "github.com/haipham22/govern/graceful"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	restHandler "golang-sample/internal/handler/rest"
	"golang-sample/pkg/config"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "serverd",
	Short: "Start production API server with govern integration",
	Long: `Start the production-ready API server with full govern stack integration.

Features:
  - govern/http: Managed HTTP server with configurable timeouts
  - govern/graceful: Graceful shutdown handling (SIGINT/SIGTERM)
  - govern/postgres: Database connection pooling
  - govern/config: Configuration management

Shutdown Sequence:
  1. Stop accepting new connections
  2. Wait for active requests to complete (configurable timeout)
  3. Close database connections
  4. Release resources

Example:
  $ serverd --port 8080 --shutdown_time 30`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		log := zap.S()

		port, err := cmd.Flags().GetInt64("port")
		if err != nil {
			return err
		}

		shutdownTime, err := cmd.Flags().GetInt64("shutdown_time")
		if err != nil {
			return err
		}

		// Load config at composition root
		cfg := config.ENV

		handler, cleanup, err := restHandler.New(log, port, cfg)
		if err != nil {
			return err
		}
		defer cleanup()

		// Create signal context for graceful shutdown
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		// Run server with govern graceful runner
		err = govern.Run(
			ctx,
			log,
			time.Duration(shutdownTime)*time.Second,
			handler,
		)

		return err
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().Int64("port", 8080, "API server port (default: 8080)")
	serverCmd.Flags().Int64("shutdown_time", 10, "Graceful shutdown timeout in seconds (default: 10)")
}
