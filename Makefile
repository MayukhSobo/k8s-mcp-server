.PHONY: build run docker-build docker-run deploy clean

# Variables
APP_NAME := k8s-mcp-server
DOCKER_IMAGE := $(APP_NAME):latest
KUBERNETES_NAMESPACE := default

# Build the application
build:
	go build -o $(APP_NAME) ./cmd/server

# Run the application
run: build
	./$(APP_NAME) serve

# Build Docker image
docker-build:
	docker build -t $(DOCKER_IMAGE) .

# Run Docker container
docker-run: docker-build
	docker run -p 8080:8080 $(DOCKER_IMAGE)

# Deploy to Kubernetes
deploy: docker-build
	kubectl apply -f deploy/kubernetes/deployment.yaml

# Clean up
clean:
	rm -f $(APP_NAME)
	docker rmi $(DOCKER_IMAGE) || true

# Download dependencies
deps:
	go mod download

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golint ./...

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  run          - Run the application locally"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  deploy       - Deploy to Kubernetes"
	@echo "  clean        - Clean up build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  test         - Run tests"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  help         - Show this help message" 