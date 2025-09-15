package ai

import "github.com/getsynq/ai/pkg/ai/tools"

type LLMRequest struct {
	Model     ModelId         `json:"model"`
	System    string          `json:"system"`
	History   History         `json:"history"`
	Tools     []tools.Tool    `json:"tools"`
	ToolUsage tools.ToolUsage `json:"tool_usage"`

	MaxCompletionTokens int     `json:"max_completion_tokens"`
	Temperature         float64 `json:"temperature"`
}

type ModelId string

const (
	Claude3Sonnet ModelId = "claude-3-7-sonnet"
	Claude4Sonnet ModelId = "claude-4-sonnet"
	Claude3Haiku  ModelId = "claude-3-5-haiku"
	Gemini25Flash ModelId = "gemini-2-5-flash"
	Gemini25Pro   ModelId = "gemini-2-5-pro"
)

type LLMRequestOpts = func(*LLMRequest)

func WithModel(model ModelId) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.Model = model
	}
}

func WithToolUsage(toolUsage tools.ToolUsage) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.ToolUsage = toolUsage
	}
}

func WithTools(tools ...tools.Tool) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.Tools = append(r.Tools, tools...)
	}
}

func WithSystem(system string) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.System = system
	}
}

func WithHistory(history History) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.History = history
	}
}

func WithAddedHistory(history History) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.History = r.History.Append(history...)
	}
}

func WithMaxCompletionTokens(maxCompletionTokens int) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.MaxCompletionTokens = maxCompletionTokens
	}
}

func WithTemperature(temperature float64) LLMRequestOpts {
	return func(r *LLMRequest) {
		r.Temperature = temperature
	}
}

func WithRequest(request *LLMRequest) LLMRequestOpts {
	return func(r *LLMRequest) {
		*r = *request
	}
}

func NewLLMRequest(opts ...LLMRequestOpts) *LLMRequest {
	r := &LLMRequest{
		ToolUsage: tools.AutoToolSelection(), // Default to auto tool selection
	}
	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *LLMRequest) With(opt LLMRequestOpts) *LLMRequest {
	opt(r)
	return r
}

func (r *LLMRequest) Clone(opts ...LLMRequestOpts) *LLMRequest {
	req := &LLMRequest{
		Model:               r.Model,
		History:             r.History,
		ToolUsage:           r.ToolUsage,
		Tools:               r.Tools,
		System:              r.System,
		MaxCompletionTokens: r.MaxCompletionTokens,
		Temperature:         r.Temperature,
	}

	for _, opt := range opts {
		opt(req)
	}

	return req
}
