package config

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Worker string `mapstructure:"WORKER_PROCESS" validate:"required"`
}

func NewConfig(validator *validator.Validate) (*Config, error) {
	config := &Config{
		Worker: viper.GetString("WORKER_PROCESS"),
	}
	if err := validator.StructCtx(context.Background(), config); err != nil {
		return nil, err
	}
	return config, nil
}
