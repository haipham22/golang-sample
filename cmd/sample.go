package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"paperback-vbook-converter/pkg/sample"
)

// sampleCmd represents the binance command
var sampleCmd = &cobra.Command{
	Use:   "sample",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

This is sample command.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := zap.S()
		obj := sample.ExampleType{
			Logger: logger,
		}
		obj.Run()
	},
}

func init() {
	rootCmd.AddCommand(sampleCmd)
}
