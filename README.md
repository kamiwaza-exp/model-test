# OpenAI Model Tool Efficiency Test

A Go application for testing AI models with function calling using local AI servers.

## Quick Start

```bash
# Clone and setup
git clone https://github.com/ilopezluna/model-test
cd model-test
go mod download

# Run with default model
make run

# Run with multiple models
make run MODELS="ai/llama3.2,ai/gemma3"
```

## Make Commands

```bash
make run                                    # Run with default model (ai/llama3.2)
make run MODELS="ai/llama3.2"             # Run with single model
make run MODELS="ai/llama3.2,ai/gemma3"   # Run with multiple models
make help                                   # Show all available commands
make list-tests                            # List available test cases
```

## Default Configuration

- **API Key**: `DMR` (fake key for local models)
- **Base URL**: `http://localhost:12434/engines/v1`
- **Model**: `ai/llama3.2` (override with `MODELS=`)
- **Temperature**: `0.8`

## Advanced Usage

```bash
# Direct go run with custom options
go run main.go -models="gpt-4" -base-url="https://api.openai.com/v1" -api-key="real-key"

# Run specific test case
go run main.go -test-case="simple_add_iphone"
```

## Requirements

- Go 1.19+
- Local AI server (e.g., Ollama) on localhost:12434
