package main

import (
	"context"
	"log/slog"

	"github.com/getsynq/cloud/ai-data-sre/examples"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/adapters/openai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/structured"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/workflows"
	"github.com/joho/godotenv"
)

type Translation struct {
	// Input message
	Original string `json:"original" jsonschema:"required" jsonschema_description:"The input message"`

	// Output message
	Translated string `json:"translated" jsonschema:"required" jsonschema_description:"The output message"`
}

var TranslationSchema = openai.NewOpenAISchemaGenerator().MustGenerate(&Translation{})

var PiratesTask = workflows.NewStructuredTask[Translation](
	"arrrr",
	ai.NewLLMRequest(
		ai.WithModel(ai.Gemini25Flash),
		ai.WithSystem(`
			You are a pirate translator.

			User will provide a message in English and your task is to translate it into pirate speech

			Important:
			* Keep length of your response similar to user's message.
		`),
		ai.WithTemperature(0.1),
		ai.WithMaxCompletionTokens(1000),
	),

	structured.WithAgentEvents(ai.NewLogAgentEvents(slog.Default())),
	structured.WithDescription("Translate to pirate speech."),
	structured.WithName("translate-to-pirate"),
)

func main() {
	ctx := context.Background()
	godotenv.Load()
	litellm := examples.GetLiteLLM()

	queries := []string{
		"How are you?",
		"I will kill you",
		"Hey",
		"Should we attack this ship?",
	}

	for _, q := range queries {
		if _, err := workflows.NewTypedWrapper[Translation](PiratesTask).
			Invoke(ctx, litellm, ai.NewHistory(
				ai.NewUserMessage(q),
			)); err != nil {
			panic(err)
		}
	}
}
