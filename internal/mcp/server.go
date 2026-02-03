package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/yourusername/oview/internal/config"
)

// Server is an MCP server that exposes oview RAG capabilities
type Server struct {
	projectPath   string
	projectConfig *config.ProjectConfig
	globalConfig  *config.GlobalConfig
	handler       *ToolHandler
}

// NewServer creates a new MCP server
func NewServer(projectPath string) *Server {
	return &Server{
		projectPath: projectPath,
	}
}

// Run starts the MCP server (stdio mode)
func (s *Server) Run() error {
	// Load configs
	var err error
	s.projectConfig, err = config.LoadProjectConfig(s.projectPath)
	if err != nil {
		return fmt.Errorf("failed to load project config: %w\nHint: Run 'oview init' first", err)
	}

	s.globalConfig, err = config.LoadGlobalConfig()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	// Initialize tool handler
	s.handler = NewToolHandler(s.projectConfig, s.globalConfig)

	// MCP protocol: read from stdin, write to stdout
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		line := scanner.Bytes()

		var request MCPRequest
		if err := json.Unmarshal(line, &request); err != nil {
			s.sendError(encoder, "", fmt.Errorf("invalid JSON: %w", err))
			continue
		}

		response := s.handleRequest(&request)
		if err := encoder.Encode(response); err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanner error: %w", err)
	}

	return nil
}

// handleRequest processes an MCP request
func (s *Server) handleRequest(req *MCPRequest) *MCPResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		return &MCPResponse{
			Jsonrpc: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: fmt.Sprintf("Method not found: %s", req.Method),
			},
		}
	}
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(req *MCPRequest) *MCPResponse {
	return &MCPResponse{
		Jsonrpc: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "oview",
				"version": "0.1.0",
			},
		},
	}
}

// handleToolsList handles the tools/list request
func (s *Server) handleToolsList(req *MCPRequest) *MCPResponse {
	tools := []Tool{
		{
			Name:        "search",
			Description: "Search the codebase using semantic similarity. Returns relevant code chunks with similarity scores.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "The search query (e.g., 'authentication logic', 'database connection', 'error handling')",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Number of results to return (default: 5, max: 20)",
						"default":     5,
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "get_context",
			Description: "Get relevant code context for a specific file or symbol. Useful before making changes to understand related code.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "File path to get context for",
					},
					"symbol": map[string]interface{}{
						"type":        "string",
						"description": "Optional: specific function/class name to focus on",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Number of related chunks to return (default: 3)",
						"default":     3,
					},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "project_info",
			Description: "Get information about the current project (stack, embeddings config, database status)",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	return &MCPResponse{
		Jsonrpc: "2.0",
		ID:      req.ID,
		Result: map[string]interface{}{
			"tools": tools,
		},
	}
}

// handleToolsCall handles the tools/call request
func (s *Server) handleToolsCall(req *MCPRequest) *MCPResponse {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	// Parse params
	paramsBytes, _ := json.Marshal(req.Params)
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return &MCPResponse{
			Jsonrpc: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: fmt.Sprintf("Invalid params: %v", err),
			},
		}
	}

	// Call the tool
	result, err := s.handler.CallTool(params.Name, params.Arguments)
	if err != nil {
		return &MCPResponse{
			Jsonrpc: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32000,
				Message: err.Error(),
			},
		}
	}

	return &MCPResponse{
		Jsonrpc: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

// sendError sends an error response
func (s *Server) sendError(encoder *json.Encoder, id interface{}, err error) {
	response := &MCPResponse{
		Jsonrpc: "2.0",
		ID:      id,
		Error: &MCPError{
			Code:    -32603,
			Message: err.Error(),
		},
	}
	encoder.Encode(response)
}

// MCPRequest represents an MCP JSON-RPC request
type MCPRequest struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPResponse represents an MCP JSON-RPC response
type MCPResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Tool represents an MCP tool definition
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}
