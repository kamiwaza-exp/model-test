#!/bin/bash

# test-all-models.sh - Run tests against all available Docker models
# Usage: ./test-all-models.sh [options]

set -e  # Exit on any error

# Default configuration
#DEFAULT_BASE_URL="http://localhost:13434/engines/v1"
DEFAULT_BASE_URL="http://localhost:11434/v1"
DEFAULT_API_KEY="DMR"
DEFAULT_CONFIG="config/test_cases_simple.json"
#DEFAULT_DOCKER_CMD="docker model ls --openai | jq '.data[].id'"
DEFAULT_DOCKER_CMD="curl http://localhost:11434/v1/models | jq '.data[].id'"

# Script configuration
BASE_URL="${BASE_URL:-$DEFAULT_BASE_URL}"
API_KEY="${API_KEY:-$DEFAULT_API_KEY}"
CONFIG_FILE="${CONFIG_FILE:-$DEFAULT_CONFIG}"
TEST_RUNS="${TEST_RUNS:-10}"
TEST_CASE=""
MODELS_OVERRIDE=""
VERBOSE=false
DRY_RUN=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${PURPLE}$1${NC}"
}

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Run tests against all available Docker models or specified models.

OPTIONS:
    -h, --help              Show this help message
    -v, --verbose           Enable verbose output
    -n, --dry-run          Show what would be executed without running tests
    -m, --models MODELS     Comma-separated list of models to test (overrides auto-discovery)
    -r, --runs NUMBER       Number of test runs per model (default: 10)
    -t, --test-case NAME    Run only specific test case
    -c, --config FILE       Path to test cases config file (default: $DEFAULT_CONFIG)
    -u, --base-url URL      API base URL (default: $DEFAULT_BASE_URL)
    -k, --api-key KEY       API key (default: $DEFAULT_API_KEY)

ENVIRONMENT VARIABLES:
    BASE_URL               API base URL
    API_KEY                API key  
    CONFIG_FILE            Test cases config file
    TEST_RUNS              Number of test runs per model (default: 10)
    DOCKER_MODEL_CMD       Custom command to get models (default: $DEFAULT_DOCKER_CMD)

EXAMPLES:
    $0                                          # Test all discovered models (10 runs each)
    $0 -m "gpt-4,gpt-3.5-turbo"               # Test specific models (10 runs each)
    $0 -r 5                                    # Test all models with 5 runs each
    $0 -m "gpt-4" -r 3                         # Test gpt-4 with 3 runs
    $0 -t "simple_add_iphone"                  # Run single test case on all models
    $0 -v -n                                   # Dry run with verbose output
    $0 -u "http://localhost:8080/v1"           # Custom API endpoint
    TEST_RUNS=20 $0                            # Use environment variable for 20 runs

OUTPUT:
    Results are saved to: results/batch_test_YYYYMMDD_HHMMSS/
    - Individual model results: <model>_results.json
    - Execution log: test_execution.log  
    - Summary report: summary_report.json

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_usage
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -n|--dry-run)
            DRY_RUN=true
            shift
            ;;
        -m|--models)
            MODELS_OVERRIDE="$2"
            shift 2
            ;;
        -r|--runs)
            TEST_RUNS="$2"
            shift 2
            ;;
        -t|--test-case)
            TEST_CASE="$2"
            shift 2
            ;;
        -c|--config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        -u|--base-url)
            BASE_URL="$2"
            shift 2
            ;;
        -k|--api-key)
            API_KEY="$2"
            shift 2
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Function to log messages
log_message() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] $1" >> "$LOG_FILE"
    if [[ "$VERBOSE" == "true" ]]; then
        print_status "$1"
    fi
}

# Function to discover models
discover_models() {
    local docker_cmd="${DOCKER_MODEL_CMD:-$DEFAULT_DOCKER_CMD}"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_status "Discovering available models..." >&2
        log_message "Executing model discovery command: $docker_cmd"
        print_warning "DRY RUN: Would execute: $docker_cmd" >&2
        echo "gpt-4o-mini,gpt-4,claude-3-sonnet"  # Mock models for dry run
        return
    fi
    
    print_status "Discovering available models..." >&2
    log_message "Executing model discovery command: $docker_cmd"
    
    # Try to get models from docker
    local models=""
    if command -v docker >/dev/null 2>&1 && command -v jq >/dev/null 2>&1; then
        models=$(eval "$docker_cmd" 2>/dev/null | tr '\n' ',' | sed 's/,$//' | sed 's/"//g') || {
            print_warning "Failed to discover models using docker command" >&2
            log_message "Model discovery failed, will use fallback"
        }
    else
        print_warning "Docker or jq not available, skipping auto-discovery" >&2
        log_message "Docker or jq not found in PATH"
    fi
    
    # Fallback to common models if discovery failed
    if [[ -z "$models" ]]; then
        print_warning "Using fallback model list" >&2
        models="gpt-4o-mini,gpt-4,claude-3-sonnet"
        log_message "Using fallback models: $models"
    fi
    
    echo "$models"
}

