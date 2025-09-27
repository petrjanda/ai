package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/getsynq/cloud/ai-data-sre/examples"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/adapters/openai"
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/workflows"

	"github.com/joho/godotenv"
)

// OpenAI Schema Generator
var schemaGenerator = openai.NewOpenAISchemaGenerator()

// TYPES

type Flight struct {
	FlightNumber     string `json:"flight_number" jsonschema:"required"`
	DepartureTime    string `json:"departure_time" jsonschema:"required"`
	ArrivalTime      string `json:"arrival_time" jsonschema:"required"`
	DepartureAirport string `json:"departure_airport" jsonschema:"required"`
	ArrivalAirport   string `json:"arrival_airport" jsonschema:"required"`
	Price            int    `json:"price" jsonschema:"required"`
}

var formatter = workflows.NewStructuredTask[Flight](
	"formatter",
	ai.NewLLMRequest(
		ai.WithModel(ai.Gemini25Flash),
		ai.WithSystem(`
				Generate example flights.
			`),
		ai.WithTemperature(1.0),
	),
)

func main() {
	ctx := context.Background()
	godotenv.Load()
	litellm := examples.GetLiteLLM()

	prompt := "Generate flight from Asia to US."
	confirmations, err := workflows.Typed[Flight](formatter).
		InvokeTyped(ctx, litellm, ai.NewHistory(ai.NewUserMessage(prompt)))

	if err != nil {
		panic(err)
	}

	payload, err := json.MarshalIndent(confirmations, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(payload))
}
