package xjtd

import jtd "github.com/jsontypedef/json-typedef-go"

// SerializableSchema is a jtd.Schema with omitempty tags.
// Without those tags, the marshaled schema is invalid according to jtd-codegen.
type SerializableSchema struct {
	Definitions          map[string]SerializableSchema `json:"definitions,omitempty"`
	Metadata             map[string]interface{}        `json:"metadata,omitempty"`
	Nullable             bool                          `json:"nullable,omitempty"`
	Ref                  *string                       `json:"ref,omitempty"`
	Type                 string                        `json:"type,omitempty"`
	Enum                 []string                      `json:"enum,omitempty"`
	Elements             *SerializableSchema           `json:"elements,omitempty"`
	Properties           map[string]SerializableSchema `json:"properties,omitempty"`
	OptionalProperties   map[string]SerializableSchema `json:"optionalProperties,omitempty"`
	AdditionalProperties bool                          `json:"additionalProperties,omitempty"`
	Values               *SerializableSchema           `json:"values,omitempty"`
	Discriminator        string                        `json:"discriminator,omitempty"`
	Mapping              map[string]SerializableSchema `json:"mapping,omitempty"`
}

func NewSerializableSchema(schema jtd.Schema) SerializableSchema {
	return SerializableSchema{
		Definitions:          NewSerializableMapSchema(schema.Definitions),
		Metadata:             schema.Metadata,
		Nullable:             schema.Nullable,
		Ref:                  schema.Ref,
		Type:                 string(schema.Type),
		Enum:                 schema.Enum,
		Elements:             NewSerializableSchemaPointer(schema.Elements),
		Properties:           NewSerializableMapSchema(schema.Properties),
		OptionalProperties:   NewSerializableMapSchema(schema.OptionalProperties),
		AdditionalProperties: schema.AdditionalProperties,
		Values:               NewSerializableSchemaPointer(schema.Values),
		Discriminator:        schema.Discriminator,
		Mapping:              NewSerializableMapSchema(schema.Mapping),
	}
}

func NewSerializableMapSchema(schema map[string]jtd.Schema) map[string]SerializableSchema {
	if schema == nil {
		return nil
	}
	out := map[string]SerializableSchema{}
	for k, v := range schema {
		out[k] = NewSerializableSchema(v)
	}
	return out
}

func NewSerializableSchemaPointer(schema *jtd.Schema) *SerializableSchema {
	if schema == nil {
		return nil
	}
	out := NewSerializableSchema(*schema)
	return &out
}
