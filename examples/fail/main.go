package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/getsynq/cloud/ai-data-sre/examples"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/adapters/openai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/agent"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/structured"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/tools"
	"github.com/joho/godotenv"
)

var schemas = openai.NewOpenAISchemaGenerator()

type Request struct {
	Name string `json:"name"`
}

type Response struct {
	Response string `json:"response"`
}

var tool = tools.NewSimpleTool("greet", "Greet someone",
	func(ctx context.Context, input *Request) (*Response, error) {
		if input == nil {
			return nil, errors.New("request cannot be nil")
		}
		if input.Name == "" {
			return nil, errors.New("name cannot be empty")
		}

		return &Response{
			Response: fmt.Sprintf("Hello, %s!", input.Name),
		}, nil
	},
	schemas,
)

func main() {
	godotenv.Load()

	llm := examples.GetLiteLLM()

	agent := agent.NewAgent(llm,
		agent.WithRetryConfig(structured.NewRetryConfig(string(ai.Claude4Sonnet), 1, 100*time.Millisecond, 2.0)),
		agent.WithEvents(ai.NewLogAgentEvents(slog.Default())),
	)

	res, err := agent.Invoke(context.Background(), ai.NewLLMRequest(
		ai.WithModel(ai.Claude4Sonnet),
		ai.WithHistory(ai.NewHistory(
			ai.NewUserMessage("Call greeting without providing a name"),
		)),
		ai.WithTools(tool),
	))

	if err != nil {
		log.Fatalf("error: %s", err)
	}

	log.Printf("usage: %s", res.Usage)
}
