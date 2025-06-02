package models

import (
	"time"
)

// TestResult represents the result of a single test execution
type TestResult struct {
	TestExecution TestExecution    `json:"test_execution"`
	Metrics       TestMetrics      `json:"metrics"`
	ActualTools   []ActualToolCall `json:"actual_tools"`
	Success       bool             `json:"success"`
	ErrorMessage  string           `json:"error_message,omitempty"`
	Timestamp     time.Time        `json:"timestamp"`
	Request       *APIRequest      `json:"request,omitempty"`
	Response      *APIResponse     `json:"response,omitempty"`
}

// TestMetrics contains all efficiency metrics for a test
type TestMetrics struct {
	ResponseTime       time.Duration `json:"response_time_ms"`
	InputTokens        int           `json:"input_tokens"`
	OutputTokens       int           `json:"output_tokens"`
	TotalTokens        int           `json:"total_tokens"`
	ToolCallAccuracy   float64       `json:"tool_call_accuracy"` // 0-1
	ArgumentAccuracy   float64       `json:"argument_accuracy"`  // 0-1
	CompletionRate     float64       `json:"completion_rate"`    // 0-1
	CorrectToolCalls   int           `json:"correct_tool_calls"`
	TotalExpectedCalls int           `json:"total_expected_calls"`
	TotalActualCalls   int           `json:"total_actual_calls"`
}

// ModelConfigPair represents a model and configuration combination
type ModelConfigPair struct {
	ModelName string     `json:"model_name"`
	Config    TestConfig `json:"config"`
	Score     float64    `json:"score"`
}

// APIRequest represents the API request details
type APIRequest struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    interface{}       `json:"body,omitempty"`
}

// APIResponse represents the API response details
type APIResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       interface{}       `json:"body,omitempty"`
	Duration   time.Duration     `json:"duration_ms"`
}

// ComparisonReport contains comparison data between models/configurations
type ComparisonReport struct {
	Timestamp time.Time `json:"timestamp"`
	TestSuite string    `json:"test_suite"`
}

type Report struct {
	Timestamp   time.Time     `json:"timestamp"`
	TestSuite   string        `json:"test_suite"`
	Results     []TestResult  `json:"results"`
	TotalTests  int           `json:"total_tests"`
	PassedTests int           `json:"passed_tests"`
	FailedTests int           `json:"failed_tests"`
	AverageTime time.Duration `json:"average_time"`
}
