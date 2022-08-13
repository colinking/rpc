package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/airplanedev/trap"
	"github.com/colinking/rpc/pkg/schema"
)

func main() {
	ctx := trap.Context()
	if err := run(ctx, "./example/api", "./example/generated"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, apiPath string, clientPath string) error {
	api, err := schema.Discover(apiPath)
	if err != nil {
		return err
	}

	fmt.Printf("API:\n")
	for _, def := range api.Definitions {
		fmt.Printf("- [def] %s\n", strings.Join(def.Path, "."))
	}
	for _, endpoint := range api.Endpoints {
		fmt.Printf("- %-4s /%s\n", endpoint.Verb, strings.Join(endpoint.Path, "/"))
	}

	if err := os.RemoveAll(clientPath); err != nil {
		return fmt.Errorf("clearing generated client path: %w", err)
	}
	if err := os.Mkdir(clientPath, 0755); err != nil {
		return fmt.Errorf("creating generated client path: %w", err)
	}

	if err := schema.Generate(ctx, api, clientPath); err != nil {
		return err
	}

	return nil
}
