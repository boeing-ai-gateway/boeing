package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	_ "time/tzdata"

	"github.com/boeing-ai-gateway/cmd"
	"github.com/boeing-ai-gateway/boeingbot/pkg/supervise"
	"github.com/boeing-ai-gateway/boeing/pkg/cli"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "_exec" {
		if err := supervise.Daemon(); err != nil {
			fmt.Printf("failed to run boeingbot daemon: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	// Don't shutdown on SIGTERM, only on SIGINT. SIGTERM is handled by the controller leader election
	cmd.ShutdownSignals = []os.Signal{os.Interrupt}
	root := cli.New()
	if err := root.ExecuteContext(cmd.SetupSignalContext()); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			os.Exit(1)
		}
		if cli.ErrorAlreadyReported(err) {
			os.Exit(1)
		}
		log.Fatal(err)
	}
}
