# Agent Loop Tool Efficiency Test

A Go application for testing AI models with function calling using an agent loop architecture. Tests tool calling
efficiency, cart management scenarios, and provides detailed performance metrics.

## Quick Start

```bash
# Clone and setup
git clone https://github.com/ilopezluna/model-test
cd model-test

# Run with default model
make run

# Run with specific model
make run MODEL="ai/llama3.2"

# Run single test case
make run TEST_CASE="simple_view_cart" MODEL="ai/gemma3"
```

## Command Line Usage

### Basic Usage

```bash
# Run all test cases with default model (gpt-4o-mini)
./model-test

# Run with specific model
./model-test --model "ai/qwen2.5"

# Run single test case
./model-test --test-case "simple_view_cart"

# Custom API settings
./model-test --model "gpt-4" --base-url "https://api.openai.com/v1" --api-key "your-key"
```

### Command Line Flags

```
  -api-key string
        OpenAI API key (or set OPENAI_API_KEY env var) (default "DMR")
  -base-url string
        OpenAI API base URL (or set OPENAI_BASE_URL env var) (default "http://localhost:13434")
  -config string
        Path to test cases configuration file (default "config/test_cases.json")
  -model string
        Model to use (or set OPENAI_MODEL env var, defaults to gpt-4o-mini)
  -test-case string
        Run only the specified test case by name
```

### Environment Variables

```bash
export OPENAI_API_KEY="your-api-key"
export OPENAI_BASE_URL="https://api.openai.com/v1"
export OPENAI_MODEL="gpt-4"
```

## Make Commands

### Basic Commands

```bash
# Run commands
make run                                    # Run with default values
make run MODEL="gpt-4"                     # Run with specific model
make run TEST_CASE="simple_view_cart"      # Run specific test case
make run MODEL="gpt-4" TEST_CASE="cart"    # Run with multiple parameters

# Test commands
make test                                  # Test all models
make test MODELS="gpt-4,claude-3"          # Test specific models
make test TEST_CASE="simple_view_cart"     # Test specific case
make test MODELS="gpt-4" TEST_CASE="cart"  # Test specific model and case

# Utility commands
make list-tests                            # List available test cases
make help                                  # Show all available commands
```

### Development Commands

```bash
make build          # Build the application
make clean          # Clean build artifacts and results
```

## Test Cases

The application includes 18 test cases covering:

- **Zero Tool Cases**: Greetings, general questions (no tools expected)
- **Simple Cases**: Single tool operations (search, add, view, remove, checkout)
- **Medium Cases**: Two-step operations (search then add, remove then add)
- **Complex Cases**: Multi-step workflows with cart management

### Example Test Cases

- `zero_greeting` - Simple greeting (no tools)
- `simple_search_electronics` - Search for electronics
- `simple_add_iphone` - Add iPhone to cart
- `medium_search_and_add` - Search and add to cart
- `complex_cart_management` - Multi-step cart organization (with initial cart state)

## Output and Results

### Result Files

Results are saved to `results/` directory with format:

```
agent_test_results_<model>_<timestamp>.json
```

Examples:

- `agent_test_results_gpt-4_20250603_112616.json`
- `agent_test_results_ai_llama3.2_20250603_112623.json`
- `agent_test_results_gpt-4o-mini_20250603_112630.json`

### Performance Metrics

```
📈 Agent Test Results
==================================================
Total Tests: 18
✅ Passed: 15
❌ Failed: 3
⏱️  Total LLM Time: 12.4s
⏱️  Average Time per Request: 1.2s
📊 Overall Success Rate: 83.33%
```

### Key Metrics

- **Total LLM Time**: Time spent in actual LLM requests (excludes framework overhead)
- **Average Time per Request**: Per individual LLM API call (not per test)
- **Tool Call Accuracy**: Matches expected tool calling patterns
- **Success Rate**: Percentage of tests that matched expected behavior

## Configuration

### Test Case Structure

```json
{
  "name": "complex_cart_management",
  "prompt": "Help me organize my shopping cart...",
  "initial_cart_state": {
    "items": [
      {
        "product_name": "iPhone",
        "quantity": 2
      },
      {
        "product_name": "Wireless Headphones",
        "quantity": 1
      }
    ]
  },
  "expected_tools_variants": [
    
  ]
}
```

### Available Tools

- `search_products` - Search by query, category, or both
- `add_to_cart` - Add products with quantity
- `remove_from_cart` - Remove products from cart
- `view_cart` - View cart contents and totals
- `checkout` - Process checkout

## Requirements

- **Go**: 1.19+
- **Local AI Server**: Docker Model Runner or Ollama
- **OR OpenAI API**: With valid API key

### Adding New Test Cases

1. Add test case to `config/test_cases.json`
2. Define expected tool call variants
3. Optionally specify initial cart state
4. Run with `make run TEST_CASE="your_test_name"`

### Model Comparison

```bash
# Test multiple models
make test MODELS="gpt-4,gpt-4o-mini,ai/llama3.2"

# Or test them individually
make run MODEL="gpt-4"
make run MODEL="gpt-4o-mini"
make run MODEL="ai/llama3.2"
```