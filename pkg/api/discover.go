package api

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/colinking/rpc/pkg/xpath"
	jtd "github.com/jsontypedef/json-typedef-go"
	"github.com/yosuke-furukawa/json5/encoding/json5"
)

func Discover(root string) (API, error) {
	api := API{
		Endpoints:   []Endpoint{},
		Definitions: []Definition{},
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relpath, err := xpath.NewRel(root, path)
		if err != nil {
			return err
		}

		name := d.Name()
		if name == "definitions.json5" {
			definitions, err := parseDefinitionFile(relpath)
			if err != nil {
				return err
			}
			api.Definitions = append(api.Definitions, definitions...)
		} else if strings.HasSuffix(name, ".json5") {
			endpoint, err := parseEndpointFile(relpath)
			if err != nil {
				return err
			}
			api.Endpoints = append(api.Endpoints, endpoint)
		}

		return nil
	})
	if err != nil {
		return API{}, fmt.Errorf("looking for definitions: %w", err)
	}

	return api, nil
}

func parseDefinitionFile(path xpath.Rel) ([]Definition, error) {
	contents, err := os.ReadFile(path.Abs())
	if err != nil {
		return nil, fmt.Errorf("reading definition file (%q): %w", path, err)
	}

	var file map[string]jtd.Schema
	if err := json5.Unmarshal(contents, &file); err != nil {
		return nil, fmt.Errorf("unmarshaling definition file (%q): %w", path, err)
	}

	definitions := []Definition{}
	for name, schema := range file {
		definitions = append(definitions, Definition{
			Name:   append(path.RelDirs(), name),
			Schema: transformSchema(schema),
		})
	}

	return definitions, nil
}

func parseEndpointFile(path xpath.Rel) (Endpoint, error) {
	endpoint := Endpoint{}

	components := strings.SplitN(path.FileName(), ".", 3)
	if len(components) != 3 {
		return Endpoint{}, fmt.Errorf("invalid endpoint file: expected <NAME>.<VERB>.json5: got %s", path.FileName())
	}
	if components[2] != "json5" {
		return Endpoint{}, fmt.Errorf("unsupported format: %q", components[2])
	}
	endpoint.Verb = strings.ToUpper(components[1])
	endpoint.Name = append(path.RelDirs(), components[0])

	contents, err := os.ReadFile(path.Abs())
	if err != nil {
		return Endpoint{}, fmt.Errorf("reading endpoint file (%s): %w", path, err)
	}

	var file struct {
		Request  jtd.Schema `json:"request"`
		Response jtd.Schema `json:"response"`
	}
	if err := json5.Unmarshal(contents, &file); err != nil {
		return Endpoint{}, fmt.Errorf("unmarshaling endpoint file (%s): %w", path, err)
	}

	endpoint.Request = transformSchema(file.Request)
	endpoint.Response = transformSchema(file.Response)

	// TODO: extract definitions from request/response schemas

	return endpoint, nil
}
