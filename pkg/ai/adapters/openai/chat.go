package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/tools"
	"github.com/pkg/errors"

	openai "github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"github.com/openai/openai-go/v2/shared"
)

// OpenAIAdapter implements the LLM interface using OpenAI's API
type OpenAIAdapter struct {
	client   *openai.Client
	endpoint string
}

// OpenAIAdapterOpts represents options for configuring the OpenAI adapter
type OpenAIAdapterOpts = func(*OpenAIAdapter)

func WithEndpoint(endpoint string) OpenAIAdapterOpts {
	return func(a *OpenAIAdapter) {
		a.endpoint = endpoint
	}
}

// NewOpenAIAdapter creates a new OpenAI adapter with the given API key and options
func NewOpenAIAdapter(apiKey string, opts ...OpenAIAdapterOpts) *OpenAIAdapter {
	client := openai.NewClient(option.WithAPIKey(apiKey))

	adapter := &OpenAIAdapter{
		client: &client,
	}

	for _, opt := range opts {
		opt(adapter)
	}

	return adapter
}

// Invoke implements the LLM interface by calling OpenAI's API
func (a *OpenAIAdapter) Invoke(ctx context.Context, request *ai.LLMRequest) (*ai.LLMResponse, error) {

	history := request.History

	if request.System != "" {
		history = append(ai.NewHistory(ai.NewSystemMessage(request.System)), request.History...)
	}

	messages, err := a.convertMessages(history)
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert messages")
	}

	chatReq := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(request.Model),
		Messages: messages,
	}

	if request.MaxCompletionTokens > 0 {
		chatReq.MaxCompletionTokens = openai.Int(int64(request.MaxCompletionTokens))
	}

	if request.Temperature > 0 {
		chatReq.Temperature = openai.Float(request.Temperature)
	}

	// Handle tool usage based on the ToolUsage strategy
	if request.ToolUsage != nil && len(request.Tools) > 0 {
		tools := a.convertTools(request.Tools)
		chatReq.Tools = tools

		// Convert tool usage to OpenAI format
		toolChoice, err := convertToolUsage(request.ToolUsage, request.Tools)
		if err != nil {
			return nil, fmt.Errorf("failed to convert tool usage: %w", err)
		}

		if toolChoice != nil {
			chatReq.ToolChoice = *toolChoice
		}
	}

	payload, _ := json.MarshalIndent(chatReq, "", "  ")
	slog.Debug("request", "request", string(payload))
	resp, err := a.client.Chat.Completions.New(ctx, chatReq, option.WithBaseURL(a.endpoint))

	payload, _ = json.MarshalIndent(resp, "", "  ")
	slog.Debug("response", "response", string(payload))

	if err != nil {
		return nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	response := ai.NewLLMResponse()

	usage := ai.NewLLMUsage(
		resp.Usage.PromptTokens,
		resp.Usage.CompletionTokens,
		resp.Usage.TotalTokens,
	)

	if len(resp.Choices) > 0 {
		choice := resp.Choices[0]
		if choice.Message.Content != "" {
			textMsg := ai.NewAssistantMessage(choice.Message.Content)
			response.AddMessage(textMsg)
		}

		if choice.Message.ToolCalls != nil {
			for _, toolCall := range choice.Message.ToolCalls {
				ourToolCall := &tools.ToolCall{
					ID:   toolCall.ID,
					Name: toolCall.Function.Name,
					Args: json.RawMessage(toolCall.Function.Arguments),
				}
				response.AddToolCall(ourToolCall)
			}
		}
	}

	response.SetUsage(usage)
	return response, nil
}

// convertMessages converts our Message interface to OpenAI's format
func (a *OpenAIAdapter) convertMessages(messages []ai.Message) ([]openai.ChatCompletionMessageParamUnion, error) {
	var openaiMessages []openai.ChatCompletionMessageParamUnion

	for _, msg := range messages {
		switch m := msg.(type) {
		case *ai.TextMessage:
			switch m.Role() {
			case ai.MessageRoleUser:
				openaiMessages = append(openaiMessages, openai.UserMessage(m.Content))
			case ai.MessageRoleAssistant:
				openaiMessages = append(openaiMessages, openai.AssistantMessage(m.Content))
			case ai.MessageRoleSystem:
				openaiMessages = append(openaiMessages, openai.SystemMessage(m.Content))
			}

		case *ai.ToolCallMessage:
			// Convert tool call to assistant message with tool_calls
			asst := openai.ChatCompletionAssistantMessageParam{
				Role: "assistant",
				ToolCalls: []openai.ChatCompletionMessageToolCallUnionParam{
					{
						OfFunction: &openai.ChatCompletionMessageFunctionToolCallParam{
							ID: m.ToolCall.ID,
							Function: openai.ChatCompletionMessageFunctionToolCallFunctionParam{
								Name:      m.ToolCall.Name,
								Arguments: string(m.ToolCall.Args),
							},
							Type: "function",
						},
					},
				},
			}
			openaiMessages = append(openaiMessages, openai.ChatCompletionMessageParamUnion{OfAssistant: &asst})

		case *ai.ToolResultMessage:
			if m.Error != "" {
				msg := map[string]string{"error": m.Error}
				payload, err := json.Marshal(msg)
				if err != nil {
					return nil, errors.Wrap(err, "failed to marshal tool result error")
				}

				openaiMessages = append(openaiMessages, openai.ToolMessage(string(payload), m.ToolCall.ID))
			} else {
				openaiMessages = append(openaiMessages, openai.ToolMessage(string(m.Result), m.ToolCall.ID))
			}

		case *ai.ToolErrorMessage:
			// Don't send tool error messages directly to OpenAI
			// Instead, we'll handle retries differently to avoid API violations
			// This case should not occur in normal operation with the new retry mechanism
			continue
		}
	}

	return openaiMessages, nil
}

// convertTools converts our Tool interface to OpenAI's format
func (a *OpenAIAdapter) convertTools(tools []tools.Tool) []openai.ChatCompletionToolUnionParam {
	var openaiTools []openai.ChatCompletionToolUnionParam

	for _, tool := range tools {
		// Parse the JSON schema to convert to FunctionParameters
		var params map[string]any
		if err := json.Unmarshal(tool.InputSchemaRaw(), &params); err != nil {
			// If we can't parse the schema, use an empty object
			params = make(map[string]any)
		}

		functionDef := shared.FunctionDefinitionParam{
			Name:        tool.Name(),
			Description: openai.String(tool.Description()),
			Parameters:  shared.FunctionParameters(params),
		}

		openaiTool := openai.ChatCompletionFunctionTool(functionDef)
		openaiTools = append(openaiTools, openaiTool)
	}

	return openaiTools
}
