.PHONY: build run test clean docker-build docker-push help

# Variables
BINARY_NAME=secret-extractor
DOCKER_IMAGE?=ghcr.io/mjasion/supabase-secret-extractor
VERSION?=latest
GO=go
GOFLAGS=-v

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) ./cmd/job
	@echo "Build complete: ./$(BINARY_NAME)"

run: build ## Build and run with default settings
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME) \
		--namespace=default \
		--service-prefix=supabase \
		--public-url=http://localhost:8000 \
		--site-url=http://localhost:3000

test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v ./...

fmt: ## Format Go code
	@echo "Formatting code..."
	$(GO) fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...

lint: ## Run golangci-lint
	@echo "Running linter..."
	golangci-lint run

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

docker-build: ## Build Docker image
	@echo "Building Docker image: $(DOCKER_IMAGE):$(VERSION)"
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	@echo "Docker image built successfully"

docker-push: docker-build ## Push Docker image
	@echo "Pushing Docker image: $(DOCKER_IMAGE):$(VERSION)"
	docker push $(DOCKER_IMAGE):$(VERSION)
	@echo "Docker image pushed successfully"

docker-run: ## Run Docker container locally
	@echo "Running Docker container..."
	docker run --rm \
		-e KUBECONFIG=/root/.kube/config \
		-v ~/.kube:/root/.kube:ro \
		$(DOCKER_IMAGE):$(VERSION) \
		--namespace=default \
		--service-prefix=supabase

install-deps: ## Install Go dependencies
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

check: fmt vet ## Run format and vet checks

all: clean install-deps build ## Clean, install deps, and build
