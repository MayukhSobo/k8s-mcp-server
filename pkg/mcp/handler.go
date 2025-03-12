package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mayukhsarkar/k8s-mcp-server/pkg/kubernetes"
	"github.com/mayukhsarkar/k8s-mcp-server/pkg/logs"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
)

// Handler handles MCP commands
type Handler struct {
	k8sClient  *kubernetes.Client
	logManager *logs.LogManager
}

// NewHandler creates a new MCP handler
func NewHandler(k8sClient *kubernetes.Client, clientset *kubernetes.Clientset) *Handler {
	return &Handler{
		k8sClient:  k8sClient,
		logManager: logs.NewLogManager(clientset),
	}
}

// HandleCommand processes an MCP command and returns a response
func (h *Handler) HandleCommand(cmd *Command) (*Response, error) {
	switch cmd.Type {
	case ListCommand:
		return h.handleListCommand(cmd)
	case GetCommand:
		return h.handleGetCommand(cmd)
	case CreateCommand:
		return h.handleCreateCommand(cmd)
	case DeleteCommand:
		return h.handleDeleteCommand(cmd)
	case LogsCommand:
		return h.handleLogsCommand(cmd)
	case SearchLogsCommand:
		return h.handleSearchLogsCommand(cmd)
	case ExportLogsCommand:
		return h.handleExportLogsCommand(cmd)
	default:
		return NewErrorResponse(fmt.Errorf("unsupported command type: %s", cmd.Type))
	}
}

// handleListCommand handles the 'list' command
func (h *Handler) handleListCommand(cmd *Command) (*Response, error) {
	if cmd.Resource == "" {
		return NewErrorResponse(fmt.Errorf("resource type is required"))
	}

	resources, err := h.k8sClient.ListResources(cmd.Resource, cmd.Namespace)
	if err != nil {
		return NewErrorResponse(err)
	}

	return NewSuccessResponse(fmt.Sprintf("Successfully listed %s", cmd.Resource), resources)
}

// handleGetCommand handles the 'get' command
func (h *Handler) handleGetCommand(cmd *Command) (*Response, error) {
	if cmd.Resource == "" || cmd.Name == "" {
		return NewErrorResponse(fmt.Errorf("resource type and name are required"))
	}

	resource, err := h.k8sClient.GetResource(cmd.Resource, cmd.Namespace, cmd.Name)
	if err != nil {
		return NewErrorResponse(err)
	}

	return NewSuccessResponse(fmt.Sprintf("Successfully retrieved %s '%s'", cmd.Resource, cmd.Name), resource)
}

// handleCreateCommand handles the 'create' command
func (h *Handler) handleCreateCommand(cmd *Command) (*Response, error) {
	if cmd.Resource == "" || cmd.Data == nil {
		return NewErrorResponse(fmt.Errorf("resource type and data are required"))
	}

	var obj unstructured.Unstructured
	if err := json.Unmarshal(cmd.Data, &obj.Object); err != nil {
		return NewErrorResponse(fmt.Errorf("invalid resource data: %v", err))
	}

	created, err := h.k8sClient.CreateResource(cmd.Resource, cmd.Namespace, &obj)
	if err != nil {
		return NewErrorResponse(err)
	}

	return NewSuccessResponse(fmt.Sprintf("Successfully created %s", cmd.Resource), created)
}

// handleDeleteCommand handles the 'delete' command
func (h *Handler) handleDeleteCommand(cmd *Command) (*Response, error) {
	if cmd.Resource == "" || cmd.Name == "" {
		return NewErrorResponse(fmt.Errorf("resource type and name are required"))
	}

	if err := h.k8sClient.DeleteResource(cmd.Resource, cmd.Namespace, cmd.Name); err != nil {
		return NewErrorResponse(err)
	}

	return NewSuccessResponse(fmt.Sprintf("Successfully deleted %s '%s'", cmd.Resource, cmd.Name), nil)
}

// handleLogsCommand handles the 'logs' command
func (h *Handler) handleLogsCommand(cmd *Command) (*Response, error) {
	if cmd.Namespace == "" || cmd.LogOptions == nil || cmd.LogOptions.Pod == "" {
		return NewErrorResponse(fmt.Errorf("namespace and pod are required"))
	}

	opts := logs.LogOptions{
		Namespace: cmd.Namespace,
		Pod:       cmd.LogOptions.Pod,
		Container: cmd.LogOptions.Container,
	}

	// Parse 'since' parameter if provided
	if cmd.LogOptions.Since != "" {
		if since, err := time.ParseDuration(cmd.LogOptions.Since); err == nil {
			sinceTime := time.Now().Add(-since)
			opts.SinceTime = &sinceTime
		} else {
			// Try parsing as a timestamp
			if sinceTime, err := time.Parse(time.RFC3339, cmd.LogOptions.Since); err == nil {
				opts.SinceTime = &sinceTime
			} else {
				return NewErrorResponse(fmt.Errorf("invalid 'since' parameter: %v", err))
			}
		}
	}

	// Parse 'tail' parameter if provided
	if cmd.LogOptions.Tail > 0 {
		tail := int64(cmd.LogOptions.Tail)
		opts.Tail = &tail
	}

	logEntries, err := h.logManager.GetLogs(opts)
	if err != nil {
		return NewErrorResponse(err)
	}

	return NewSuccessResponse(fmt.Sprintf("Successfully retrieved logs from pod '%s'", cmd.LogOptions.Pod), logEntries)
}

