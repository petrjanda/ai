package main

import (
	"context"
	"log"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	ctx := context.Background()

	// Create a new client, with no features.
	client := mcp.NewClient(&mcp.Implementation{Name: "mcp-client", Version: "v1.0.0"}, nil)

	// Connect to a server over stdin/stdout
	transport := &mcp.CommandTransport{Command: exec.Command("./server")}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	// transport := &mcp.StreamableClientTransport{Endpoint: "http://localhost:8080"}
	// session, err := client.Connect(ctx, transport, nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer session.Close()

	// Call a tool on the server.
	params := &mcp.CallToolParams{
		Name:      "cityTime",
		Arguments: map[string]any{"city": "nyc"},
	}
	res, err := session.CallTool(ctx, params)
	if err != nil {
		log.Fatalf("CallTool failed: %v", err)
	}
	if res.IsError {
		log.Fatal("tool failed")
	}
	for _, c := range res.Content {
		log.Print(c.(*mcp.TextContent).Text)
	}
}
