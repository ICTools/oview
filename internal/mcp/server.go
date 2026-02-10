package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/yourusername/oview/internal/config"
	"github.com/yourusername/oview/internal/logger"
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
		logger.Error("Failed to load project config", map[string]interface{}{
			"error": err.Error(),
			"hint":  "Run 'oview init' first",
		})
		return fmt.Errorf("failed to load project config: %w\nHint: Run 'oview init' first", err)
	}

	s.globalConfig, err = config.LoadGlobalConfig()
	if err != nil {
		logger.Error("Failed to load global config", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to load global config: %w", err)
	}

	// Initialize tool handler
	s.handler = NewToolHandler(s.projectConfig, s.globalConfig)

	logger.Info("MCP server ready to accept requests", map[string]interface{}{
		"project": s.projectConfig.ProjectSlug,
		"version": "0.2.0",
	})

	// MCP protocol: read from stdin, write to stdout
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		line := scanner.Bytes()

		var request MCPRequest
		if err := json.Unmarshal(line, &request); err != nil {
			logger.Error("Invalid JSON received", map[string]interface{}{
				"error": err.Error(),
				"raw":   string(line),
			})
			s.sendError(encoder, "", fmt.Errorf("invalid JSON: %w", err))
			continue
		}

		// Log incoming request
		logger.Debug("Received MCP request", map[string]interface{}{
			"method": request.Method,
			"id":     request.ID,
		})

		startTime := time.Now()
		response := s.handleRequest(&request)
		duration := time.Since(startTime)

		// Log response
		if response.Error != nil {
			logger.Warn("Request failed", map[string]interface{}{
				"method":   request.Method,
				"id":       request.ID,
				"error":    response.Error.Message,
				"duration": duration.String(),
			})
		} else {
			logger.Debug("Request handled successfully", map[string]interface{}{
				"method":   request.Method,
				"id":       request.ID,
				"duration": duration.String(),
			})
		}

		if err := encoder.Encode(response); err != nil {
			logger.Error("Failed to write response", map[string]interface{}{
				"error": err.Error(),
			})
			return fmt.Errorf("failed to write response: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Error("Scanner error", map[string]interface{}{
			"error": err.Error(),
		})
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
				"version": "0.2.0",
			},
		},
	}
}

// handleToolsList handles the tools/list request
func (s *Server) handleToolsList(req *MCPRequest) *MCPResponse {
	tools := []Tool{
		{
			Name:        "search",
			Description: "Search the codebase using semantic similarity with advanced filters and query strategies. Returns relevant code chunks with similarity scores.",
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
					"strategy": map[string]interface{}{
						"type":        "string",
						"description": "Search strategy: 'default', 'analysis' (broad context), 'debug' (focused on code), 'exploration' (diverse results), 'documentation' (docs-focused)",
						"enum":        []string{"default", "analysis", "debug", "exploration", "documentation"},
						"default":     "default",
					},
					"language": map[string]interface{}{
						"type":        "string",
						"description": "Filter by programming language (e.g., 'PHP', 'JavaScript')",
					},
					"languages": map[string]interface{}{
						"type":        "array",
						"description": "Filter by multiple languages",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Filter by chunk type: 'code', 'doc', 'config', 'test'",
						"enum":        []string{"code", "doc", "config", "test"},
					},
					"types": map[string]interface{}{
						"type":        "array",
						"description": "Filter by multiple chunk types",
						"items": map[string]interface{}{
							"type": "string",
							"enum": []string{"code", "doc", "config", "test"},
						},
					},
					"path_pattern": map[string]interface{}{
						"type":        "string",
						"description": "Filter by path pattern (supports * wildcard, e.g., 'src/Service/*')",
					},
					"component": map[string]interface{}{
						"type":        "string",
						"description": "Filter by component/module name",
					},
					"components": map[string]interface{}{
						"type":        "array",
						"description": "Filter by multiple components",
						"items": map[string]interface{}{
							"type": "string",
						},
					},
					"symbol_pattern": map[string]interface{}{
						"type":        "string",
						"description": "Filter by symbol pattern (function/class name, supports * wildcard)",
					},
					"min_similarity": map[string]interface{}{
						"type":        "number",
						"description": "Minimum similarity threshold (0.0-1.0)",
						"minimum":     0,
						"maximum":     1,
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
		logger.Error("Invalid tool call params", map[string]interface{}{
			"error": err.Error(),
		})
		return &MCPResponse{
			Jsonrpc: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: fmt.Sprintf("Invalid params: %v", err),
			},
		}
	}

	// Log tool call
	logger.Info("Calling tool", map[string]interface{}{
		"tool":      params.Name,
		"arguments": params.Arguments,
	})

	// Call the tool with context
	// TODO: Extract timeout from request params if provided
	ctx := context.Background()
	startTime := time.Now()
	result, err := s.handler.CallTool(ctx, params.Name, params.Arguments)
	duration := time.Since(startTime)

	if err != nil {
		logger.Error("Tool call failed", map[string]interface{}{
			"tool":     params.Name,
			"error":    err.Error(),
			"duration": duration.String(),
		})
		return &MCPResponse{
			Jsonrpc: "2.0",
			ID:      req.ID,
			Error: &MCPError{
				Code:    -32000,
				Message: err.Error(),
			},
		}
	}

	// Log success with result summary
	resultSummary := getResultSummary(params.Name, result)
	logger.Info("Tool call succeeded", map[string]interface{}{
		"tool":     params.Name,
		"duration": duration.String(),
		"summary":  resultSummary,
	})

	// Format result as MCP content blocks
	// MCP protocol requires results in a specific format
	resultJSON, _ := json.MarshalIndent(result, "", "  ")

	mcpResult := map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": string(resultJSON),
			},
		},
	}

	return &MCPResponse{
		Jsonrpc: "2.0",
		ID:      req.ID,
		Result:  mcpResult,
	}
}

// getResultSummary returns a summary of the tool result for logging
func getResultSummary(toolName string, result interface{}) string {
	switch toolName {
	case "search":
		if m, ok := result.(map[string]interface{}); ok {
			if count, ok := m["count"].(int); ok {
				return fmt.Sprintf("%d results", count)
			}
		}
	case "get_context":
		if m, ok := result.(map[string]interface{}); ok {
			if count, ok := m["count"].(int); ok {
				return fmt.Sprintf("%d chunks", count)
			}
		}
	case "project_info":
		return "project info retrieved"
	}
	return "completed"
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
