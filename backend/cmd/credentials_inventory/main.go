package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/omni/backend/internal/services"
)

func main() {
	jsonOutput := flag.Bool("json", false, "write redacted JSON instead of text")
	flag.Parse()

	ctx := context.Background()
	basePath := os.Getenv("CONFIG_PATH")
	if basePath == "" {
		basePath = "/app/config"
	}

	svc := services.NewCredentialsInventoryService(basePath)
	report, err := svc.Run(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	var writeErr error
	if *jsonOutput {
		writeErr = services.WriteInventoryReport(os.Stdout, report)
	} else {
		writeErr = services.WriteInventoryTextReport(os.Stdout, report)
	}
	if writeErr != nil {
		fmt.Fprintln(os.Stderr, writeErr.Error())
		os.Exit(1)
	}
}
