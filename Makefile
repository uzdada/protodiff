.PHONY: build run test docker-build deploy clean

# Variables
BINARY_NAME=protodiff
DOCKER_IMAGE=protodiff:latest
KUBECTL=kubectl

# Build the Go binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) ./cmd/protodiff

# Run the application locally
run:
	@echo "Running $(BINARY_NAME)..."
	@go run ./cmd/protodiff

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .

# Deploy to Kubernetes
deploy:
	@echo "Deploying to Kubernetes..."
	@$(KUBECTL) apply -f deploy/k8s/install.yaml

# Undeploy from Kubernetes
undeploy:
	@echo "Removing from Kubernetes..."
	@$(KUBECTL) delete -f deploy/k8s/install.yaml

# Port-forward to access dashboard
port-forward:
	@echo "Port-forwarding to dashboard..."
	@$(KUBECTL) port-forward -n protodiff-system svc/protodiff 8080:80

# View logs
logs:
	@echo "Viewing logs..."
	@$(KUBECTL) logs -n protodiff-system -l app.kubernetes.io/name=protodiff -f

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -rf bin/ dist/

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Build and push Docker image
docker-push: docker-build
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_IMAGE)

# Full deployment workflow
full-deploy: docker-build deploy
	@echo "Full deployment completed!"

# Help
help:
	@echo "ProtoDiff - Makefile Commands"
	@echo ""
	@echo "  make build          - Build the binary"
	@echo "  make run            - Run the application locally"
	@echo "  make test           - Run tests"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make deploy         - Deploy to Kubernetes"
	@echo "  make undeploy       - Remove from Kubernetes"
	@echo "  make port-forward   - Port-forward to dashboard"
	@echo "  make logs           - View application logs"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make deps           - Install dependencies"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Run linter"
	@echo "  make help           - Show this help message"
