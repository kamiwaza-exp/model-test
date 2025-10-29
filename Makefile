# Variables
BINARY_NAME=model-test
GO_FILES=$(shell find . -name "*.go" -type f)
MODEL ?= gpt-4o-mini
BASE_URL ?= http://localhost:13434
API_KEY ?= DMR
TEST_CASE ?=
MODELS ?= all
PROVIDER ?= default
KAMIWAZA_URL ?= https://localhost
KAMIWAZA_MODEL ?=

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

# Run the application with all parameters
run: build
	@echo "Running with provider: $(PROVIDER)"
	@if [ "$(PROVIDER)" = "kamiwaza" ]; then \
		echo "Kamiwaza model: $(KAMIWAZA_MODEL)"; \
		echo "Kamiwaza URL: $(KAMIWAZA_URL)"; \
		echo "Test case: $(TEST_CASE)"; \
		./$(BINARY_NAME) \
			--provider=kamiwaza \
			--kamiwaza-model="$(KAMIWAZA_MODEL)" \
			--kamiwaza-url="$(KAMIWAZA_URL)" \
			--api-key="$(API_KEY)" \
			--test-case="$(TEST_CASE)"; \
	else \
		echo "Model: $(MODEL)"; \
		echo "Base URL: $(BASE_URL)"; \
		echo "Test case: $(TEST_CASE)"; \
		./$(BINARY_NAME) \
			--model="$(MODEL)" \
			--base-url="$(BASE_URL)" \
			--api-key="$(API_KEY)" \
			--test-case="$(TEST_CASE)"; \
	fi

# Run tests against models
test: build
	@echo "Running tests..."
	@echo "Models: $(MODELS)"
	@echo "Test case: $(TEST_CASE)"
	@echo "Base URL: $(BASE_URL)"
	@echo "API Key: $(API_KEY)"
	./test-all-models.sh \
		--models="$(MODELS)" \
		--test-case="$(TEST_CASE)" \
		--base-url="$(BASE_URL)" \
		--api-key="$(API_KEY)"

# List available test cases
list-tests:
	@echo "Available test cases:"
	@jq -r '.[] | "  - " + .name + " (" + .prompt[0:50] + "...)"' config/test_cases.json 2>/dev/null || \
		awk '/"name":/ && !/"arguments"/ {gsub(/[,"]/, ""); print "  - " $2}' config/test_cases.json

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

# Analyze multiple batches as a combined analysis
analyze-multi-batch: build-analyzer
	@if [ -z "$(BATCH_DIRS)" ]; then \
		echo "Usage: make analyze-multi-batch BATCH_DIRS=\"batch1/ batch2/ batch3/\""; \
		exit 1; \
	fi
	@echo "Analyzing multiple batches: $(BATCH_DIRS)"
	./analyze-batch $(BATCH_DIRS)

# Analyze multiple batches with JSON output
analyze-multi-batch-json: build-analyzer
	@if [ -z "$(BATCH_DIRS)" ]; then \
		echo "Usage: make analyze-multi-batch-json BATCH_DIRS=\"batch1/ batch2/ batch3/\""; \
		exit 1; \
	fi
	@echo "Analyzing multiple batches: $(BATCH_DIRS) (JSON output)"
	./analyze-batch --format json $(BATCH_DIRS)

# Help target with comprehensive information
help:
	@echo "╔══════════════════════════════════════════════════════════════════════════════╗"
	@echo "║                    Agent Loop Tool Efficiency Test                          ║"
	@echo "╚══════════════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "🔨 MAKEFILE TARGETS:"
	@echo "  build              - Build the application binary"
	@echo "  clean              - Clean build artifacts and results"
	@echo "  run                - Run the application with all parameters"
	@echo "  test               - Run tests against models"
	@echo "  list-tests         - List all available test cases"
	@echo "  build-analyzer     - Build the batch analysis tool"
	@echo "  analyze-batch      - Analyze specific batch (use BATCH_DIR=)"
	@echo "  analyze-batch-json - Analyze batch with JSON output (use BATCH_DIR=)"
	@echo "  analyze-multi-batch - Analyze multiple batches (use BATCH_DIRS=)"
	@echo "  analyze-multi-batch-json - Analyze multiple batches with JSON (use BATCH_DIRS=)"
	@echo "  help               - Show this help message"
	@echo ""
	@echo "🚀 USAGE EXAMPLES:"
	@echo "  # Standard providers:"
	@echo "  make run                                    # Run with default values"
	@echo "  make run MODEL=\"gpt-4\"                    # Run with specific model"
	@echo "  make run TEST_CASE=\"simple_view_cart\"     # Run specific test case"
	@echo "  make run MODEL=\"gpt-4\" TEST_CASE=\"cart\" # Run with multiple parameters"
	@echo ""
	@echo "  # Kamiwaza provider:"
	@echo "  make run PROVIDER=kamiwaza KAMIWAZA_MODEL=\"GLM-4.5-Air-GGUF\""
	@echo "  make run PROVIDER=kamiwaza KAMIWAZA_MODEL=\"Qwen3-Coder-30B-A3B-Instruct-GGUF\" TEST_CASE=\"simple_view_cart\""
	@echo ""
	@echo "  # Batch testing:"
	@echo "  make test                                  # Test all models"
	@echo "  make test MODELS=\"gpt-4,claude-3\"         # Test specific models"
	@echo "  make test TEST_CASE=\"simple_view_cart\"    # Test specific case"
	@echo ""
	@echo "🔧 CONFIGURATION:"
	@echo "  PROVIDER           - Provider type: default, kamiwaza (default: default)"
	@echo "  MODEL              - Model to use for default provider (default: gpt-4o-mini)"
	@echo "  MODELS             - Models to test in batch mode (default: all)"
	@echo "  BASE_URL           - API base URL (default: http://localhost:13434)"
	@echo "  API_KEY            - API key (default: DMR)"
	@echo "  TEST_CASE          - Specific test case to run (default: all)"
	@echo "  KAMIWAZA_URL       - Kamiwaza base URL (default: https://localhost)"
	@echo "  KAMIWAZA_MODEL     - Kamiwaza model name (m_name from deployments)"
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
.PHONY: build clean run test list-tests build-analyzer analyze-batch analyze-batch-json analyze-multi-batch analyze-multi-batch-json help
