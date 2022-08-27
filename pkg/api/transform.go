package api

import (
	"fmt"

	jtd "github.com/jsontypedef/json-typedef-go"
)

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
