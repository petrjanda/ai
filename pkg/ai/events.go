package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/tools"
)

type LLMEvents interface {
	OnRequest(ctx context.Context, request *LLMRequest)
	OnResponse(ctx context.Context, request *LLMRequest, response *LLMResponse)
	OnRequestError(ctx context.Context, request *LLMRequest, err error)
}

type AgentEvents interface {
	LLMEvents

	OnToolCall(ctx context.Context, toolCall *tools.ToolCall)
	OnToolError(ctx context.Context, toolCall *tools.ToolCall, attempt int, err error)
	OnToolResult(ctx context.Context, toolCall *tools.ToolCall, result json.RawMessage)
}

type NoopAgentEvents struct{}

func NewNoopAgentEvents() *NoopAgentEvents {
	return &NoopAgentEvents{}
}

func (e *NoopAgentEvents) OnRequest(ctx context.Context, request *LLMRequest) {}
func (e *NoopAgentEvents) OnResponse(ctx context.Context, request *LLMRequest, response *LLMResponse) {
}
func (e *NoopAgentEvents) OnRequestError(ctx context.Context, request *LLMRequest, err error) {}
func (e *NoopAgentEvents) OnToolError(ctx context.Context, toolCall *tools.ToolCall, attempt int, err error) {
}
func (e *NoopAgentEvents) OnToolCall(ctx context.Context, toolCall *tools.ToolCall) {}
func (e *NoopAgentEvents) OnToolResult(ctx context.Context, toolCall *tools.ToolCall, result json.RawMessage) {
}

type LogAgentEvents struct {
	logger *slog.Logger
}

func NewLogAgentEvents(logger *slog.Logger) *LogAgentEvents {
	return &LogAgentEvents{logger: logger}
}

func NewJSONFileLogAgentEvents(path string) *LogAgentEvents {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	return &LogAgentEvents{logger: slog.New(slog.NewJSONHandler(file, nil))}
}

func (e *LogAgentEvents) OnRequest(ctx context.Context, request *LLMRequest) {
	if msg := printMessage(request.History.Last(), true); msg != "" {
		e.logger.Info("message", "message", msg)
	}
}

func (e *LogAgentEvents) OnResponse(ctx context.Context, request *LLMRequest, response *LLMResponse) {
	for _, message := range response.Messages {
		if msg := printMessage(message, false); msg != "" {
			e.logger.Info("message", "message", msg)
		}
	}

	e.logger.Info("usage", "usage", response.Usage)
}

func (e *LogAgentEvents) OnRequestError(ctx context.Context, request *LLMRequest, err error) {
	e.logger.Error("request error", "error", err)
}

func (e *LogAgentEvents) OnToolCall(ctx context.Context, toolCall *tools.ToolCall) {
	e.logger.Info("tool call", "tool", toolCall.Name, "args", toolCall.Args)
}

func (e *LogAgentEvents) OnToolError(ctx context.Context, toolCall *tools.ToolCall, attempt int, err error) {
	e.logger.Info("tool call failed",
		"tool", toolCall.Name,
		"attempt", attempt+1,
		"error", err.Error(),
	)
}

func (e *LogAgentEvents) OnToolResult(ctx context.Context, toolCall *tools.ToolCall, result json.RawMessage) {
	e.logger.Info("tool call result", "tool", toolCall.Name, "result", string(result))
}

func printMessage(message Message, textOnly bool) string {
	switch t := message.(type) {
	case *TextMessage:
		return fmt.Sprintf("%s: %s", t.Role(), t.Content)

	case *ToolCallMessage:
		if textOnly {
			return ""
		}

		return fmt.Sprintf("tool call: calling %s with args: %s", t.ToolCall.Name, t.ToolCall.Args)

	case *ToolResultMessage:
		if textOnly {
			return ""
		}

		if t.Error != "" {
			return fmt.Sprintf("tool error: %s -> %s", t.ToolCall.Name, t.Error)
		}

		return fmt.Sprintf("tool result: %s -> %s", t.ToolCall.Name, t.Result)

	case *ToolErrorMessage:
		if textOnly {
			return ""
		}

		return fmt.Sprintf("tool error: %s -> %s", t.ToolCall.Name, t.Error)

	default:
		return fmt.Sprintf("unknown message type: %T", t)
	}
}

type MultiplexEvents struct {
	events []AgentEvents
}

func NewMultiplexEvents(events ...AgentEvents) *MultiplexEvents {
	return &MultiplexEvents{events: events}
}

func (e *MultiplexEvents) Add(events ...AgentEvents) {
	e.events = append(e.events, events...)
}

func (e *MultiplexEvents) OnRequest(ctx context.Context, request *LLMRequest) {
	for _, event := range e.events {
		event.OnRequest(ctx, request)
	}
}

func (e *MultiplexEvents) OnResponse(ctx context.Context, request *LLMRequest, response *LLMResponse) {
	for _, event := range e.events {
		event.OnResponse(ctx, request, response)
	}
}

func (e *MultiplexEvents) OnRequestError(ctx context.Context, request *LLMRequest, err error) {
	for _, event := range e.events {
		event.OnRequestError(ctx, request, err)
	}
}

func (e *MultiplexEvents) OnToolCall(ctx context.Context, toolCall *tools.ToolCall) {
	for _, event := range e.events {
		event.OnToolCall(ctx, toolCall)
	}
}

func (e *MultiplexEvents) OnToolError(ctx context.Context, toolCall *tools.ToolCall, attempt int, err error) {
	for _, event := range e.events {
		event.OnToolError(ctx, toolCall, attempt, err)
	}
}

func (e *MultiplexEvents) OnToolResult(ctx context.Context, toolCall *tools.ToolCall, result json.RawMessage) {
	for _, event := range e.events {
		event.OnToolResult(ctx, toolCall, result)
	}
}
