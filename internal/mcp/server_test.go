package mcp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	server := NewServer("/test/path")

	assert.NotNil(t, server)
	assert.Equal(t, "/test/path", server.projectPath)
}

func TestServer_handleRequest_Initialize(t *testing.T) {
	server := &Server{
		projectPath: "/test/path",
	}

	req := &MCPRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "initialize",
	}

	response := server.handleRequest(req)

	assert.NotNil(t, response)
	assert.Equal(t, "2.0", response.Jsonrpc)
	assert.Equal(t, 1, response.ID)
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Result)

	result, ok := response.Result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "2024-11-05", result["protocolVersion"])

	serverInfo, ok := result["serverInfo"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "oview", serverInfo["name"])
	assert.Equal(t, "0.2.0", serverInfo["version"])
}

func TestServer_handleRequest_ToolsList(t *testing.T) {
	server := &Server{
		projectPath: "/test/path",
	}

	req := &MCPRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	response := server.handleRequest(req)

	assert.NotNil(t, response)
	assert.Equal(t, "2.0", response.Jsonrpc)
	assert.Equal(t, 1, response.ID)
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Result)

	result, ok := response.Result.(map[string]interface{})
	assert.True(t, ok)

	tools, ok := result["tools"].([]Tool)
	assert.True(t, ok)
	assert.Equal(t, 3, len(tools))

	// Verify tools
	toolNames := []string{}
	for _, tool := range tools {
		toolNames = append(toolNames, tool.Name)
	}
	assert.Contains(t, toolNames, "search")
	assert.Contains(t, toolNames, "get_context")
	assert.Contains(t, toolNames, "project_info")

	// Verify search tool schema
	searchTool := tools[0]
	assert.Equal(t, "search", searchTool.Name)
	assert.NotEmpty(t, searchTool.Description)
	assert.NotNil(t, searchTool.InputSchema)

	schema := searchTool.InputSchema
	props, ok := schema["properties"].(map[string]interface{})
	assert.True(t, ok)
	assert.NotNil(t, props["query"])
	assert.NotNil(t, props["limit"])
	assert.NotNil(t, props["strategy"])
	assert.NotNil(t, props["language"])
	assert.NotNil(t, props["type"])
	assert.NotNil(t, props["path_pattern"])
}

func TestServer_handleRequest_UnknownMethod(t *testing.T) {
	server := &Server{
		projectPath: "/test/path",
	}

	req := &MCPRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "unknown_method",
	}

	response := server.handleRequest(req)

	assert.NotNil(t, response)
	assert.Equal(t, "2.0", response.Jsonrpc)
	assert.Equal(t, 1, response.ID)
	assert.NotNil(t, response.Error)
	assert.Equal(t, -32601, response.Error.Code)
	assert.Contains(t, response.Error.Message, "Method not found")
}

func TestServer_handleInitialize(t *testing.T) {
	server := &Server{}

	req := &MCPRequest{
		Jsonrpc: "2.0",
		ID:      "test-id",
		Method:  "initialize",
	}

	response := server.handleInitialize(req)

	assert.NotNil(t, response)
	assert.Equal(t, "2.0", response.Jsonrpc)
	assert.Equal(t, "test-id", response.ID)
	assert.Nil(t, response.Error)

	result, ok := response.Result.(map[string]interface{})
	assert.True(t, ok)

	capabilities, ok := result["capabilities"].(map[string]interface{})
	assert.True(t, ok)
	assert.NotNil(t, capabilities["tools"])
}

func TestGetResultSummary(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		result   interface{}
		expected string
	}{
		{
			name:     "search with results",
			toolName: "search",
			result: map[string]interface{}{
				"count": 5,
			},
			expected: "5 results",
		},
		{
			name:     "get_context with chunks",
			toolName: "get_context",
			result: map[string]interface{}{
				"count": 3,
			},
			expected: "3 chunks",
		},
		{
			name:     "project_info",
			toolName: "project_info",
			result:   map[string]interface{}{},
			expected: "project info retrieved",
		},
		{
			name:     "unknown tool",
			toolName: "unknown",
			result:   map[string]interface{}{},
			expected: "completed",
		},
		{
			name:     "search with no count",
			toolName: "search",
			result:   map[string]interface{}{},
			expected: "completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getResultSummary(tt.toolName, tt.result)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMCPRequest_Unmarshal(t *testing.T) {
	jsonData := `{
		"jsonrpc": "2.0",
		"id": 1,
		"method": "tools/call",
		"params": {
			"name": "search",
			"arguments": {
				"query": "test"
			}
		}
	}`

	var req MCPRequest
	err := json.Unmarshal([]byte(jsonData), &req)

	assert.NoError(t, err)
	assert.Equal(t, "2.0", req.Jsonrpc)
	assert.Equal(t, float64(1), req.ID) // JSON numbers become float64
	assert.Equal(t, "tools/call", req.Method)
	assert.NotNil(t, req.Params)
}

func TestMCPResponse_Marshal(t *testing.T) {
	response := &MCPResponse{
		Jsonrpc: "2.0",
		ID:      1,
		Result: map[string]interface{}{
			"success": true,
		},
	}

	data, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "jsonrpc")
	assert.Contains(t, string(data), "2.0")
	assert.Contains(t, string(data), "success")
}

func TestMCPError_Marshal(t *testing.T) {
	response := &MCPResponse{
		Jsonrpc: "2.0",
		ID:      1,
		Error: &MCPError{
			Code:    -32600,
			Message: "Invalid request",
		},
	}

	data, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "error")
	assert.Contains(t, string(data), "-32600")
	assert.Contains(t, string(data), "Invalid request")
	assert.NotContains(t, string(data), "result") // Should not have result field
}

func TestTool_Schema(t *testing.T) {
	tool := Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param1": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"param1"},
		},
	}

	assert.Equal(t, "test_tool", tool.Name)
	assert.Equal(t, "A test tool", tool.Description)
	assert.NotNil(t, tool.InputSchema)

	schema := tool.InputSchema
	assert.Equal(t, "object", schema["type"])
	assert.NotNil(t, schema["properties"])
	assert.NotNil(t, schema["required"])
}
