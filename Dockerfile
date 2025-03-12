FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o k8s-mcp-server ./cmd/server

# Create a minimal image
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/k8s-mcp-server .

# Expose the port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app/k8s-mcp-server", "serve"] 