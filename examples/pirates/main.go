package main

import (
	"context"
	"fmt"

	"github.com/getsynq/cloud/ai-data-sre/examples"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/adapters/openai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/workflows"
	"github.com/joho/godotenv"
)

type Translation struct {
	// Input message
	In string `json:"in" jsonschema:"required"`

	// Output message
	Out string `json:"out" jsonschema:"required"`
}

var TranslationSchema = openai.NewOpenAISchemaGenerator().MustGenerate(&Translation{})

var PiratesTask = workflows.NewStructuredTask[Translation](
	"arrrr",
	ai.NewLLMRequest(
		ai.WithModel(ai.Claude3Sonnet),
		ai.WithSystem(`
			You are a pirate translator.

			User will provide a message in English and your task is to translate it into pirate speech.

			Important:
			* Keep length of your response similar to user's message.
		`),
		ai.WithTemperature(1.0),
		ai.WithMaxCompletionTokens(1000),
	),
)

var ReversePiratesTask = workflows.NewStructuredTask[Translation](
	"no-arrrr",
	ai.NewLLMRequest(
		ai.WithModel(ai.Gemini25Pro),
		ai.WithSystem(`
			You are a pirate translator.

			User will provide a message in pirate speech and your task is to translate it into English.

			Important:
			* Keep length of your response similar to user's message.
		`),
		ai.WithTemperature(1.0),
		ai.WithMaxCompletionTokens(1000),
	),
)

func main() {
	ctx := context.Background()
	godotenv.Load()
	litellm := examples.GetLiteLLM()

	queries := []string{
		"How are you?",
		"I love you",
		"Hey",
		"Should we attack this ship?",
	}

	for _, q := range queries {
		// Translate to pirate
		pirates, _ := workflows.NewTypedWrapper[Translation](PiratesTask).
			Invoke(ctx, litellm, ai.NewHistory(
				ai.NewUserMessage(q),
			))

		// Print the conversation

		fmt.Printf("%s ==> %s\n", pirates.In, pirates.Out)

		// Translate back to English
		englishman, _ := workflows.NewTypedWrapper[Translation](ReversePiratesTask).
			Invoke(ctx, litellm, ai.NewHistory(
				ai.NewUserMessage(pirates.Out),
			))

		fmt.Printf("%s ==> %s\n", englishman.In, englishman.Out)
		fmt.Println("========================================")
	}
}
