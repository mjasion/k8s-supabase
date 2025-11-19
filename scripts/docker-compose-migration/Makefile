.PHONY: build run clean install test help

# Variables
BINARY_NAME=supabase-converter
OUTPUT_DIR=../../
GO=go
GOFLAGS=-v

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the converter binary
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

run: build ## Build and run the converter with default settings
	@echo "Running converter..."
	./$(BINARY_NAME)

run-local: build ## Run converter with local docker-compose files
	@echo "Running converter with local files..."
	./$(BINARY_NAME) -skip-fetch -local ./test-data

install: ## Install dependencies
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

clean: ## Clean build artifacts and generated files
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -rf $(OUTPUT_DIR)/kustomize
	rm -rf $(OUTPUT_DIR)/charts
	@echo "Clean complete"

test: ## Run tests
	@echo "Running tests..."
	$(GO) test -v ./...

fmt: ## Format Go code
	@echo "Formatting code..."
	$(GO) fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	$(GO) vet ./...

lint: ## Run golangci-lint (requires golangci-lint installed)
	@echo "Running linter..."
	golangci-lint run

check: fmt vet ## Run format and vet checks

all: clean install build ## Clean, install deps, and build

generate: build ## Generate Kubernetes manifests
	@echo "Generating manifests..."
	./$(BINARY_NAME)
	@echo "Generation complete!"
	@echo ""
	@echo "Generated files:"
	@echo "  - Kustomize: $(OUTPUT_DIR)/kustomize/"
	@echo "  - Helm Chart: $(OUTPUT_DIR)/charts/supabase/"

verify-kustomize: ## Verify kustomize manifests
	@echo "Verifying Kustomize manifests..."
	kubectl kustomize $(OUTPUT_DIR)/kustomize/base > /dev/null
	@echo "✓ Kustomize manifests are valid"

verify-helm: ## Verify Helm chart
	@echo "Verifying Helm chart..."
	helm lint $(OUTPUT_DIR)/charts/supabase
	@echo "✓ Helm chart is valid"

verify: verify-kustomize verify-helm ## Verify both Kustomize and Helm outputs

deploy-kustomize: ## Deploy using Kustomize
	@echo "Deploying with Kustomize..."
	kubectl apply -k $(OUTPUT_DIR)/kustomize/base

deploy-helm: ## Deploy using Helm
	@echo "Deploying with Helm..."
	helm install supabase $(OUTPUT_DIR)/charts/supabase -n supabase --create-namespace

undeploy-kustomize: ## Remove Kustomize deployment
	@echo "Removing Kustomize deployment..."
	kubectl delete -k $(OUTPUT_DIR)/kustomize/base

undeploy-helm: ## Remove Helm deployment
	@echo "Removing Helm deployment..."
	helm uninstall supabase -n supabase
