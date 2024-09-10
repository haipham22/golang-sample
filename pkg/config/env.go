package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// EnvConfigMap define mapping struct field and environment field
type EnvConfigMap struct {
	APP struct {
		DEBUG bool   `mapstructure:"DEBUG" validate:"required"`
		ENV   string `mapstructure:"ENV" validate:"required"`
	} `mapstructure:"APP"`
}

// ENV is global variable for using config in other place
var ENV EnvConfigMap

// LoadConfig read env file and loaded to environment and global ENV variable
func LoadConfig(cfgFile string) error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigFile(".env")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err == nil {
		_, _ = fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		return err
	}

	err = viper.Unmarshal(&ENV)
	if err != nil {
		return err
	}

	err = validator.New().Struct(ENV)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Validate error: %v", err)
		return err
	}

	return nil
}