# Function to sanitize model name for filename
sanitize_model_name() {
    echo "$1" | sed 's/[^a-zA-Z0-9._-]/_/g'
}

# Function to run test for a single model (single run)
run_model_test() {
    local model="$1"
    local run_number="$2"
    local sanitized_model=$(sanitize_model_name "$model")
    local model_log_file="$BATCH_DIR/${sanitized_model}_run${run_number}_test.log"
    
    print_status "Model $model - Run $run_number/$TEST_RUNS"
    log_message "Starting test run $run_number for model: $model"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_warning "DRY RUN: Would test model $model (run $run_number/$TEST_RUNS)"
        log_message "DRY RUN: Skipped test run $run_number for model $model"
        return 0
    fi
    
    # Build the application first
    print_status "Building application..."
    if ! make build >> "$model_log_file" 2>&1; then
        print_error "Failed to build application for model $model"
        log_message "Build failed for model: $model"
        return 1
    fi
    
    # Prepare test command
    local test_cmd="./model-test --model=\"$model\" --config=\"$CONFIG_FILE\" --base-url=\"$BASE_URL\" --api-key=\"$API_KEY\""
    
    if [[ -n "$TEST_CASE" ]]; then
        test_cmd="$test_cmd --test-case=\"$TEST_CASE\""
    fi
    
    log_message "Executing test command: $test_cmd"
    print_status "Running tests for $model..."
    
    # Run the test
    local start_time=$(date +%s)
    if eval "$test_cmd" >> "$model_log_file" 2>&1; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        print_success "Model $model completed in ${duration}s"
        log_message "Test completed successfully for model $model in ${duration}s"
        
        # Move result files to batch directory with model prefix
        local timestamp_pattern=$(date '+%Y%m%d')
        find results/ -name "*${timestamp_pattern}*" -newer "$BATCH_DIR" 2>/dev/null | while read -r file; do
            if [[ -f "$file" ]]; then
                local basename=$(basename "$file")
                local new_name="${sanitized_model}_${basename}"
                mv "$file" "$BATCH_DIR/$new_name"
                log_message "Moved result file: $file -> $BATCH_DIR/$new_name"
            fi
        done
        
        return 0
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        print_error "Model $model failed after ${duration}s"
        log_message "Test failed for model $model after ${duration}s"
        return 1
    fi
}

# Function to generate summary report
generate_summary() {
    local summary_file="$BATCH_DIR/summary_report.json"
    local total_models=0
    local successful_models=0
    local failed_models=0
    
    print_status "Generating summary report..."
    
    # Count results
    for model in $(echo "$TESTED_MODELS" | tr ',' '\n'); do
        total_models=$((total_models + 1))
        local sanitized_model=$(sanitize_model_name "$model")
        if ls "$BATCH_DIR/${sanitized_model}_agent_test_results_"* >/dev/null 2>&1; then
            successful_models=$((successful_models + 1))
        else
            failed_models=$((failed_models + 1))
        fi
    done
    
    # Create summary JSON
    cat > "$summary_file" << EOF
{
  "batch_test_summary": {
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "configuration": {
      "base_url": "$BASE_URL",
      "config_file": "$CONFIG_FILE",
      "test_case_filter": "$TEST_CASE"
    },
    "statistics": {
      "total_models": $total_models,
      "successful_models": $successful_models,
      "failed_models": $failed_models,
      "success_rate": $(echo "scale=2; $successful_models * 100 / $total_models" | bc -l 2>/dev/null || echo "0")
    },
    "models_tested": [
$(echo "$TESTED_MODELS" | tr ',' '\n' | sed 's/^/      "/' | sed 's/$/"/' | paste -sd ',' -)
    ],
    "results_directory": "$BATCH_DIR"
  }
}
EOF
    
    print_success "Summary report generated: $summary_file"
    log_message "Summary report created at $summary_file"
}

