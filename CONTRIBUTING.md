# Contributing to Kubernetes MCP Server

Thank you for your interest in contributing to the Kubernetes MCP Server! This document provides guidelines and instructions for contributing.

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported by searching the [Issues](https://github.com/mayukhsarkar/k8s-mcp-server/issues).
2. If you don't find an open issue addressing the problem, [open a new one](https://github.com/mayukhsarkar/k8s-mcp-server/issues/new?template=bug_report.md).

### Suggesting Enhancements

1. Check if the enhancement has already been suggested by searching the [Issues](https://github.com/mayukhsarkar/k8s-mcp-server/issues).
2. If you don't find an open issue for your enhancement, [open a new one](https://github.com/mayukhsarkar/k8s-mcp-server/issues/new?template=feature_request.md).

### Pull Requests

1. Fork the repository.
2. Create a new branch for your feature or bugfix: `git checkout -b feature/your-feature-name` or `git checkout -b fix/your-bugfix-name`.
3. Make your changes.
4. Run tests to ensure your changes don't break existing functionality: `go test ./...`.
5. Commit your changes with a descriptive commit message.
6. Push your branch to your fork: `git push origin your-branch-name`.
7. [Submit a pull request](https://github.com/mayukhsarkar/k8s-mcp-server/compare) to the main branch.

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/mayukhsarkar/k8s-mcp-server.git
   cd k8s-mcp-server
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the application:
   ```bash
   go build -o k8s-mcp-server ./cmd/server
   ```

4. Run the application:
   ```bash
   ./k8s-mcp-server serve
   ```

## Coding Standards

- Follow Go's [official style guide](https://golang.org/doc/effective_go.html).
- Write tests for your code.
- Document your code with comments.
- Keep your code clean and maintainable.

## License

By contributing to this project, you agree that your contributions will be licensed under the project's [MIT License](LICENSE). 