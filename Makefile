# Variables
BINARY_NAME=model-test
GO_FILES=$(shell find . -name "*.go" -type f)
MODEL ?= gpt-4o-mini
BASE_URL ?= http://localhost:13434
API_KEY ?= DMR

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
	rm -rf results/
	@echo "Clean complete"

# Run the application with default settings
run: build
	@echo "Running with model: $(MODEL)"
	OPENAI_API_KEY=$(API_KEY) OPENAI_BASE_URL=$(BASE_URL) ./$(BINARY_NAME) --model="$(MODEL)"

# Run a specific test case
run-test: build
	@echo "Running test case: $(TEST_CASE) with model: $(MODEL)"
	OPENAI_API_KEY=$(API_KEY) OPENAI_BASE_URL=$(BASE_URL) ./$(BINARY_NAME) --model="$(MODEL)" --test-case="$(TEST_CASE)"

# Run with custom model
run-model: build
	@echo "Running with custom model: $(MODEL)"
	./$(BINARY_NAME) --model="$(MODEL)" --base-url="$(BASE_URL)" --api-key="$(API_KEY)"

# List available test cases
list-tests:
	@echo "Available test cases:"
	@jq -r '.[] | "  - " + .name + " (" + .prompt[0:50] + "...)"' config/test_cases.json 2>/dev/null || \
		awk '/"name":/ && !/"arguments"/ {gsub(/[,"]/, ""); print "  - " $2}' config/test_cases.json

# Show recent results
show-results:
	@echo "Recent test results:"
	@ls -la results/ | tail -5 || echo "No results found. Run 'make run' first."

# Help target with comprehensive information
help:
	@echo "╔══════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                    Agent Loop Tool Efficiency Test                          ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "🔨 MAKEFILE TARGETS:"
	@echo "  build              - Build the application binary"
	@echo "  clean              - Clean build artifacts and results"
	@echo "  run                - Run all test cases (use MODEL= to specify model)"
	@echo "  run-test           - Run specific test case (use TEST_CASE= and MODEL=)"
	@echo "  run-model          - Run with custom model and settings"
	@echo "  list-tests         - List all available test cases"
	@echo "  show-results       - Show recent test results"
	@echo "  help               - Show this help message"
	@echo ""
	@echo "🚀 USAGE EXAMPLES:"
	@echo "  make run                                    # Default model (gpt-4o-mini)"
	@echo "  make run MODEL=\"gpt-4\"                    # Specific model"
	@echo "  make run MODEL=\"ai/llama3.2\"             # Local model"
	@echo "  make run-test TEST_CASE=\"simple_view_cart\" # Single test case"
	@echo "  make run-model MODEL=\"claude-3-sonnet\"    # Custom model"
	@echo "  make list-tests                             # List available test cases"
	@echo "  make show-results                           # View recent results"
	@echo ""
	@echo "🔧 CONFIGURATION:"
	@echo "  MODEL              - Model to use (default: gpt-4o-mini)"
	@echo "  BASE_URL           - API base URL (default: http://localhost:13434)"
	@echo "  API_KEY            - API key (default: DMR)"
	@echo "  TEST_CASE          - Specific test case to run"
	@echo ""
	@echo "📁 OUTPUT:"
	@echo "  Results are saved to 'results/' with format:"
	@echo "  agent_test_results_<model>_<timestamp>.json"
	@echo ""
	@echo "📊 FEATURES:"
	@echo "  • Agent loop with up to 5 LLM iterations"
	@echo "  • Tool calling efficiency testing"
	@echo "  • Cart state initialization for complex tests"
	@echo "  • LLM request timing metrics"
	@echo "  • Model name in result filenames"

# Phony targets
.PHONY: build clean run run-test run-model list-tests show-results help
