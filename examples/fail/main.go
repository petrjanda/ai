package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

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

	llm := GetLiteLLM()

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

func GetLiteLLM() ai.LLM {
	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	endpoint := os.Getenv("OPENAI_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}

	// Create OpenAI adapter
	return openai.NewOpenAIAdapter(apiKey, openai.WithEndpoint(endpoint))
}
