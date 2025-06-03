package models

import "github.com/openai/openai-go"

// TestCase represents a single test scenario
type TestCase struct {
	Name                 string             `json:"name"`
	Prompt               string             `json:"prompt"`
	ExpectedToolVariants []ExpectedToolPath `json:"expected_tools_variants"` // Multi-path format
}

// ExpectedToolPath represents one valid execution path
type ExpectedToolPath struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Tools       []ExpectedToolCall `json:"tools"`
}

// ExpectedToolCall represents the expected function call
type ExpectedToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// TestConfig holds configuration parameters for the test
type TestConfig struct {
	SystemPrompt string  `json:"system_prompt,omitempty"`
	Temperature  float32 `json:"temperature,omitempty"`
	TopK         int     `json:"top_k,omitempty"`
	MaxTokens    int     `json:"max_tokens,omitempty"`
}

// TestExecution represents a single test execution
type TestExecution struct {
	TestCase  TestCase   `json:"test_case"`
	ModelName string     `json:"model_name"`
	Config    TestConfig `json:"config"`
}

// ActualToolCall represents the actual function call made by the model
type ActualToolCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ChatCompletionRequest represents the request sent to OpenAI
type ChatCompletionRequest struct {
	Model       string                                   `json:"model"`
	Messages    []openai.ChatCompletionMessageParamUnion `json:"messages"`
	Tools       []openai.ChatCompletionToolParam         `json:"tools"`
	Temperature *float32                                 `json:"temperature,omitempty"`
	TopP        *float32                                 `json:"top_p,omitempty"`
	MaxTokens   *int                                     `json:"max_tokens,omitempty"`
}
