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

	grpchandler "github.com/haipham22/golang-sample/internal/handler/grpc"
)

// grpcCmd starts the sample gRPC server.
var grpcCmd = &cobra.Command{
	Use:   "grpcd",
	Short: "Start the sample gRPC server (Greeter)",
	Long: `Start the sample gRPC server.

This is a minimal template: govern does not ship a gRPC helper, so this command
uses google.golang.org/grpc directly with a hand-written Greeter service. In a
real project, generate stubs with protoc-gen-go-grpc and register them in
internal/handler/grpc.

Example:
  $ grpcd --addr :9090`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		log := zap.S()

		addr, err := cmd.Flags().GetString("addr")
		if err != nil {
			return err
		}
		shutdownTime, err := cmd.Flags().GetInt64("shutdown_time")
		if err != nil {
			return err
		}

		greeter := grpchandler.NewGreeter(log)
		server, err := grpchandler.New(log, addr, greeter)
		if err != nil {
			return fmt.Errorf("grpcd: %w", err)
		}

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		return govern.Run(
			ctx,
			log,
			time.Duration(shutdownTime)*time.Second,
			server,
		)
	},
}

func init() {
	rootCmd.AddCommand(grpcCmd)
	grpcCmd.Flags().String("addr", ":9090", "gRPC listen address (default: :9090)")
	grpcCmd.Flags().Int64("shutdown_time", 10, "Graceful shutdown timeout in seconds (default: 10)")
}
