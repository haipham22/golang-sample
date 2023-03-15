package config

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Worker worker `validate:"required"`
}

type worker struct {
	MaxProcess int64 `mapstructure:"MAX_WORKER" validate:"required" json:"max_process,omitempty"`
}

func NewConfig(validator *validator.Validate) (*Config, error) {
	config := &Config{
		Worker: worker{
			MaxProcess: viper.GetInt64("MAX_WORKER"),
		},
	}
	if err := validator.StructCtx(context.Background(), config); err != nil {
		return nil, err
	}
	return config, nil
}
