package openai

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/invopop/jsonschema"
)

// OpenAISchemaGenerator generates JSON schemas compatible with OpenAI's tool system
type OpenAISchemaGenerator struct {
	reflector *jsonschema.Reflector
}

// NewOpenAISchemaGenerator creates a new OpenAI-compatible schema generator
func NewOpenAISchemaGenerator() *OpenAISchemaGenerator {
	reflector := &jsonschema.Reflector{
		// OpenAI supports JSON Schema Draft 2020-12
		// Note: We can't set SchemaID directly, but the library uses 2020-12 by default

		// OpenAI doesn't support $defs, so we expand all structs
		ExpandedStruct: true,

		// Don't use references since OpenAI doesn't support them well
		DoNotReference: true,

		// Use required tags for validation
		RequiredFromJSONSchemaTags: true,

		// OpenAI doesn't support additionalProperties well, so we allow them
		AllowAdditionalProperties: true,
	}

	if err := reflector.AddGoComments("github.com/getsynq/ai/pkg", "./"); err != nil {
		// deal with error
		panic(err)
	}

	return &OpenAISchemaGenerator{
		reflector: reflector,
	}
}

// GenerateSchema generates a JSON schema from a Go struct that's compatible with OpenAI tools
func (g *OpenAISchemaGenerator) Generate(v interface{}) (json.RawMessage, error) {
	if v == nil {
		return nil, fmt.Errorf("cannot generate schema from nil value")
	}

	// Get the type of the value
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Ensure we're working with a struct
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("can only generate schemas from structs, got %s", t.Kind())
	}

	// Generate the schema
	schema := g.reflector.Reflect(v)

	// Post-process the schema to ensure OpenAI compatibility
	g.postProcessSchema(schema)

	// Convert to JSON
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal schema: %w", err)
	}

	// Write schema to schema.json for debugging/reference
	if err := os.WriteFile("schema.json", schemaBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write schema file: %w", err)
	}

	return json.RawMessage(schemaBytes), nil
}

func (g *OpenAISchemaGenerator) MustGenerate(v interface{}) json.RawMessage {
	schema, err := g.Generate(v)
	if err != nil {
		panic(err)
	}

	return schema
}

// postProcessSchema ensures the schema is compatible with OpenAI's tool system
func (g *OpenAISchemaGenerator) postProcessSchema(schema *jsonschema.Schema) {
	if schema == nil {
		return
	}

	// OpenAI doesn't support additionalProperties well, so remove it
	schema.AdditionalProperties = nil

	// OpenAI doesn't support $defs, so we need to ensure all types are expanded
	// This is handled by ExpandedStruct: true in the reflector

	// Recursively process nested schemas
	if schema.Properties != nil {
		// Properties is an OrderedMap, so we need to iterate differently
		// For now, we'll skip this as the main compatibility issues are at the top level
	}

	// Process array items
	if schema.Items != nil {
		g.postProcessSchema(schema.Items)
	}

	// Process oneOf/anyOf/allOf
	if schema.OneOf != nil {
		for _, s := range schema.OneOf {
			g.postProcessSchema(s)
		}
	}
	if schema.AnyOf != nil {
		for _, s := range schema.AnyOf {
			g.postProcessSchema(s)
		}
	}
	if schema.AllOf != nil {
		for _, s := range schema.AllOf {
			g.postProcessSchema(s)
		}
	}
}

// GenerateSchemaFromType generates a schema from a Go type (not an instance)
func (g *OpenAISchemaGenerator) GenerateSchemaFromType(t reflect.Type) (json.RawMessage, error) {
	if t == nil {
		return nil, fmt.Errorf("cannot generate schema from nil type")
	}

	// Create a zero value of the type
	zeroValue := reflect.New(t).Elem().Interface()
	return g.Generate(zeroValue)
}

// GenerateSchemaFromStruct is a convenience function that generates a schema from a struct type
func GenerateSchemaFromStruct[T any]() (json.RawMessage, error) {
	var zero T
	generator := NewOpenAISchemaGenerator()
	return generator.Generate(zero)
}

// ValidateOpenAICompatibility checks if a schema is compatible with OpenAI's tool system
func (g *OpenAISchemaGenerator) ValidateOpenAICompatibility(schema json.RawMessage) error {
	var schemaMap map[string]interface{}
	if err := json.Unmarshal(schema, &schemaMap); err != nil {
		return fmt.Errorf("invalid JSON schema: %w", err)
	}

	// Check for unsupported features
	if _, hasDefs := schemaMap["$defs"]; hasDefs {
		return fmt.Errorf("OpenAI doesn't support $defs, use ExpandedStruct: true")
	}

	if _, hasRefs := schemaMap["$ref"]; hasRefs {
		return fmt.Errorf("OpenAI doesn't support $ref well, use DoNotReference: true")
	}

	if _, hasAdditionalProps := schemaMap["additionalProperties"]; hasAdditionalProps {
		return fmt.Errorf("OpenAI doesn't support additionalProperties well")
	}

	return nil
}
