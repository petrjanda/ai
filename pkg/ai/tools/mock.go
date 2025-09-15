package tools

import (
	"context"
	"encoding/json"
	"time"

	"gopkg.in/yaml.v3"
)

// ToolCallRecord represents a recorded tool call for testing and assertion purposes
type ToolCallRecord struct {
	// ToolName is the name of the tool that was called
	ToolName string `json:"tool_name"`

	// Args are the arguments passed to the tool
	Args json.RawMessage `json:"args"`

	// Timestamp is when the tool was called
	Timestamp time.Time `json:"timestamp"`
}

// MockTool is a tool that records all calls and allows mocking of responses
type MockTool struct {
	Name_        string          `yaml:"name"`
	Description_ string          `yaml:"description"`
	InputSchema  json.RawMessage `yaml:"inputSchema"`
	OutputSchema json.RawMessage `yaml:"outputSchema"`

	// records stores all tool calls for later inspection
	records []ToolCallRecord

	// MockResponse is the response to return when the tool is called
	MockResponse json.RawMessage `yaml:"mockResponse"`

	// MockError is the error to return when the tool is called
	MockError error `yaml:"mockError"`
}

// NewStubTool creates a new stub tool by inspecting the provided tool
func NewMockTool(original Tool) *MockTool {
	return &MockTool{
		Name_:        original.Name(),
		Description_: original.Description(),
		InputSchema:  original.InputSchemaRaw(),
		OutputSchema: original.OutputSchemaRaw(),
		records:      make([]ToolCallRecord, 0),
	}
}

func NewMockToolRaw(name, description string, inputSchema json.RawMessage) *MockTool {
	return &MockTool{
		Name_:        name,
		Description_: description,
		InputSchema:  inputSchema,
		records:      make([]ToolCallRecord, 0),
	}
}

// Name returns the name of the tool
func (s *MockTool) Name() string {
	return s.Name_
}

// Description returns the description of the tool
func (s *MockTool) Description() string {
	return s.Description_
}

// InputSchemaRaw returns the JSON schema for the tool's input
func (s *MockTool) InputSchemaRaw() json.RawMessage {
	return s.InputSchema
}

// OutputSchemaRaw returns the JSON schema for the tool's output
func (s *MockTool) OutputSchemaRaw() json.RawMessage {
	return s.OutputSchema
}

// Execute executes the tool and records the call
func (s *MockTool) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	record := ToolCallRecord{
		ToolName:  s.Name_,
		Args:      args,
		Timestamp: time.Now(),
	}

	s.records = append(s.records, record)

	if s.MockError != nil {
		return nil, s.MockError
	}

	return s.MockResponse, nil
}

// SetMockResponse sets the response to return when the tool is called
func (s *MockTool) RespondWith(response json.RawMessage) *MockTool {
	s.MockError = nil
	s.MockResponse = response

	return s
}

// SetMockError sets the error to return when the tool is called
func (s *MockTool) ErrorsWith(err error) *MockTool {
	s.MockResponse = nil
	s.MockError = err

	return s
}

// ClearRecords clears all recorded calls
func (s *MockTool) Clear() {
	s.records = make([]ToolCallRecord, 0)
}

// GetRecords returns a copy of all recorded tool calls
func (s *MockTool) Calls() []ToolCallRecord {
	records := make([]ToolCallRecord, len(s.records))
	copy(records, s.records)
	return records
}

// GetLastCall returns the most recent tool call, or nil if no calls have been made
func (s *MockTool) LastCall() *ToolCallRecord {
	if len(s.records) == 0 {
		return nil
	}

	lastCall := s.records[len(s.records)-1]

	return &lastCall
}

// WasCalled returns true if the tool was called at least once
func (s *MockTool) WasCalled() bool {
	return len(s.Calls()) > 0
}

// func (s *MockTool) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(map[string]interface{}{
// 		"name":         s.Name_,
// 		"description":  s.Description_,
// 		"inputSchema":  s.InputSchema,
// 		"outputSchema": s.OutputSchema,
// 	})
// }

