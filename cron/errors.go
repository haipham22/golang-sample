package cron

import "errors"

var (
	// ErrSchedulerAlreadyStarted is returned when Start is called on an already started scheduler
	ErrSchedulerAlreadyStarted = errors.New("scheduler already started")

	// ErrSchedulerNotStarted is returned when Shutdown is called on a scheduler that hasn't been started
	ErrSchedulerNotStarted = errors.New("scheduler not started")

	// ErrSchedulerAlreadyShutdown is returned when Shutdown is called on an already shutdown scheduler
	ErrSchedulerAlreadyShutdown = errors.New("scheduler already shutdown")
)
