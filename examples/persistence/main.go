package main

import (
	"context"
	"encoding/json"
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
	"github.com/getsynq/cloud/ai-data-sre/pkg/ai/workflows"
	"github.com/joho/godotenv"
)

// OpenAI Schema Generator
var schemaGenerator = openai.NewOpenAISchemaGenerator()

var logger = ai.NewLogAgentEvents(slog.Default())

// TYPES

type Flight struct {
	FlightNumber string `json:"flight_number"`
	Price        int    `json:"price"`
}

type Booking struct {
	ConfirmationNumber string `json:"confirmation_number"`
	FlightNumber       string `json:"flight_number"`
	Price              int    `json:"price"`
	DepartureDate      string `json:"departure_date"`
}

type Itinerary struct {
	Outbound Booking `json:"outbound" jsonschema:"required"`
	Inbound  Booking `json:"inbound" jsonschema:"required"`
}

// SEARCH

type SearchFlightsRequest struct {
	Destination string `json:"destination" jsonschema:"required"`
	Date        string `json:"date" jsonschema:"required"`
}

type SearchFlightsResponse struct {
	Flights []Flight `json:"flights"`
}

var searchFlightsTool = tools.NewSimpleTool(
	"SearchFlights",
	"Searches for flights for the user",

	func(ctx context.Context, args *SearchFlightsRequest) (*SearchFlightsResponse, error) {
		return &SearchFlightsResponse{
			Flights: []Flight{
				{FlightNumber: "US23456", Price: 100},
				{FlightNumber: "SK23456", Price: 200},
			},
		}, nil
	},
)

// BOOK

type BookRequest struct {
	FlightNumber string `json:"flight_number"`
}

type BookFlightTool struct {
}

func NewBookFlightTool() *BookFlightTool {
	return &BookFlightTool{}
}

func (b *BookFlightTool) Name() string {
	return "BookFlight"
}

func (b *BookFlightTool) Description() string {
	return "Books a flight for the user"
}

func (b *BookFlightTool) InputSchemaRaw() json.RawMessage {
	return schemaGenerator.MustGenerate(new(BookRequest))
}

func (b *BookFlightTool) Run(ctx context.Context, args *BookRequest) (*Booking, error) {
	return &Booking{
		ConfirmationNumber: "#ARK495",
	}, nil
}

// AGENT

var travelAgent = workflows.NewAgentTask(
	"travel",
	ai.NewLLMRequest(
		ai.WithSystem(`
			You search and book flights for a user.

			User will provide a destination and dates of travel.

			You will use the following tools to book the flight:
			* BookFlight - books a flight for the user
			* SearchFlights - searches for flights for the user

			Each flight is separate which means there should be two confirmation numbers.

			You should search flights and then book the cheapest 
			one and return the confirmation numbers, flight numbers, departure dates and prices for outbound and inbound flights.
		`),
		ai.WithModel(ai.Gemini25Flash),
		ai.WithTemperature(0.0),
		ai.WithMaxCompletionTokens(1000),
		ai.WithTools(tools.NewAdapter(NewBookFlightTool()), searchFlightsTool),
	),
	agent.WithEvents(logger),
)

// FINAL FORMATTER

var formatter = workflows.NewStructuredTask[Itinerary](
	"formatter",
	ai.NewLLMRequest(
		ai.WithSystem(`
			Format flight information into specific format.
		`),
		ai.WithModel(ai.Gemini25Flash),
		ai.WithTemperature(0.0),
	),
	structured.WithAgentEvents(logger),
)

type WorkflowInput struct {
	Prompt string
}

var bookFlights = workflows.NewFunctionWork("book-flights",
	func(ctx context.Context, llm ai.LLM, in *string) (*Itinerary, error) {
		return workflows.Typed[Itinerary](
			travelAgent.
				Pipe(formatter),
		).
			InvokeTyped(ctx, llm, ai.NewHistory(
				ai.NewUserMessage(*in),
			))
	},
)

func main() {
	ctx := context.Background()
	godotenv.Load()
	litellm := examples.GetLiteLLM()

	memory := workflows.NewMemoryStorageProvider()
	ctx = workflows.WithStorage(ctx, memory.Storage("travel-1"))

	prompt := "I want to book a flight to Tokyo, 1st Oct and back 8th Oct"

	{

		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		confirmations, err := bookFlights.
			Invoke(ctx, litellm, &prompt)

		if err != nil {
			slog.Error("error", "error", err)
		}

		payload, err := json.MarshalIndent(confirmations, "", "  ")
		if err != nil {
			panic(err)
		}

		fmt.Println(string(payload))
	}

	slog.Info("First run")

	{
		confirmations, err := bookFlights.
			Invoke(ctx, litellm, &prompt)

		if err != nil {
			panic(err)
		}

		payload, err := json.MarshalIndent(confirmations, "", "  ")
		if err != nil {
			panic(err)
		}

		log.Println(string(payload))
	}

	slog.Info("Second run")

	// {
	// 	"outbound": {
	// 	  "confirmation_number": "#ARK495",
	// 	  "flight_number": "US23456",
	// 	  "price": 100,
	// 	  "departure_date": "October 1st"
	// 	},
	// 	"inbound": {
	// 	  "confirmation_number": "#ARK495",
	// 	  "flight_number": "US23456",
	// 	  "price": 100,
	// 	  "departure_date": "October 8th"
	// 	}
	// }

}
