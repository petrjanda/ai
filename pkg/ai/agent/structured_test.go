package agent

// import (
// 	"encoding/json"
// 	"testing"
// )

// func TestBaseLLMWithStructuredOutput_JSONSerialization(t *testing.T) {
// 	// Create a test instance
// 	llm := &LLM{
// 		name:        "test_formatter",
// 		description: "Test formatter tool",
// 	}

// 	// Marshal to JSON
// 	jsonData, err := json.Marshal(llm)
// 	if err != nil {
// 		t.Fatalf("Failed to marshal to JSON: %v", err)
// 	}

// 	// Unmarshal to verify the structure
// 	var result map[string]interface{}
// 	if err := json.Unmarshal(jsonData, &result); err != nil {
// 		t.Fatalf("Failed to unmarshal JSON: %v", err)
// 	}

// 	// Verify the fields
// 	if result["name"] != "test_formatter" {
// 		t.Errorf("Expected name to be 'test_formatter', got %v", result["name"])
// 	}

// 	if result["description"] != "Test formatter tool" {
// 		t.Errorf("Expected description to be 'Test formatter tool', got %v", result["description"])
// 	}

// 	if result["type"] != "tool" {
// 		t.Errorf("Expected type to be 'tool', got %v", result["type"])
// 	}

// 	// Verify that internal fields are not exposed
// 	if _, exists := result["inputSchema"]; exists {
// 		t.Error("inputSchema should not be exposed in JSON")
// 	}

// 	if _, exists := result["llm"]; exists {
// 		t.Error("llm should not be exposed in JSON")
// 	}

// 	if _, exists := result["events"]; exists {
// 		t.Error("events should not be exposed in JSON")
// 	}
// }

// func TestBaseLLMWithStructuredOutput_JSONSerializationWithDefaultName(t *testing.T) {
// 	// Create a test instance with default name
// 	llm := &LLM{
// 		name:        "formatter",
// 		description: "Must be called to provide structured output",
// 	}

// 	// Marshal to JSON
// 	jsonData, err := json.Marshal(llm)
// 	if err != nil {
// 		t.Fatalf("Failed to marshal to JSON: %v", err)
// 	}

// 	// Unmarshal to verify the structure
// 	var result map[string]interface{}
// 	if err := json.Unmarshal(jsonData, &result); err != nil {
// 		t.Fatalf("Failed to unmarshal JSON: %v", err)
// 	}

// 	// Verify the fields
// 	if result["name"] != "formatter" {
// 		t.Errorf("Expected name to be 'formatter', got %v", result["name"])
// 	}

// 	if result["description"] != "Must be called to provide structured output" {
// 		t.Errorf("Expected description to be 'Must be called to provide structured output', got %v", result["description"])
// 	}

// 	if result["type"] != "tool" {
// 		t.Errorf("Expected type to be 'tool', got %v", result["type"])
// 	}
// }
