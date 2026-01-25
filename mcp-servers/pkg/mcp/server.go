// Package mcp provides common MCP protocol implementation for all servers
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

// InputSchema defines tool input schema
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

// Property defines a schema property
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Items       *Items `json:"items,omitempty"`
}

// Items defines array items schema
type Items struct {
	Type string `json:"type"`
}

// TextContent represents text content in MCP response
type TextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolResult represents a tool call result
type ToolResult struct {
	Content []TextContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolHandler is the function signature for handling tool calls
type ToolHandler func(name string, args map[string]interface{}) (interface{}, error)

// Server handles MCP communication
type Server struct {
	name    string
	version string
	tools   []Tool
	handler ToolHandler
}

// NewServer creates a new MCP server
func NewServer(name, version string) *Server {
	return &Server{
		name:    name,
		version: version,
	}
}

// RegisterTools sets the available tools
func (s *Server) RegisterTools(tools []Tool) {
	s.tools = tools
}

// SetHandler sets the tool call handler
func (s *Server) SetHandler(handler ToolHandler) {
	s.handler = handler
}

// Run starts the MCP server
func (s *Server) Run() error {
	fmt.Fprintf(os.Stderr, "%s v%s running\n", s.name, s.version)

	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, len(buf))

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			s.sendError(nil, -32700, "Parse error", err.Error())
			continue
		}

		s.handleRequest(req)
	}

	return scanner.Err()
}

func (s *Server) handleRequest(req JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "tools/list":
		s.handleListTools(req)
	case "tools/call":
		s.handleCallTool(req)
	default:
		s.sendError(req.ID, -32601, "Method not found", req.Method)
	}
}

func (s *Server) handleInitialize(req JSONRPCRequest) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    s.name,
			"version": s.version,
		},
	}
	s.sendResult(req.ID, result)
}

func (s *Server) handleListTools(req JSONRPCRequest) {
	result := map[string]interface{}{
		"tools": s.tools,
	}
	s.sendResult(req.ID, result)
}

func (s *Server) handleCallTool(req JSONRPCRequest) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.sendError(req.ID, -32602, "Invalid params", err.Error())
		return
	}

	if s.handler == nil {
		s.sendError(req.ID, -32603, "No handler", "No tool handler registered")
		return
	}

	result, err := s.handler(params.Name, params.Arguments)
	if err != nil {
		toolResult := ToolResult{
			Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Error: %v", err)}},
			IsError: true,
		}
		s.sendResult(req.ID, toolResult)
		return
	}

	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		toolResult := ToolResult{
			Content: []TextContent{{Type: "text", Text: fmt.Sprintf("Error marshaling result: %v", err)}},
			IsError: true,
		}
		s.sendResult(req.ID, toolResult)
		return
	}

	toolResult := ToolResult{
		Content: []TextContent{{Type: "text", Text: string(jsonBytes)}},
	}
	s.sendResult(req.ID, toolResult)
}

func (s *Server) sendResult(id interface{}, result interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	s.sendResponse(resp)
}

func (s *Server) sendError(id interface{}, code int, message string, data interface{}) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	s.sendResponse(resp)
}

func (s *Server) sendResponse(resp JSONRPCResponse) {
	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling response: %v\n", err)
		return
	}
	fmt.Println(string(jsonBytes))
}
