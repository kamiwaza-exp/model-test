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
	rm -rf logs/
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

# Test all available models
test-all-models:
	@echo "Running tests against all available models..."
	./test-all-models.sh

# Test specific models
test-models:
	@echo "Running tests against specified models: $(MODELS)"
	./test-all-models.sh -m "$(MODELS)"

# Test all models with specific test case
test-all-models-case:
	@echo "Running test case $(TEST_CASE) against all models"
	./test-all-models.sh -t "$(TEST_CASE)"

# Dry run to see what would be tested
test-dry-run:
	@echo "Dry run - showing what would be tested"
	./test-all-models.sh -n

# Build analysis tool
build-analyzer:
	@echo "Building batch analysis tool..."
	cd cmd/analyze-batch && go build -o ../../analyze-batch .
	@echo "Analysis tool built: analyze-batch"

# Analyze specific batch
analyze-batch: build-analyzer
	@if [ -z "$(BATCH_DIR)" ]; then \
		echo "Usage: make analyze-batch BATCH_DIR=results/batch_test_YYYYMMDD_HHMMSS"; \
		exit 1; \
	fi
	@echo "Analyzing batch: $(BATCH_DIR)"
	./analyze-batch "$(BATCH_DIR)"

# Analyze batch with JSON output
analyze-batch-json: build-analyzer
	@if [ -z "$(BATCH_DIR)" ]; then \
		echo "Usage: make analyze-batch-json BATCH_DIR=results/batch_test_YYYYMMDD_HHMMSS"; \
		exit 1; \
	fi
	@echo "Analyzing batch: $(BATCH_DIR) (JSON output)"
	./analyze-batch --format json "$(BATCH_DIR)"

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
	@echo "  test-all-models    - Test all available models (auto-discovered)"
	@echo "  test-models        - Test specific models (use MODELS=\"model1,model2\")"
	@echo "  test-all-models-case - Test specific case on all models (use TEST_CASE=)"
	@echo "  test-dry-run       - Dry run to see what would be tested"
	@echo "  build-analyzer     - Build the batch analysis tool"
	@echo "  analyze-batch      - Analyze specific batch (use BATCH_DIR=)"
	@echo "  analyze-batch-json - Analyze batch with JSON output (use BATCH_DIR=)"
	@echo "  help               - Show this help message"
	@echo ""
	@echo "🚀 USAGE EXAMPLES:"
	@echo "  make run                                    # Default model (gpt-4o-mini)"
	@echo "  make run MODEL=\"gpt-4\"                    # Specific model"
	@echo "  make run MODEL=\"ai/llama3.2\"             # Local model"
	@echo "  make run-test TEST_CASE=\"simple_view_cart\" # Single test case"
	@echo "  make run-model MODEL=\"claude-3-sonnet\"    # Custom model"
	@echo "  make list-tests                             # List available test cases"
	@echo "  make test-all-models                        # Test all discovered models"
	@echo "  make test-models MODELS=\"gpt-4,claude-3\"   # Test specific models"
	@echo "  make test-dry-run                           # See what would be tested"
	@echo "  make analyze-batch BATCH_DIR=results/batch_test_20240604_112030 # Analyze specific batch"
	@echo ""
	@echo "🔧 CONFIGURATION:"
	@echo "  MODEL              - Model to use (default: gpt-4o-mini)"
	@echo "  BASE_URL           - API base URL (default: http://localhost:13434)"
	@echo "  API_KEY            - API key (default: DMR)"
	@echo "  TEST_CASE          - Specific test case to run"
	@echo ""
	@echo "📁 OUTPUT:"
	@echo "  Results: results/agent_test_results_<model>_<timestamp>.json"
	@echo "  Logs:    logs/agent_test_logs_<model>_<timestamp>.log"
	@echo ""
	@echo "📊 FEATURES:"
	@echo "  • Agent loop with up to 5 LLM iterations"
	@echo "  • Tool calling efficiency testing"
	@echo "  • Cart state initialization for complex tests"
	@echo "  • LLM request timing metrics"
	@echo "  • Model name in result filenames"
	@echo "  • Structured JSON request/response logging"

# Phony targets
.PHONY: build clean run run-test run-model list-tests test-all-models test-models test-all-models-case test-dry-run build-analyzer analyze-batch analyze-batch-json help
