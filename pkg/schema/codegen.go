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
		name := strings.Join(def.Path, ".")
		definitions[name] = def.Schema
	}

	result, err := generateSchema(ctx, path.Join(dir, "definitions.go"), "definitions", jtd.Schema{
		Definitions: definitions,
		Metadata: map[string]interface{}{
			// HACK: not sure how to generate the definitions without having to also generate
			// another schema.
			"description": "Definitions is a no-op used for generation purposes.",
		},
	}, []string{})
	if err != nil {
		return err
	}

	// HACK: `metadata.goType` doesn't seem to work with top-level schemas.
	// We don't want to generate definition types in each file.
	// To workaround this, we codegen each as an "any" type which is always one line
	// and then look for and remove those lines from the generated code.
	externalDefinitions := []string{}
	for k := range definitions {
		definitions[k] = jtd.Schema{}
		externalDefinitions = append(externalDefinitions, result.DefinitionNames[k])
	}

	for _, endpoint := range api.Endpoints {
		name := strings.Join(endpoint.Path, ".")
		endpoint.Request.Definitions = definitions
		if _, err := generateSchema(ctx, path.Join(dir, name+".request.go"), name+".request.", endpoint.Request, externalDefinitions); err != nil {
			return err
		}

		endpoint.Response.Definitions = definitions
		if _, err := generateSchema(ctx, path.Join(dir, name+".response.go"), name+".response.", endpoint.Response, externalDefinitions); err != nil {
			return err
		}
	}

	return nil
}

type generationResult struct {
	DefinitionNames map[string]string `json:"definition_names"`
}

func generateSchema(ctx context.Context, file string, name string, schema jtd.Schema, externalDefinitions []string) (generationResult, error) {
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

	// HACK: remove external definitions. See HACK comment above.
	originalLines := strings.Split(string(contents), "\n")
	lines := []string{}
	for _, line := range originalLines {
		include := true
		for _, def := range externalDefinitions {
			if strings.HasPrefix(line, "type "+def) {
				include = false
				break
			}
		}
		if include {
			lines = append(lines, line)
		}
	}
	contents = []byte(strings.Join(lines, "\n"))

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
