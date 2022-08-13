package schema

import (
	"github.com/yosuke-furukawa/json5/encoding/json5"
	jtd "github.com/jsontypedef/json-typedef-go"
)

type API struct {
	Endpoints []Endpoint
}

type Endpoint struct {
	Path []string
	Verb string
	Request jtd.Schema
	Response jtd.Schema
}

func Discover(path string) (API, error) {
	api := API{}

	definitions := []string{}
	endpoints := []string{}
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, "definitions.json5") {
			definitions = append(definitions, path)
		} else if strings.HasSuffix(path, ".json5") {
			endpoints = append(definitions, path)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("looking for definitions: %w", err)
	}

	for _, ep := range endpoints {
		api.Endpoints = append(api.Endpoints, Endpoint{
			Path: []string{ep},Ø
		})
	}

	return api, nil
}
