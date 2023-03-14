package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	repositories []string
)

var rootCmd = &cobra.Command{
	Use:   "paperback-utils",
	Short: "paperback convert utils",
	//Run: func(cmd *cobra.Command, args []string) {
	//	cmd.Flags().StringSliceVarP(&slice, "repo", "s", []string{}, "")
	//},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLog, initConfig)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.PersistentFlags().StringSliceVarP(&repositories, "repo", "r", []string{}, "repository")
}

func initConfig() {
	// read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
	viper.SetDefault("WORKER_PROCESS", 5)
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
