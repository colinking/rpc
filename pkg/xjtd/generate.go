package xjtd

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

const GoPackageName = "generated"

type GenerateSchemaResult struct {
	RootName        string            `json:"root_name"`
	DefinitionNames map[string]string `json:"definition_names"`
}

func GenerateSchema(ctx context.Context, file string, name string, schema jtd.Schema, externalDefinitions []string) (GenerateSchemaResult, error) {
	content, err := json.MarshalIndent(NewSerializableSchema(schema), "", "\t")
	if err != nil {
		return GenerateSchemaResult{}, fmt.Errorf("marshaling schema: %w", err)
	}
	fmt.Fprintln(os.Stderr, string(content))

	tmpDir, err := os.MkdirTemp("", "rpc-*")
	if err != nil {
		return GenerateSchemaResult{}, fmt.Errorf("creating temporary directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error while cleaning up temporary directory (%q): %+v\n", tmpDir, err)
		}
	}()

	// Requires the JTD CLI: https://github.com/jsontypedef/json-typedef-codegen
	cmd := exec.CommandContext(ctx, "jtd-codegen", "-", "--go-out", tmpDir, "--go-package", GoPackageName, "--root-name", name, "--log-format", "json")
	cmd.Stdin = bytes.NewBuffer(content)
	// RUST_BACKTRACE helps debug jtd-codegen issues.
	cmd.Env = append(cmd.Env, "RUST_BACKTRACE=1")
	out, err := cmd.CombinedOutput()
	fmt.Fprintln(os.Stderr, string(out))
	if err != nil {
		return GenerateSchemaResult{}, fmt.Errorf("running jtd-codegen: %w", err)
	}

	var result struct {
		Go GenerateSchemaResult `json:"go"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return GenerateSchemaResult{}, fmt.Errorf("parsing jtd-codegen output: %w", err)
	}

	contents, err := os.ReadFile(path.Join(tmpDir, GoPackageName+".go"))
	if err != nil {
		return GenerateSchemaResult{}, fmt.Errorf("reading generated code: %w", err)
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
		return GenerateSchemaResult{}, fmt.Errorf("formatting generated code: %w", err)
	}
	if err := os.WriteFile(file, formattedContents, 0755); err != nil {
		return GenerateSchemaResult{}, fmt.Errorf("writing formatted generated code: %w", err)
	}

	return result.Go, nil
}
