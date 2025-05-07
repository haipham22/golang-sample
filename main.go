package main

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/getsentry/sentry-go"
	_ "go.uber.org/automaxprocs"

	"golang-sample/cmd"
)

func main() {
	defer sentry.Flush(time.Second * 2)
	defer func() {
		// manually capture panic so we can do our own logging
		r := recover()
		if r != nil {
			fmt.Println("------------------", r, string(debug.Stack()))
			defer sentry.Recover()
			panic(r)
		}
	}()

	cmd.Execute()
}
