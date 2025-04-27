package cmd

import (
	"fmt"
	"os"

	"ebookgen/pkg/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:   "ebookgen",
	Short: "Tool for generate ebooks",
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".env", "config file (default is $APPLICATION_DIR/.env)")
}

func initDependency() {
	initConfig()
	initLog()
}

func initConfig() {
	if err := config.LoadConfig(cfgFile); err != nil {
		panic("Can't load config from environment")
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
