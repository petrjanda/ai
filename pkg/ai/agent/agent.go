package agent

import (
	"context"

	_ "embed"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/structured"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/tools"
)

// Agent represents an agent that can use tools and interact with an LLM
type Agent struct {
	llm ai.LLM

	retryManager *structured.RetryManager

	events ai.AgentEvents

	totalUsage *ai.LLMUsage
}

// AgentOpts represents options for configuring an agent
type AgentOpts = func(*Agent)

// WithRetryConfig sets the retry configuration for tool calls with structured output support
func WithRetryConfig(config *structured.RetryConfig) AgentOpts {
	return func(a *Agent) {
		a.retryManager = structured.NewRetryManager(a.llm, config)
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
	if a.retryManager == nil {
		a.retryManager = structured.NewRetryManager(a.llm, structured.DefaultRetryConfig())
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

			message, err := a.callTool(ctx, request.Tools, toolCall)
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

// CallTool executes a tool call with retry logic using the retry component
func (a *Agent) callTool(ctx context.Context, toolbox tools.Toolbox, toolCall *tools.ToolCall) (ai.Message, error) {
	// Find the tool to get its input schema
	targetTool, err := toolbox.FindTool(toolCall.Name)
	if err != nil {
		return nil, err
	}

	// Execute with retry
	return structured.ExecuteWithRetry(a.retryManager, ctx, NewToolCallOperation(toolCall, targetTool, a.events))
}
