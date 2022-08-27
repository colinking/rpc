package golang

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"go/format"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/colinking/rpc/pkg/api"
	"github.com/colinking/rpc/pkg/xjtd"
	jtd "github.com/jsontypedef/json-typedef-go"
)

func Server(ctx context.Context, api api.API, dir string) error {
	definitionNames, err := generateDefinitions(ctx, api.Definitions, path.Join(dir, "definitions.go"))
	if err != nil {
		return err
	}

	var routes []route
	for _, endpoint := range api.Endpoints {
		route, err := generateEndpoint(ctx, definitionNames, endpoint, dir)
		if err != nil {
			return err
		}
		routes = append(routes, route)
	}

	if err := generateHandler(ctx, path.Join(dir, "handler.go"), routes); err != nil {
		return err
	}

	return nil
}

func generateDefinitions(ctx context.Context, definitions []api.Definition, path string) (map[string]string, error) {
	schema := jtd.Schema{
		Definitions: map[string]jtd.Schema{},
		Metadata: map[string]interface{}{
			// HACK: not sure how to generate the definitions without having to also generate
			// another schema.
			"description": "Definitions is a no-op used for generation purposes.",
		},
	}
	for _, def := range definitions {
		name := strings.Join(def.Name, ".")
		schema.Definitions[name] = def.Schema
	}

	result, err := xjtd.GenerateSchema(ctx, path, "definitions", schema, []string{})
	if err != nil {
		return nil, err
	}

	return result.DefinitionNames, nil
}

func generateEndpoint(ctx context.Context, definitionNames map[string]string, endpoint api.Endpoint, dir string) (route, error) {
	// HACK: `metadata.goType` doesn't seem to work with top-level schemas.
	// We don't want to generate definition types in each file.
	// To workaround this, we codegen each as an "any" type which is always one line
	// and then look for and remove those lines from the generated code.
	definitionsEmpty := map[string]jtd.Schema{}
	externalDefinitions := []string{}
	for k, v := range definitionNames {
		definitionsEmpty[k] = jtd.Schema{}
		externalDefinitions = append(externalDefinitions, v)
	}

	name := strings.Join(endpoint.Name, ".")
	requestDefinitions := endpoint.Request.Definitions
	endpoint.Request.Definitions = definitionsEmpty
	genReq, err := xjtd.GenerateSchema(ctx, path.Join(dir, name+".request.go"), name+".request.", endpoint.Request, externalDefinitions)
	if err != nil {
		return route{}, err
	}
	endpoint.Request.Definitions = requestDefinitions

	responseDefinitions := endpoint.Response.Definitions
	endpoint.Response.Definitions = definitionsEmpty
	genResp, err := xjtd.GenerateSchema(ctx, path.Join(dir, name+".response.go"), name+".response.", endpoint.Response, externalDefinitions)
	if err != nil {
		return route{}, err
	}
	endpoint.Response.Definitions = responseDefinitions

	handlerName := strings.TrimSuffix(genReq.RootName, "Request")

	if err := generateSchemas(ctx, path.Join(dir, name+".schemas.go"), handlerName, endpoint); err != nil {
		return route{}, err
	}

	return route{
		Path:         "/" + strings.Join(endpoint.Name, "/"),
		Verb:         endpoint.Verb,
		HandlerName:  handlerName,
		RequestType:  genReq.RootName,
		ResponseType: genResp.RootName,
	}, nil
}

type route struct {
	Path         string
	Verb         string
	HandlerName  string
	RequestType  string
	ResponseType string
}

//go:embed handler.go.tmpl
var handlerTemplate string

func generateHandler(ctx context.Context, file string, routes []route) error {
	t, err := template.New("handler").Parse(handlerTemplate)
	if err != nil {
		return fmt.Errorf("parsing routes template: %w", err)
	}

	data := struct {
		PackageName string
		Routes      []route
	}{
		PackageName: xjtd.GoPackageName,
		Routes:      routes,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Errorf("evaluating template: %w", err)
	}

	// Ensure the generated code is gofmt-ed:
	formattedContents, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting handler code: %w", err)
	}
	if err := os.WriteFile(file, formattedContents, 0755); err != nil {
		return fmt.Errorf("writing formatted handler code: %w", err)
	}

	return nil
}

//go:embed schemas.go.tmpl
var schemasTemplate string

func generateSchemas(ctx context.Context, file string, name string, endpoint api.Endpoint) error {
	t, err := template.New("schemas").Parse(schemasTemplate)
	if err != nil {
		return fmt.Errorf("parsing schemas template: %w", err)
	}

	data := struct {
		PackageName    string
		Name           string
		RequestSchema  string
		ResponseSchema string
	}{
		PackageName: xjtd.GoPackageName,
		Name:        name,
	}

	req, err := json.Marshal(xjtd.NewSerializableSchema(endpoint.Request))
	if err != nil {
		return fmt.Errorf("marshaling request schema: %w", err)
	}
	data.RequestSchema = string(req)

	resp, err := json.Marshal(xjtd.NewSerializableSchema(endpoint.Response))
	if err != nil {
		return fmt.Errorf("marshaling response schema: %w", err)
	}
	data.ResponseSchema = string(resp)

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Errorf("evaluating template: %w", err)
	}

	// Ensure the generated code is gofmt-ed:
	formattedContents, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting schemas code: %w", err)
	}
	if err := os.WriteFile(file, formattedContents, 0755); err != nil {
		return fmt.Errorf("writing formatted schemas code: %w", err)
	}

	return nil
}
