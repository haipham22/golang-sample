package cmd

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"paperback-vbook-converter/internal/pkg/command"

	config2 "paperback-vbook-converter/config"
)

// convertSourceCmd represents the binance command
var convertSourceCmd = &cobra.Command{
	Use:   "convert-all",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

This is sample command.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := zap.S()

		validate := validator.New()
		config, err := config2.NewConfig(validate)

		if err != nil {
			logger.Warn("config2.NewConfig", err.Error())
			return //Stop func when repository empty
		}

		obj := command.NewCheckForUpdateCmd(logger, config)

		logger.Fatal(obj.Run(repositories))
	},
}

func init() {
	rootCmd.AddCommand(convertSourceCmd)
}
