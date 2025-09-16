package agent

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/adapters/openai"
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

var simpleTool = tools.NewSimpleTool("greet", "Greet someone",
	func(ctx context.Context, input *Req) (*Res, error) {
		return &Res{Response: "Hello, " + input.Name + "!"}, errors.New("test error")
	},
	openai.NewOpenAISchemaGenerator(),
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
		ai.WithTools(simpleTool),
	),
	)

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
