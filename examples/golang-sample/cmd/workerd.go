package cmd

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	govern "github.com/haipham22/govern/graceful"

	"github.com/haipham22/golang-sample/internal/handler/job"
	"github.com/haipham22/golang-sample/pkg/config"
)

// workerCmd starts the background worker: the cron scheduler (and, when wired,
// the asynq consumer) via govern graceful.Run.
var workerCmd = &cobra.Command{
	Use:   "workerd",
	Short: "Start background workers (cron scheduler + asynq consumer)",
	Long: `Start the background worker process.

Currently runs the sample cron cleanup job via govern/cron. The asynq message
consumer (internal/handler/message) requires a Redis connection; wire it into
the services slice below once APP_REDIS_URL is configured in production.

Example:
  $ workerd --shutdown_time 30`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		log := zap.S()

		shutdownTime, err := cmd.Flags().GetInt64("shutdown_time")
		if err != nil {
			return err
		}

		// Build the cron scheduler with the sample cleanup job registered.
		scheduler, err := job.New(log)
		if err != nil {
			return fmt.Errorf("workerd: %w", err)
		}
		defer scheduler.Cleanup()

		// Compose the graceful services. Add the asynq Server here once a Redis
		// client is wired (govern/database/redis.New(cfg.Redis.URL)).
		services := []govern.Service{scheduler}

		// Redis URL is optional in dev; only mention it when unset.
		if config.ENV != nil && config.ENV.Redis.URL == "" {
			log.Warn("APP_REDIS_URL not set; asynq consumer disabled")
		}

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		return govern.Run(
			ctx,
			log,
			time.Duration(shutdownTime)*time.Second,
			services...,
		)
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)
	workerCmd.Flags().Int64("shutdown_time", 10, "Graceful shutdown timeout in seconds (default: 10)")
}
