package examples

import (
	"log"
	"os"

	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/adapters/openai"
)

func GetLiteLLM() ai.LLM {
	// Get OpenAI API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is requireds")
	}

	endpoint := os.Getenv("OPENAI_ENDPOINT")
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}

	// Create OpenAI adapter
	return openai.NewOpenAIAdapter(apiKey, openai.WithEndpoint(endpoint))
}
