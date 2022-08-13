package schema

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path"
	"strings"

	jtd "github.com/jsontypedef/json-typedef-go"
)

func Generate(ctx context.Context, api API, dir string) error {
	definitions := map[string]jtd.Schema{}
	for _, def := range api.Definitions {
		definitions[strings.Join(def.Path, ".")] = def.Schema
	}

	result, err := generateSchema(ctx, path.Join(dir, "definitions.go"), "definitions", jtd.Schema{
		Definitions: definitions,
		Metadata: map[string]interface{}{
			"description": "Definitions is a no-op used for generation purposes.",
		},
	})
	if err != nil {
		return err
	}

	// Disable definition generation going forward.
	for k, v := range definitions {
		if v.Metadata == nil {
			v.Metadata = map[string]interface{}{}
		}
		v.Metadata["goType"] = result.DefinitionNames[k]
		definitions[k] = v
	}

	for _, endpoint := range api.Endpoints {
		name := strings.Join(endpoint.Path, ".")
		endpoint.Request.Definitions = definitions
		if _, err := generateSchema(ctx, path.Join(dir, name+".request.go"), name+".request.", endpoint.Request); err != nil {
			return err
		}

		endpoint.Response.Definitions = definitions
		if _, err := generateSchema(ctx, path.Join(dir, name+".response.go"), name+".response.", endpoint.Response); err != nil {
			return err
		}
	}

	return nil
}

type generationResult struct {
	DefinitionNames map[string]string `json:"definition_names"`
}

func generateSchema(ctx context.Context, file string, name string, schema jtd.Schema) (generationResult, error) {
	content, err := json.MarshalIndent(toSerializableSchema(schema), "", "\t")
	if err != nil {
		return generationResult{}, fmt.Errorf("marshaling schema: %w", err)
	}
	fmt.Fprintln(os.Stderr, string(content))

	tmpDir, err := os.MkdirTemp("", "rpc-*")
	if err != nil {
		return generationResult{}, fmt.Errorf("creating temporary directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error while cleaning up temporary directory (%q): %+v\n", tmpDir, err)
		}
	}()

	// TODO: write to temporary directory to ensure atomic
	// Requires the JTD CLI: https://github.com/jsontypedef/json-typedef-codegen
	pkgName := "generated"
	cmd := exec.CommandContext(ctx, "jtd-codegen", "-", "--go-out", tmpDir, "--go-package", pkgName, "--root-name", name, "--log-format", "json")
	cmd.Stdin = bytes.NewBuffer(content)
	// RUST_BACKTRACE helps debug jtd-codegen issues.
	cmd.Env = append(cmd.Env, "RUST_BACKTRACE=1")
	out, err := cmd.CombinedOutput()
	fmt.Fprintln(os.Stderr, string(out))
	if err != nil {
		return generationResult{}, fmt.Errorf("running jtd-codegen: %w", err)
	}

	var result struct {
		Go generationResult `json:"go"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return generationResult{}, fmt.Errorf("parsing jtd-codegen output: %w", err)
	}

	contents, err := os.ReadFile(path.Join(tmpDir, pkgName+".go"))
	if err != nil {
		return generationResult{}, fmt.Errorf("reading generated code: %w", err)
	}

	// Ensure the generated code is gofmt-ed:
	formattedContents, err := format.Source(contents)
	if err != nil {
		return generationResult{}, fmt.Errorf("formatting generated code: %w", err)
	}
	if err := os.WriteFile(file, formattedContents, 0755); err != nil {
		return generationResult{}, fmt.Errorf("writing formatted generated code: %w", err)
	}

	return result.Go, nil
}
