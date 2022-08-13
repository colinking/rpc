package schema

import jtd "github.com/jsontypedef/json-typedef-go"

// schemaOmitEmpty is a jtd.Schema with omitempty tags.
// Without those tags, the marshaled schema is invalid according to jtd-codegen.
type schemaOmitEmpty struct {
	Definitions          map[string]schemaOmitEmpty `json:"definitions,omitempty"`
	Metadata             map[string]interface{}     `json:"metadata,omitempty"`
	Nullable             bool                       `json:"nullable,omitempty"`
	Ref                  *string                    `json:"ref,omitempty"`
	Type                 string                     `json:"type,omitempty"`
	Enum                 []string                   `json:"enum,omitempty"`
	Elements             *schemaOmitEmpty           `json:"elements,omitempty"`
	Properties           map[string]schemaOmitEmpty `json:"properties,omitempty"`
	OptionalProperties   map[string]schemaOmitEmpty `json:"optionalProperties,omitempty"`
	AdditionalProperties bool                       `json:"additionalProperties,omitempty"`
	Values               *schemaOmitEmpty           `json:"values,omitempty"`
	Discriminator        string                     `json:"discriminator,omitempty"`
	Mapping              map[string]schemaOmitEmpty `json:"mapping,omitempty"`
}

func toSerializableSchema(schema jtd.Schema) schemaOmitEmpty {
	return schemaOmitEmpty{
		Definitions:          toSerializableMapSchema(schema.Definitions),
		Metadata:             schema.Metadata,
		Nullable:             schema.Nullable,
		Ref:                  schema.Ref,
		Type:                 string(schema.Type),
		Enum:                 schema.Enum,
		Elements:             toSerializableSchemaPointer(schema.Elements),
		Properties:           toSerializableMapSchema(schema.Properties),
		OptionalProperties:   toSerializableMapSchema(schema.OptionalProperties),
		AdditionalProperties: schema.AdditionalProperties,
		Values:               toSerializableSchemaPointer(schema.Values),
		Discriminator:        schema.Discriminator,
		Mapping:              toSerializableMapSchema(schema.Mapping),
	}
}

func toSerializableMapSchema(schema map[string]jtd.Schema) map[string]schemaOmitEmpty {
	if schema == nil {
		return nil
	}
	out := map[string]schemaOmitEmpty{}
	for k, v := range schema {
		out[k] = toSerializableSchema(v)
	}
	return out
}

func toSerializableSchemaPointer(schema *jtd.Schema) *schemaOmitEmpty {
	if schema == nil {
		return nil
	}
	out := toSerializableSchema(*schema)
	return &out
}
