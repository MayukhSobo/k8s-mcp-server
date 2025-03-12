package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mayukhsarkar/k8s-mcp-server/pkg/kubernetes"
	"github.com/mayukhsarkar/k8s-mcp-server/pkg/mcp"
)

// Server represents the HTTP API server
type Server struct {
	port       int
	k8sClient  *kubernetes.Client
	mcpHandler *mcp.Handler
}

// NewServer creates a new HTTP API server
func NewServer(port int, kubeconfigPath string) *Server {
	// Create Kubernetes client
	k8sClient, err := kubernetes.NewClient(kubeconfigPath)
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Get the clientset from the client
	clientset := k8sClient.GetClientset()

	// Create MCP handler
	mcpHandler := mcp.NewHandler(k8sClient, clientset)

	return &Server{
		port:       port,
		k8sClient:  k8sClient,
		mcpHandler: mcpHandler,
	}
}

// Start starts the HTTP API server
func (s *Server) Start() error {
	// Register API routes
	http.HandleFunc("/api/v1/mcp", s.handleMCPRequest)
	http.HandleFunc("/api/v1/resources/", s.handleResourceRequest)
	http.HandleFunc("/api/v1/logs/", s.handleLogRequest)
	http.HandleFunc("/health", s.handleHealthCheck)

	// Start the server
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting server on %s", addr)
	return http.ListenAndServe(addr, nil)
}

// handleMCPRequest handles MCP protocol requests
func (s *Server) handleMCPRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	// Parse MCP command
	cmd, err := mcp.ParseCommand(body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse command: %v", err), http.StatusBadRequest)
		return
	}

	// Handle command
	resp, err := s.mcpHandler.HandleCommand(cmd)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to handle command: %v", err), http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if !resp.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(resp)
}

// handleResourceRequest handles Kubernetes resource requests
func (s *Server) handleResourceRequest(w http.ResponseWriter, r *http.Request) {
	// Parse path: /api/v1/resources/{resource_type}/{name}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	resourceType := parts[3]
	var name string
	if len(parts) > 4 {
		name = parts[4]
	}

	// Get namespace from query parameters
	namespace := r.URL.Query().Get("namespace")

	// Create MCP command based on HTTP method
	var cmd *mcp.Command

	switch r.Method {
	case http.MethodGet:
		if name == "" {
			// List resources
			cmd = &mcp.Command{
				Type:      mcp.ListCommand,
				Resource:  resourceType,
				Namespace: namespace,
			}
		} else {
			// Get resource
			cmd = &mcp.Command{
				Type:      mcp.GetCommand,
				Resource:  resourceType,
				Name:      name,
				Namespace: namespace,
			}
		}
	case http.MethodPost:
		// Create resource
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
			return
		}

		cmd = &mcp.Command{
			Type:      mcp.CreateCommand,
			Resource:  resourceType,
			Namespace: namespace,
			Data:      body,
		}
	case http.MethodDelete:
		// Delete resource
		if name == "" {
			http.Error(w, "Resource name is required for DELETE", http.StatusBadRequest)
			return
		}

		cmd = &mcp.Command{
			Type:      mcp.DeleteCommand,
			Resource:  resourceType,
			Name:      name,
			Namespace: namespace,
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Handle command
	resp, err := s.mcpHandler.HandleCommand(cmd)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to handle command: %v", err), http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if !resp.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(resp)
}

// handleLogRequest handles log requests
func (s *Server) handleLogRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse path: /api/v1/logs/{namespace}/{pod}
	// or /api/v1/logs/search or /api/v1/logs/export
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	container := r.URL.Query().Get("container")
	since := r.URL.Query().Get("since")
	tailStr := r.URL.Query().Get("tail")
	pattern := r.URL.Query().Get("pattern")
	logLevel := r.URL.Query().Get("level")
	format := r.URL.Query().Get("format")

	// Parse tail parameter
	var tail int
	if tailStr != "" {
		var err error
		tail, err = strconv.Atoi(tailStr)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid tail parameter: %v", err), http.StatusBadRequest)
			return
		}
	}

	// Create log options
	logOptions := &mcp.LogOptions{
		Container: container,
		Since:     since,
		Tail:      tail,
		Pattern:   pattern,
		LogLevel:  logLevel,
		Format:    format,
	}

	var cmd *mcp.Command

	// Handle different log endpoints
	switch parts[3] {
	case "search":
		// Search logs
		namespace := r.URL.Query().Get("namespace")
		pod := r.URL.Query().Get("pod")
		if namespace == "" || pod == "" {
			http.Error(w, "Namespace and pod are required for log search", http.StatusBadRequest)
			return
		}

		logOptions.Pod = pod
		cmd = &mcp.Command{
			Type:       mcp.SearchLogsCommand,
			Namespace:  namespace,
			LogOptions: logOptions,
		}
	case "export":
		// Export logs
		namespace := r.URL.Query().Get("namespace")
		pod := r.URL.Query().Get("pod")
		if namespace == "" || pod == "" {
			http.Error(w, "Namespace and pod are required for log export", http.StatusBadRequest)
			return
		}

		if format == "" {
			http.Error(w, "Format is required for log export", http.StatusBadRequest)
			return
		}

		logOptions.Pod = pod
		cmd = &mcp.Command{
			Type:       mcp.ExportLogsCommand,
			Namespace:  namespace,
			LogOptions: logOptions,
		}
	default:
		// Get logs for a specific pod
		namespace := parts[3]
		if len(parts) < 5 {
			http.Error(w, "Pod name is required", http.StatusBadRequest)
			return
		}
		pod := parts[4]

		logOptions.Pod = pod
		cmd = &mcp.Command{
			Type:       mcp.LogsCommand,
			Namespace:  namespace,
			LogOptions: logOptions,
		}
	}

	// Handle command
	resp, err := s.mcpHandler.HandleCommand(cmd)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to handle command: %v", err), http.StatusInternalServerError)
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	if !resp.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(resp)
}

// handleHealthCheck handles health check requests
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
