# Variables
BINARY_NAME=model-test
GO_FILES=$(shell find . -name "*.go" -type f)
MODELS ?= ai/llama3.2

# Default target
.DEFAULT_GOAL := help

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	go clean
	rm -f $(BINARY_NAME)
	@echo "Clean complete"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod verify
	@echo "Dependencies installed"

# Run the application
run:
	OPENAI_API_KEY=DMR OPENAI_BASE_URL=http://localhost:12434/engines/v1 go run main.go -models="$(MODELS)"

# Format Go code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Formatting complete"

# Vet Go code
vet:
	@echo "Running go vet..."
	go vet ./...
	@echo "Vet complete"

# Run tests (if any exist)
test:
	@echo "Running tests..."
	go test ./...

# Show application help
app-help: build
	./$(BINARY_NAME) -help

# List available test cases
list-tests:
	@echo "Available test cases:"
	@jq -r '.[] | "  - " + .name' config/test_cases.json 2>/dev/null || \
		awk '/"name":/ && !/"arguments"/ {gsub(/[,"]/, ""); print "  - " $2}' config/test_cases.json

# Help target with comprehensive information
help:
	@echo "╔══════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                    OpenAI Model Tool Efficiency Test                        ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "🔨 MAKEFILE TARGETS:"
	@echo "  build              - Build the application binary"
	@echo "  clean              - Clean build artifacts"
	@echo "  deps               - Install and verify dependencies"
	@echo "  run                - Run the application (use MODELS= to specify models)"
	@echo "  fmt                - Format Go code"
	@echo "  vet                - Run go vet for code analysis"
	@echo "  test               - Run Go tests"
	@echo "  app-help           - Show application help (builds first if needed)"
	@echo "  list-tests         - List all available test cases"
	@echo "  help               - Show this help message"
	@echo ""
	@echo "🚀 USAGE EXAMPLES:"
	@echo "  make run                                    # Default model (ai/llama3.2)"
	@echo "  make run MODELS=\"ai/llama3.2\"             # Single model"
	@echo "  make run MODELS=\"ai/llama3.2,ai/gemma3\"   # Multiple models"
	@echo "  make list-tests                             # List available test cases"
	@echo ""
	@echo "⚙️  DEFAULT CONFIGURATION:"
	@echo "  API Key: DMR (fake key for local models)"
	@echo "  Base URL: http://localhost:12434/engines/v1"
	@echo "  Models: ai/llama3.2 (override with MODELS=)"
	@echo "  Temperature: 0.0"
	@echo ""
	@echo "📁 OUTPUT:"
	@echo "  Results are saved to the 'results/' directory with timestamped filenames"

# Phony targets
.PHONY: build clean deps run fmt vet test app-help list-tests help