# Function to print final summary
print_final_summary() {
    local total_models=$(echo "$TESTED_MODELS" | tr ',' '\n' | wc -l)
    local successful_count=0
    local failed_count=0
    
    print_header "═══════════════════════════════════════════════════════════════"
    print_header "                    BATCH TEST SUMMARY"
    print_header "═══════════════════════════════════════════════════════════════"
    
    echo -e "${CYAN}Configuration:${NC}"
    echo "  Base URL: $BASE_URL"
    echo "  Config File: $CONFIG_FILE"
    echo "  Test Case Filter: ${TEST_CASE:-"All test cases"}"
    echo "  Results Directory: $BATCH_DIR"
    echo ""
    
    echo -e "${CYAN}Models Tested:${NC}"
    for model in $(echo "$TESTED_MODELS" | tr ',' '\n'); do
        local sanitized_model=$(sanitize_model_name "$model")
        if ls "$BATCH_DIR/${sanitized_model}_agent_test_results_"* >/dev/null 2>&1; then
            echo -e "  ${GREEN}✓${NC} $model"
            successful_count=$((successful_count + 1))
        else
            echo -e "  ${RED}✗${NC} $model"
            failed_count=$((failed_count + 1))
        fi
    done
    
    echo ""
    echo -e "${CYAN}Statistics:${NC}"
    echo "  Total Models: $total_models"
    echo -e "  ${GREEN}Successful: $successful_count${NC}"
    echo -e "  ${RED}Failed: $failed_count${NC}"
    
    if [[ $total_models -gt 0 ]]; then
        local success_rate=$(echo "scale=1; $successful_count * 100 / $total_models" | bc -l 2>/dev/null || echo "0")
        echo "  Success Rate: ${success_rate}%"
    fi
    
    print_header "═══════════════════════════════════════════════════════════════"
}

# Main execution
main() {
    # Create batch directory with timestamp
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    BATCH_DIR="results/batch_test_$timestamp"
    LOG_FILE="$BATCH_DIR/test_execution.log"
    
    # Create directories
    mkdir -p "$BATCH_DIR"
    mkdir -p results logs
    
    # Initialize log
    echo "Batch test execution started at $(date)" > "$LOG_FILE"
    
    print_header "🚀 Agent Loop Tool Efficiency - Batch Model Testing"
    print_header "Starting batch test at $(date)"
    
    # Discover or use override models
    local models=""
    if [[ -n "$MODELS_OVERRIDE" ]]; then
        models="$MODELS_OVERRIDE"
        print_status "Using manually specified models: $models"
        log_message "Using manual model override: $models"
    else
        models=$(discover_models)
        print_status "Discovered models: $models"
    fi
    
    if [[ -z "$models" ]]; then
        print_error "No models found to test"
        exit 1
    fi
    
    # Store tested models for summary
    TESTED_MODELS="$models"
    
    # Test each model multiple times
    local total_models=$(echo "$models" | tr ',' '\n' | wc -l)
    local current_model=0
    local successful_tests=0
    local total_runs=$((total_models * TEST_RUNS))
    local current_run=0
    
    print_status "Configuration: $TEST_RUNS runs per model, $total_models models = $total_runs total test runs"
    log_message "Starting batch test with $TEST_RUNS runs per model"
    
    for model in $(echo "$models" | tr ',' '\n'); do
        current_model=$((current_model + 1))
        print_header "Testing model $current_model/$total_models: $model"
        
        local model_successful_runs=0
        
        # Run the test suite multiple times for this model
        for run in $(seq 1 $TEST_RUNS); do
            current_run=$((current_run + 1))
            print_status "Overall progress: $current_run/$total_runs"
            
            if run_model_test "$model" "$run"; then
                model_successful_runs=$((model_successful_runs + 1))
            fi
            
            # Add small delay between runs to ensure unique timestamps
            if [[ "$DRY_RUN" != "true" && $run -lt $TEST_RUNS ]]; then
                sleep 2
            fi
        done
        
        # Report model completion
        print_success "Model $model completed: $model_successful_runs/$TEST_RUNS successful runs"
        log_message "Model $model completed with $model_successful_runs/$TEST_RUNS successful runs"
        
        if [[ $model_successful_runs -gt 0 ]]; then
            successful_tests=$((successful_tests + 1))
        fi
        
        echo ""  # Add spacing between models
    done
    
    # Generate summary
    generate_summary
    
    # Print final summary
    print_final_summary
    
    print_success "Batch testing completed!"
    print_status "Results saved to: $BATCH_DIR"
    
    log_message "Batch test execution completed at $(date)"
    log_message "Total models tested: $total_models, Successful: $successful_tests"
}

# Check dependencies
check_dependencies() {
    local missing_deps=()
    
    if ! command -v make >/dev/null 2>&1; then
        missing_deps+=("make")
    fi
    
    if ! command -v go >/dev/null 2>&1; then
        missing_deps+=("go")
    fi
    
    if [[ -z "$MODELS_OVERRIDE" ]] && ! command -v docker >/dev/null 2>&1; then
        print_warning "Docker not found - model auto-discovery will be limited"
    fi
    
    if [[ -z "$MODELS_OVERRIDE" ]] && ! command -v jq >/dev/null 2>&1; then
        print_warning "jq not found - model auto-discovery will be limited"
    fi
    
    if ! command -v bc >/dev/null 2>&1; then
        print_warning "bc not found - success rate calculation will be limited"
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        print_error "Missing required dependencies: ${missing_deps[*]}"
        exit 1
    fi
}

# Run dependency check and main function
check_dependencies
main "$@"
