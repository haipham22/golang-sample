package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"golang-sample/pkg/config"
)

var rootCmd = &cobra.Command{
	Use:   "golang-sample",
	Short: "Sample Golang application with best practices",
}

var (
	cfgFile string
)

// Execute root execute function
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initDependency)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".env", "config file (default is .env)")
}

func initDependency() {
	initLog()
	initConfig()
}

func initConfig() {
	logger := zap.L()
	if _, err := config.LoadConfig(cfgFile, logger); err != nil {
		logger.Error("Failed to load config", zap.String("file", cfgFile), zap.Error(err))
		panic(fmt.Sprintf("can't load config from %s: %v", cfgFile, err))
	}
}

func initLog() {
	logger, _ := zap.NewDevelopment()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			fmt.Println(err)
		}
	}(logger)
	zap.ReplaceGlobals(logger)
}
