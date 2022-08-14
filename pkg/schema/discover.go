package schema

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	jtd "github.com/jsontypedef/json-typedef-go"
	"github.com/yosuke-furukawa/json5/encoding/json5"
)

type API struct {
	Endpoints   []Endpoint
	Definitions []Definition
}

type Definition struct {
	Path   []string
	Schema jtd.Schema
}

type Endpoint struct {
	Path     []string
	Verb     string
	Request  jtd.Schema
	Response jtd.Schema
}

type endpointFile struct {
	Request  jtd.Schema `json:"request"`
	Response jtd.Schema `json:"response"`
}

func Discover(root string) (API, error) {
	api := API{}

	definitions := []string{}
	endpoints := []string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		name := d.Name()
		if name == "definitions.json5" {
			definitions = append(definitions, path)
		} else if strings.HasSuffix(name, ".json5") {
			endpoints = append(endpoints, path)
		}

		return nil
	})
	if err != nil {
		return API{}, fmt.Errorf("looking for definitions: %w", err)
	}

	for _, path := range definitions {
		relpath, err := filepath.Rel(root, path)
		if err != nil {
			return API{}, fmt.Errorf("invalid path: %w", err)
		}

		dir, _ := filepath.Split(relpath)
		components := []string{}
		if len(dir) > 0 {
			components = strings.Split(filepath.Clean(dir), "/")
		}

		contents, err := os.ReadFile(path)
		if err != nil {
			return API{}, fmt.Errorf("reading definition file (%q): %w", path, err)
		}

		var file map[string]jtd.Schema
		if err := json5.Unmarshal(contents, &file); err != nil {
			return API{}, fmt.Errorf("unmarshaling definition file (%q): %w", path, err)
		}

		for name, schema := range file {
			api.Definitions = append(api.Definitions, Definition{
				Path:   append(components, name),
				Schema: schema,
			})
		}
	}

	for _, path := range endpoints {
		relpath, err := filepath.Rel(root, path)
		if err != nil {
			return API{}, fmt.Errorf("invalid path: %w", err)
		}

		endpoint := Endpoint{
			Path: []string{},
		}

		dir, fileName := filepath.Split(relpath)
		if len(dir) > 0 {
			components := strings.Split(filepath.Clean(dir), "/")
			endpoint.Path = append(endpoint.Path, components...)
		}

		components := strings.SplitN(fileName, ".", 3)
		if len(components) != 3 {
			return API{}, fmt.Errorf("invalid endpoint file: expected <NAME>.<VERB>.json5: got %s", fileName)
		}
		if components[2] != "json5" {
			return API{}, fmt.Errorf("unsupported format: %q", components[2])
		}
		endpoint.Verb = strings.ToUpper(components[1])
		endpoint.Path = append(endpoint.Path, components[0])

		contents, err := os.ReadFile(path)
		if err != nil {
			return API{}, fmt.Errorf("reading endpoint file (%q): %w", path, err)
		}

		var file endpointFile
		if err := json5.Unmarshal(contents, &file); err != nil {
			return API{}, fmt.Errorf("unmarshaling endpoint file (%q): %w", path, err)
		}

		endpoint.Request = transformSchema(file.Request)
		endpoint.Response = transformSchema(file.Response)

		api.Endpoints = append(api.Endpoints, endpoint)
	}

	return api, nil
}

func transformSchema(schema jtd.Schema) jtd.Schema {
	// Perform transformations
	switch schema.Form() {
	case jtd.FormProperties:
		// For backwards/forwards compatibility, always ignore unknown properties.
		schema.AdditionalProperties = true
	}

	// Recursively transform any sub-schemas
	for k, v := range schema.Definitions {
		schema.Definitions[k] = transformSchema(v)
	}

	switch schema.Form() {
	case jtd.FormEmpty, jtd.FormRef, jtd.FormType, jtd.FormEnum:
		// no-op
	case jtd.FormElements:
		*schema.Elements = transformSchema(*schema.Elements)
	case jtd.FormProperties:
		for k, v := range schema.Properties {
			schema.Properties[k] = transformSchema(v)
		}
		for k, v := range schema.OptionalProperties {
			schema.OptionalProperties[k] = transformSchema(v)
		}
	case jtd.FormValues:
		*schema.Values = transformSchema(*schema.Values)
	case jtd.FormDiscriminator:
		for k, v := range schema.Mapping {
			schema.Mapping[k] = transformSchema(v)
		}
	default:
		panic(fmt.Errorf("unexpected schema form: %s", schema.Form()))
	}

	return schema
}
