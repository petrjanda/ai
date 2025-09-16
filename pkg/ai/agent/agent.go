package agent

import (
	"context"
	"encoding/json"
	"fmt"

	_ "embed"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/structured"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/tools"
)

// Agent represents an agent that can use tools and interact with an LLM
type Agent struct {
	llm ai.LLM

	retryConfig *structured.RetryConfig

	events ai.AgentEvents

	totalUsage *ai.LLMUsage
}

// AgentOpts represents options for configuring an agent
type AgentOpts = func(*Agent)

// WithRetryConfig sets the retry configuration for tool calls with structured output support
func WithRetryConfig(config *structured.RetryConfig) AgentOpts {
	return func(a *Agent) {
		a.retryConfig = config
	}
}

func WithEvents(events ai.AgentEvents) AgentOpts {
	return func(a *Agent) {
		a.events = events
	}
}

// NewAgent creates a new agent with the given LLM and tools
func NewAgent(llm_ ai.LLM, opts ...AgentOpts) ai.LLM {
	a := &Agent{
		llm:        llm_,
		events:     ai.NewNoopAgentEvents(),
		totalUsage: ai.NewLLMUsage(0, 0, 0),
	}

	// Apply options first to set up events
	for _, opt := range opts {
		opt(a)
	}

	// Initialize retry manager with default config if not set
	if a.retryConfig == nil {
		a.retryConfig = structured.DefaultRetryConfig()
	}

	return a
}

// Loop processes the conversation loop, handling tool calls and LLM responses
func (a *Agent) Invoke(ctx context.Context, request *ai.LLMRequest) (*ai.LLMResponse, error) {
	a.events.OnRequest(ctx, request)

	response, err := a.llm.Invoke(ctx, request)
	if err != nil {
		a.events.OnRequestError(ctx, request, err)
		return nil, err
	}

	a.events.OnResponse(ctx, request, response)
	a.totalUsage.Add(response.Usage)

	if len(response.ToolCalls()) > 0 {
		for _, toolCall := range response.ToolCalls() {
			a.events.OnToolCall(ctx, toolCall)

			// Find the tool to get its input schema
			targetTool, err := request.Tools.FindTool(toolCall.Name)
			if err != nil {
				return nil, err
			}

			retrier := structured.NewRetrier(a.retryConfig, NewToolCallRetriable(a.llm, toolCall, targetTool, a.events))

			message, err := retrier.Execute(ctx, a.llm)
			if err != nil {
				response.AddMessage(ai.NewToolResultErrorMessage(toolCall, err.Error()))
				a.totalUsage.AddToolCall(toolCall, err)
			} else {
				response.AddMessage(message)
				a.totalUsage.AddToolCall(toolCall, nil)
			}
		}

		request = request.Clone(
			ai.WithHistory(request.History.Append(response.Messages...)),
		)

		// Return usage 'to-date' rather than just the last response's usage
		response.SetUsage(a.totalUsage)

		return a.Invoke(ctx, request)
	}

	// Return usage 'to-date' rather than just the last response's usage
	response.SetUsage(a.totalUsage)

	return response, nil
}

// RETRIABLE

type ToolCallRetriable struct {
	llm ai.LLM

	toolCall   *tools.ToolCall
	targetTool tools.Tool
	events     ai.AgentEvents
}

func NewToolCallRetriable(llm ai.LLM, toolCall *tools.ToolCall, targetTool tools.Tool, events ai.AgentEvents) *ToolCallRetriable {
	return &ToolCallRetriable{llm: llm, toolCall: toolCall, targetTool: targetTool, events: events}
}

func (t *ToolCallRetriable) Retry(ctx context.Context, attempt int) (ai.Message, error) {
	result, err := t.targetTool.Execute(ctx, t.toolCall.Args)
	if err != nil {
		t.events.OnToolError(ctx, t.toolCall, attempt, err)
		return nil, err
	}

	t.events.OnToolResult(ctx, t.toolCall, result)
	return ai.NewToolResultMessage(t.toolCall, result), nil
}

func (t *ToolCallRetriable) OnFailure(ctx context.Context, attempt int, err error) error {
	corrector := structured.NewCorrector(ai.Claude4Sonnet, t.targetTool.InputSchemaRaw(), "You are a tool call corrector. You are given a tool call that failed and you need to correct it.")

	corrected, err := corrector.Execute(ctx, t.llm, ai.NewHistory(
		ai.NewUserMessage(fmt.Sprintf(`
		Tool call to '%s' failed with error: %s
		Failed parameters: %s
		Use 'formatter' tool to generate corrected parameters that match the tool's input schema above.`, t.toolCall.Name, err.Error(), prettyJSON(t.toolCall.Args))),
	))

	if err != nil {
		return err
	}

	t.toolCall.Args = corrected

	return nil
}

func prettyJSON(v any) string {
	json, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("Failed to marshal JSON: %v", err)
	}
	return string(json)
}
