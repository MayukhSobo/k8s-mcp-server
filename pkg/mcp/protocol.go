package mcp

import (
	"encoding/json"
	"fmt"
)

// CommandType represents the type of MCP command
type CommandType string

const (
	// Kubernetes resource operations
	ListCommand   CommandType = "list"
	GetCommand    CommandType = "get"
	CreateCommand CommandType = "create"
	DeleteCommand CommandType = "delete"

	// Log operations
	LogsCommand       CommandType = "logs"
	SearchLogsCommand CommandType = "search_logs"
	ExportLogsCommand CommandType = "export_logs"
)

// Command represents an MCP command
type Command struct {
	Type       CommandType     `json:"type"`
	Resource   string          `json:"resource,omitempty"`
	Name       string          `json:"name,omitempty"`
	Namespace  string          `json:"namespace,omitempty"`
	Data       json.RawMessage `json:"data,omitempty"`
	LogOptions *LogOptions     `json:"log_options,omitempty"`
}

// LogOptions represents options for log commands
type LogOptions struct {
	Pod       string `json:"pod,omitempty"`
	Container string `json:"container,omitempty"`
	Since     string `json:"since,omitempty"`
	Tail      int    `json:"tail,omitempty"`
	Pattern   string `json:"pattern,omitempty"`
	LogLevel  string `json:"log_level,omitempty"`
	Format    string `json:"format,omitempty"`
}

// Response represents an MCP response
type Response struct {
	Success bool            `json:"success"`
	Message string          `json:"message,omitempty"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(message string, data interface{}) (*Response, error) {
	var rawData json.RawMessage
	var err error

	if data != nil {
		rawData, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response data: %v", err)
		}
	}

	return &Response{
		Success: true,
		Message: message,
		Data:    rawData,
	}, nil
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err error) (*Response, error) {
	return &Response{
		Success: false,
		Error:   err.Error(),
	}, nil
}

// ParseCommand parses a JSON string into a Command
func ParseCommand(data []byte) (*Command, error) {
	var cmd Command
	if err := json.Unmarshal(data, &cmd); err != nil {
		return nil, fmt.Errorf("failed to parse command: %v", err)
	}
	return &cmd, nil
}
