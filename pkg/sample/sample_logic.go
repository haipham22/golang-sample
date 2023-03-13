package sample

import "go.uber.org/zap"

type ExampleType struct {
	Logger *zap.SugaredLogger
}

func (c *ExampleType) Run() {
	logger := c.Logger
	logger.Info("Sample func called")
}