// func (s *MockTool) UnmarshalJSON(data []byte) error {
// 	var tmp map[string]interface{}
// 	if err := json.Unmarshal(data, &tmp); err != nil {
// 		return err
// 	}
// 	s.Name_ = tmp["name"].(string)
// 	s.Description_ = tmp["description"].(string)
// 	s.InputSchema = tmp["inputSchema"].(json.RawMessage)
// 	s.OutputSchema = tmp["outputSchema"].(json.RawMessage)
// 	return nil
// }

// MarshalYAML implements yaml.Marshaler interface for YAML export
func (s *MockTool) MarshalYAML() (interface{}, error) {
	// Parse input schema to interface{} for better YAML representation
	var inputSchema interface{}
	if err := json.Unmarshal(s.InputSchema, &inputSchema); err != nil {
		// If JSON unmarshalling fails, treat as raw JSON string
		inputSchema = json.RawMessage(s.InputSchema)
	}

	// Parse output schema to interface{} for better YAML representation
	var outputSchema interface{}
	if len(s.OutputSchema) > 0 {
		if err := json.Unmarshal(s.OutputSchema, &outputSchema); err != nil {
			// If JSON unmarshalling fails, treat as raw JSON string
			outputSchema = json.RawMessage(s.OutputSchema)
		}
	}

	var mockResponse interface{}
	if len(s.MockResponse) > 0 {
		if err := json.Unmarshal(s.MockResponse, &mockResponse); err != nil {
			mockResponse = json.RawMessage(s.MockResponse)
		}
	}

	var mockError interface{}
	if s.MockError != nil {
		mockError = s.MockError
	}

	base := map[string]interface{}{
		"name":          s.Name_,
		"description":   s.Description_,
		"input_schema":  inputSchema,
		"output_schema": outputSchema,
	}

	if mockResponse != nil {
		base["mock_response"] = mockResponse
	}

	if mockError != nil {
		base["mock_error"] = mockError
	}

	return base, nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface for YAML import
func (s *MockTool) UnmarshalYAML(value *yaml.Node) error {
	// Create a temporary struct to hold the YAML data
	var tmp struct {
		Name         string      `yaml:"name"`
		Description  string      `yaml:"description"`
		InputSchema  interface{} `yaml:"input_schema"`
		OutputSchema interface{} `yaml:"output_schema"`
		MockResponse interface{} `yaml:"mock_response"`
		MockError    error       `yaml:"mock_error"`
	}

	if err := value.Decode(&tmp); err != nil {
		return err
	}

	s.Name_ = tmp.Name
	s.Description_ = tmp.Description
	s.MockError = tmp.MockError

	// Convert input schema to json.RawMessage
	inputSchemaBytes, err := json.Marshal(tmp.InputSchema)
	if err != nil {
		return err
	}
	s.InputSchema = json.RawMessage(inputSchemaBytes)

	// Convert output schema to json.RawMessage if present
	if tmp.OutputSchema != nil {
		outputSchemaBytes, err := json.Marshal(tmp.OutputSchema)
		if err != nil {
			return err
		}
		s.OutputSchema = json.RawMessage(outputSchemaBytes)
	}

	// Convert mock response to json.RawMessage if present
	if tmp.MockResponse != nil {
		mockResponseBytes, err := json.Marshal(tmp.MockResponse)
		if err != nil {
			return err
		}
		s.MockResponse = json.RawMessage(mockResponseBytes)
	}

	// Initialize records slice
	s.records = make([]ToolCallRecord, 0)

	return nil
}

func ImportFromYAML(data []byte) Toolbox {
	var tools []*MockTool
	if err := yaml.Unmarshal(data, &tools); err != nil {
		panic(err)
	}

	var toolbox Toolbox
	for _, tool := range tools {
		toolbox = append(toolbox, tool)
	}

	return toolbox
}
