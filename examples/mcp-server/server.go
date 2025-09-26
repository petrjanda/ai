package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GetTimeParams defines the parameters for the cityTime tool.
type GetTimeParams struct {
	City string `json:"city" jsonschema:"City to get time for (nyc, sf, or boston)"`
}

// getTime implements the tool that returns the current time for a given city.
func getTime(ctx context.Context, req *mcp.CallToolRequest, params *GetTimeParams) (*mcp.CallToolResult, any, error) {
	// Define time zones for each city
	locations := map[string]string{
		"nyc":    "America/New_York",
		"sf":     "America/Los_Angeles",
		"boston": "America/New_York",
	}

	city := params.City
	if city == "" {
		city = "nyc" // Default to NYC
	}

	// Get the timezone.
	tzName, ok := locations[city]
	if !ok {
		return nil, nil, fmt.Errorf("unknown city: %s", city)
	}

	// Load the location.
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load timezone: %w", err)
	}

	// Get current time in that location.
	now := time.Now().In(loc)

	// Format the response.
	cityNames := map[string]string{
		"nyc":    "New York City",
		"sf":     "San Francisco",
		"boston": "Boston",
	}

	response := fmt.Sprintf("The current time in %s is %s",
		cityNames[city],
		now.Format(time.RFC3339))

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: response},
		},
	}, nil, nil
}

type GreetParams struct {
	Name string `json:"name" jsonschema:"Name to greet"`
}

func greet(ctx context.Context, req *mcp.CallToolRequest, params *GreetParams) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: "Hello, " + params.Name},
		},
	}, nil, nil
}

func main() {
	// Create an MCP server.
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "time-server",
		Version: "1.0.0",
	}, nil)

	// Add the cityTime tool.
	mcp.AddTool(server, &mcp.Tool{
		Name:        "cityTime",
		Description: "Get the current time in NYC, San Francisco, or Boston",
	}, getTime)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "greet",
		Description: "Greet a person",
	}, greet)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}

	// // Create the streamable HTTP handler.
	// handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
	// 	return server
	// }, nil)

	// url := "localhost:8080"

	// log.Printf("MCP server listening on %s", url)
	// log.Printf("Available tool: cityTime (cities: nyc, sf, boston)")

	// // Start the HTTP server with logging handler.
	// if err := http.ListenAndServe(url, handler); err != nil {
	// 	log.Fatalf("Server failed: %v", err)
	// }
}