// handleSearchLogsCommand handles the 'search_logs' command
func (h *Handler) handleSearchLogsCommand(cmd *Command) (*Response, error) {
	if cmd.Namespace == "" || cmd.LogOptions == nil || cmd.LogOptions.Pod == "" {
		return NewErrorResponse(fmt.Errorf("namespace and pod are required"))
	}

	if cmd.LogOptions.Pattern == "" {
		return NewErrorResponse(fmt.Errorf("search pattern is required"))
	}

	opts := logs.LogOptions{
		Namespace: cmd.Namespace,
		Pod:       cmd.LogOptions.Pod,
		Container: cmd.LogOptions.Container,
		Pattern:   cmd.LogOptions.Pattern,
		LogLevel:  cmd.LogOptions.LogLevel,
	}

	// Parse 'since' parameter if provided
	if cmd.LogOptions.Since != "" {
		if since, err := time.ParseDuration(cmd.LogOptions.Since); err == nil {
			sinceTime := time.Now().Add(-since)
			opts.SinceTime = &sinceTime
		} else {
			// Try parsing as a timestamp
			if sinceTime, err := time.Parse(time.RFC3339, cmd.LogOptions.Since); err == nil {
				opts.SinceTime = &sinceTime
			} else {
				return NewErrorResponse(fmt.Errorf("invalid 'since' parameter: %v", err))
			}
		}
	}

	// Parse 'tail' parameter if provided
	if cmd.LogOptions.Tail > 0 {
		tail := int64(cmd.LogOptions.Tail)
		opts.Tail = &tail
	}

	logEntries, err := h.logManager.GetLogs(opts)
	if err != nil {
		return NewErrorResponse(err)
	}

	return NewSuccessResponse(fmt.Sprintf("Successfully searched logs from pod '%s'", cmd.LogOptions.Pod), logEntries)
}

// handleExportLogsCommand handles the 'export_logs' command
func (h *Handler) handleExportLogsCommand(cmd *Command) (*Response, error) {
	if cmd.Namespace == "" || cmd.LogOptions == nil || cmd.LogOptions.Pod == "" {
		return NewErrorResponse(fmt.Errorf("namespace and pod are required"))
	}

	if cmd.LogOptions.Format == "" {
		return NewErrorResponse(fmt.Errorf("export format is required"))
	}

	opts := logs.LogOptions{
		Namespace: cmd.Namespace,
		Pod:       cmd.LogOptions.Pod,
		Container: cmd.LogOptions.Container,
		Pattern:   cmd.LogOptions.Pattern,
		LogLevel:  cmd.LogOptions.LogLevel,
	}

	// Parse 'since' parameter if provided
	if cmd.LogOptions.Since != "" {
		if since, err := time.ParseDuration(cmd.LogOptions.Since); err == nil {
			sinceTime := time.Now().Add(-since)
			opts.SinceTime = &sinceTime
		} else {
			// Try parsing as a timestamp
			if sinceTime, err := time.Parse(time.RFC3339, cmd.LogOptions.Since); err == nil {
				opts.SinceTime = &sinceTime
			} else {
				return NewErrorResponse(fmt.Errorf("invalid 'since' parameter: %v", err))
			}
		}
	}

	// Parse 'tail' parameter if provided
	if cmd.LogOptions.Tail > 0 {
		tail := int64(cmd.LogOptions.Tail)
		opts.Tail = &tail
	}

	logEntries, err := h.logManager.GetLogs(opts)
	if err != nil {
		return NewErrorResponse(err)
	}

	var buf bytes.Buffer
	if err := h.logManager.ExportLogs(logEntries, cmd.LogOptions.Format, &buf); err != nil {
		return NewErrorResponse(err)
	}

	// Create a response with the exported logs as a string
	return NewSuccessResponse(
		fmt.Sprintf("Successfully exported logs from pod '%s' in %s format", cmd.LogOptions.Pod, cmd.LogOptions.Format),
		map[string]string{"exported_logs": buf.String()},
	)
}
