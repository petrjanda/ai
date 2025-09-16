package agent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/structured"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/tools"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AgentSuite struct {
	suite.Suite
}

func TestAgentSuite(t *testing.T) {
	suite.Run(t, new(AgentSuite))
}

func (s *AgentSuite) SetupTest() {

}

type Req struct {
	Name string `json:"name"`
}

type Res struct {
	Response string `json:"response"`
}

var greetTool = tools.NewSimpleTool("greet", "Greet someone",
	func(ctx context.Context, input *Req) (*Res, error) {
		return &Res{Response: "Hello, " + input.Name + "!"}, nil
	},
)

func (s *AgentSuite) TestAgent() {
	llm := NewMockLLM()
	llm.
		On("Invoke", mock.Anything, mock.Anything).
		Return(ai.NewLLMResponse(
			ai.NewToolCallMessage(tools.NewToolCall("1", "greet", json.RawMessage(`{"name": "John"}`))),
		), nil).
		Once()

	llm.
		On("Invoke", mock.Anything, mock.Anything).
		Return(ai.NewLLMResponse(ai.NewAssistantMessage("Done.")), nil).
		Once()

	agent := NewAgent(llm, WithRetryConfig(structured.NewRetryConfig(string(ai.Claude4Sonnet), 1, 100*time.Millisecond, 2.0)))
	res, err := agent.Invoke(context.Background(), ai.NewLLMRequest(
		ai.WithModel(ai.Claude4Sonnet),
		ai.WithHistory(ai.NewHistory(
			ai.NewUserMessage("Greet John"),
		)),
		ai.WithTools(greetTool),
	),
	)

	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(1, len(res.Messages))
	s.Require().Equal(ai.NewAssistantMessage("Done."), res.Messages[0])
}

var greetFailingTool = tools.NewSimpleTool("greet", "Greet someone",
	func(ctx context.Context, input *Req) (*Res, error) {
		if input.Name == "John" {
			return &Res{Response: "Hello, " + input.Name + "!"}, nil // On retry things will pass
		}
		return nil, errors.New("test error")
	},
)

func (s *AgentSuite) TestAgentWithFailingTool() {
	llm := NewMockLLM()
	llm.
		On("Invoke", mock.Anything, mock.Anything).
		Return(ai.NewLLMResponse(
			ai.NewToolCallMessage(tools.NewToolCall("1", "greet", json.RawMessage(`{"name": "Tom"}`))),
		), nil).
		Once()

	// Retry call for tool correction (when tool fails)
	llm.
		On("Invoke", mock.Anything, mock.Anything).
		Return(ai.NewLLMResponse(
			ai.NewToolCallMessage(tools.NewToolCall("2", "formatter", json.RawMessage(`{"name": "John"}`))), // This will fix it
		), nil).
		Once()

	// Second attempt after correction
	llm.
		On("Invoke", mock.Anything, mock.Anything).
		Return(ai.NewLLMResponse(ai.NewAssistantMessage("Done.")), nil).
		Once()

	agent := NewAgent(llm, WithRetryConfig(structured.NewRetryConfig(string(ai.Claude4Sonnet), 2, 100*time.Millisecond, 2.0)))
	res, err := agent.Invoke(context.Background(), ai.NewLLMRequest(
		ai.WithModel(ai.Claude4Sonnet),
		ai.WithHistory(ai.NewHistory(
			ai.NewUserMessage("Greet Tom"),
		)),
		ai.WithTools(greetFailingTool),
	))

	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(1, len(res.Messages))
	s.Require().Equal(ai.NewAssistantMessage("Done."), res.Messages[0])
}

type MockLLM struct {
	mock.Mock
}

func NewMockLLM() *MockLLM {
	return &MockLLM{}
}

func (m *MockLLM) Invoke(ctx context.Context, request *ai.LLMRequest) (*ai.LLMResponse, error) {
	args := m.Called(ctx, request)

	return args.Get(0).(*ai.LLMResponse), args.Error(1)
}
