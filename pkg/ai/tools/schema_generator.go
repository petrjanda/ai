package tools

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
)

var DefaultSchemaGenerator SchemaGenerator = NewGenericSchemaGenerator()

func SetDefaultSchemaGenerator(generator SchemaGenerator) {
	DefaultSchemaGenerator = generator
}

type SchemaGenerator interface {
	Generate(v interface{}) (json.RawMessage, error)
	MustGenerate(v interface{}) json.RawMessage
}

type GenericSchemaGenerator struct {
	reflector *jsonschema.Reflector
}

func NewGenericSchemaGenerator() *GenericSchemaGenerator {
	return &GenericSchemaGenerator{
		reflector: &jsonschema.Reflector{
			ExpandedStruct:             true,
			DoNotReference:             true,
			RequiredFromJSONSchemaTags: true,
			AllowAdditionalProperties:  true,
		},
	}
}

func (g *GenericSchemaGenerator) Generate(v interface{}) (json.RawMessage, error) {
	schema := g.reflector.Reflect(v)

	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	return json.RawMessage(schemaBytes), nil
}

func (g *GenericSchemaGenerator) MustGenerate(v interface{}) json.RawMessage {
	if schema, err := g.Generate(v); err != nil {
		panic(err)
	} else {
		return schema
	}
}
