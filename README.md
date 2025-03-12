# Kubernetes MCP Server

A backend system that provides an interactive and extensible interface for managing Kubernetes resources, retrieving and analyzing logs, and formatting logs for export through the Model Context Protocol (MCP).

## Features

- CRUD operations on Kubernetes resources (Pods, Services, Namespaces, Deployments, etc.)
- Log retrieval and pattern searching
- Log formatting and exporting in multiple formats (Plaintext, JSON, CSV, NDJSON)
- Extensible architecture for future enhancements

## Requirements

- Go 1.24+
- Kubernetes cluster access
- kubectl configured

## Installation

```bash
# Clone the repository
git clone https://github.com/mayukhsarkar/k8s-mcp-server.git
cd k8s-mcp-server

# Build the binary
go build -o k8s-mcp-server

# Run the server
./k8s-mcp-server serve
```

## Usage

```bash
# Start the MCP server
./k8s-mcp-server serve

# Get help
./k8s-mcp-server --help
```

## API Documentation

The MCP server exposes HTTP endpoints for interacting with Kubernetes resources and logs.

### Kubernetes Operations

- `POST /api/v1/resources/{resource_type}` - Create a resource
- `GET /api/v1/resources/{resource_type}` - List resources
- `GET /api/v1/resources/{resource_type}/{name}` - Get resource details
- `DELETE /api/v1/resources/{resource_type}/{name}` - Delete a resource

### Log Operations

- `GET /api/v1/logs/{namespace}/{pod}` - Get logs from a pod
- `GET /api/v1/logs/search` - Search logs with pattern matching
- `GET /api/v1/logs/export` - Export logs in various formats

## License

MIT 