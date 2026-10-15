package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/s0cks/prose2json/cmd"
)

func main() {
	code := run()
	os.Exit(code)
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err, stopModel := cmd.Exec(ctx)
	if stopModel != nil {
		defer func() {
			if stopErr := stopModel(); stopErr != nil {
				fmt.Fprintln(os.Stderr, "failed to stop model:", stopErr)
			}
		}()
	}

	if err != nil {
		if ctx.Err() != nil || (errors.Is(err, context.Canceled)) { // && pipeline.IsCanceled(err)
			return 130
		}

		fmt.Fprintln(os.Stderr, "error:", err)
		return 1
	}

	return 0
}
