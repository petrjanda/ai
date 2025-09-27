package structured

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/pkg/errors"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/tools"

	"github.com/xeipuuv/gojsonschema"
)

// StructuredLLM implements the StructuredLLM interface to provide structured output formatting
// It ignores tool call directives and forces the use of its own formatting tool
type StructuredLLM interface {
	tools.Tool
	ai.LLM
}

// LLM provides a basic implementation of the LLMWithStructuredOutput interface
type LLM struct {
	name        string
	description string
	inputSchema json.RawMessage
	llm         ai.LLM // The underlying LLM to delegate to

	retryConfig *RetryConfig
	events      ai.AgentEvents
}

// LLMOpts represents options for configuring an LLM with structured output
type LLMOpts = func(*LLM)

// WithName sets a custom name for the tool (defaults to "formatter")
func WithName(name string) LLMOpts {
	return func(f *LLM) {
		f.name = name
	}
}

// WithDescription sets a custom description for the tool
func WithDescription(description string) LLMOpts {
	return func(f *LLM) {
		f.description = description
	}
}

// WithAgentEvents converts AgentEvents to LLMEvents for use with structured tasks
func WithAgentEvents(events ai.AgentEvents) LLMOpts {
	return func(f *LLM) {
		f.events = events
	}
}

// NewLLM creates a new base LLM with structured output
// Uses sensible defaults: name="formatter", description="Must be called to provide structured output"
func NewLLM(inputSchema json.RawMessage, llm_ ai.LLM, opts ...LLMOpts) StructuredLLM {
	f := &LLM{
		name:        "formatter",
		description: "Must be called to provide structured output",
		inputSchema: inputSchema,
		llm:         llm_,
		events:      ai.NewNoopAgentEvents(),
	}

	// Apply options first to set up events
	for _, opt := range opts {
		opt(f)
	}

	// Initialize retry manager with default config if not set
	if f.retryConfig == nil {
		f.retryConfig = DefaultRetryConfig()
	}

	return f
}

// Name returns the name of the LLM with structured output
func (f *LLM) Name() string {
	return f.name
}

// Description returns the description of the LLM with structured output
func (f *LLM) Description() string {
	return f.description
}

// InputSchemaRaw returns the input schema as raw JSON
func (f *LLM) InputSchemaRaw() json.RawMessage {
	return f.inputSchema
}

// OutputSchemaRaw returns the output schema as raw JSON
func (f *LLM) OutputSchemaRaw() json.RawMessage {
	return nil
}

// Execute executes the LLM with structured output tool, returning the input as output (echo behavior)
func (f *LLM) Execute(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	// Create a schema loader and compile the schema
	schemaLoader := gojsonschema.NewBytesLoader(f.inputSchema)
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	// Create a document loader for the input
	documentLoader := gojsonschema.NewBytesLoader(args)

	// Validate the input against the schema
	result, err := schema.Validate(documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if !result.Valid() {
		// Collect all validation errors
		var errMsgs []string
		for _, err := range result.Errors() {
			errMsgs = append(errMsgs, err.String())
		}
		return nil, fmt.Errorf("validation errors: %s", strings.Join(errMsgs, "; "))
	}

	// For LLMs with structured output, we return the input as output to enforce structure
	return args, nil
}

// Invoke implements the LLM interface
// It ignores tool call directives and forces the use of this LLM with structured output
func (f *LLM) Invoke(ctx context.Context, request *ai.LLMRequest) (*ai.LLMResponse, error) {
	if f.llm == nil {
		return nil, fmt.Errorf("no underlying LLM configured")
	}

	// Create a new request that forces the use of the structured output formatter
	forcedRequest := request.Clone(
		// ai.WithSystem(`
		// 	You will be given a 'formatter' tool with input schema,
		// 	and a previous attempt of LLM to generate structured output that matches that schema.
		// 	Your job is to call the formatter tool with corrected arguments that match the schema.
		// `),
		ai.WithTools(f), // Only include the structured output formatter as a tool
		ai.WithToolUsage(tools.ForceTool(f.Name())), // Force the use of the formatter
	)

	// Log actual internal request
	f.events.OnRequest(ctx, forcedRequest)

	retrier := NewRetrier(f.retryConfig, NewStructuredRetriable(f.llm, forcedRequest, f.events, f))

	// Execute with retry
	return retrier.Execute(ctx, f.llm)
}

// MarshalJSON implements custom JSON marshaling for BaseLLMWithStructuredOutput
// This ensures that when the struct is serialized to JSON, it reports the tool name
func (f *LLM) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":         f.name,
		"description":  f.description,
		"input_schema": f.inputSchema,
	})
}

// RETRIABLE

type StructuredRetriable struct {
	llm       ai.LLM
	request   *ai.LLMRequest
	events    ai.AgentEvents
	formatter StructuredLLM

	previousError error
}

func NewStructuredRetriable(llm ai.LLM, request *ai.LLMRequest, events ai.AgentEvents, formatter StructuredLLM) *StructuredRetriable {
	return &StructuredRetriable{
		llm:           llm,
		request:       request,
		events:        events,
		formatter:     formatter,
		previousError: nil,
	}
}

func (s *StructuredRetriable) Retry(ctx context.Context, attempt int) (*ai.LLMResponse, error) {
	// Add the previous error to the request history
	if s.previousError != nil {
		s.request = s.request.Clone(ai.WithAddedHistory(ai.NewHistory(ai.NewUserMessage(s.previousError.Error()))))
	}

	// Delegate to the underlying LLM
	response, err := s.llm.Invoke(ctx, s.request)
	if err != nil {
		return nil, errors.Wrap(err, "underlying LLM invocation failed")
	}

	// Verify the LLM followed forced tool usage
	toolCalls := response.ToolCalls()
	if len(toolCalls) == 0 {
		return nil, errors.New("no tool call found in response - LLM did not follow forced tool usage")
	}

	// Verify the format of produced structured output
	toolCall := toolCalls[0]
	result, err := s.formatter.Execute(ctx, toolCall.Args)
	if err != nil {
		s.events.OnToolError(ctx, toolCall, attempt, err)
		return nil, errors.Wrap(err, "invalid structured output")
	}

	// Verify it can be marshalled
	if _, err := json.Marshal(result); err != nil {
		s.events.OnToolError(ctx, toolCall, attempt, err)
		return nil, errors.Wrap(err, "structured output marshalling failed")
	}

	payload := ai.NewLLMResponse(ai.NewAssistantMessage(string(result)))
	s.events.OnResponse(ctx, s.request, payload, true)

	return payload, nil
}

func (s *StructuredRetriable) OnFailure(ctx context.Context, attempt int, err error) error {
	s.previousError = err
	return nil
}
