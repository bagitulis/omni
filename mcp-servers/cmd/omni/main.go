// Omni Management MCP Server
package main

import (
	"fmt"
	"os"

	"mcp-servers/pkg/mcp"
)

func main() {
	var err error
	omniClient, err = NewOmniClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize client: %v\n", err)
		os.Exit(1)
	}

	server := mcp.NewServer("mcp-omni", "1.0.0")
	server.RegisterTools(buildToolDefinitions())

	server.SetHandler(func(name string, args map[string]interface{}) (interface{}, error) {
		return handleToolCall(name, args)
	})

	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
