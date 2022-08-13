package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/colinking/rpc/pkg/schema"
)

func main() {
	if err := run("./example/api"); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func run(path string) error {
	api, err := schema.Discover(path)
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

	return nil
}
