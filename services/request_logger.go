package services

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/openai/openai-go"
)

// RequestLogger handles logging of HTTP requests and responses
type RequestLogger struct {
	logFile *os.File
}

// LogEntry represents a single request/response log entry
type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	TestCase  string      `json:"test_case"`
	Iteration int         `json:"iteration"`
	Request   LogRequest  `json:"request"`
	Response  LogResponse `json:"response"`
	Error     string      `json:"error,omitempty"`
}

// LogRequest represents the request part of a log entry
type LogRequest struct {
	Method string      `json:"method"`
	URL    string      `json:"url"`
	Body   interface{} `json:"body"`
}

// LogResponse represents the response part of a log entry
type LogResponse struct {
	StatusCode int         `json:"status_code"`
	Body       interface{} `json:"body"`
}

// NewRequestLogger creates a new request logger with the specified log file
func NewRequestLogger(logFilePath string) (*RequestLogger, error) {
	// Ensure logs directory exists
	if err := os.MkdirAll("logs", 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Create or open the log file
	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}

	return &RequestLogger{
		logFile: logFile,
	}, nil
}

// LogRequest logs a successful request/response pair
func (rl *RequestLogger) LogRequest(testCase string, iteration int, requestParams openai.ChatCompletionNewParams, response *openai.ChatCompletion, baseURL string) error {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		TestCase:  testCase,
		Iteration: iteration,
		Request: LogRequest{
			Method: "POST",
			URL:    fmt.Sprintf("%s/chat/completions", baseURL),
			Body:   requestParams,
		},
		Response: LogResponse{
			StatusCode: 200,
			Body:       response,
		},
	}

	return rl.writeLogEntry(entry)
}

// LogError logs a failed request
func (rl *RequestLogger) LogError(testCase string, iteration int, requestParams openai.ChatCompletionNewParams, err error, baseURL string) error {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		TestCase:  testCase,
		Iteration: iteration,
		Request: LogRequest{
			Method: "POST",
			URL:    fmt.Sprintf("%s/chat/completions", baseURL),
			Body:   requestParams,
		},
		Response: LogResponse{
			StatusCode: 0, // Unknown status code for errors
			Body:       nil,
		},
		Error: err.Error(),
	}

	return rl.writeLogEntry(entry)
}

// writeLogEntry writes a log entry to the file
func (rl *RequestLogger) writeLogEntry(entry LogEntry) error {
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	// Write JSON entry followed by newline
	if _, err := rl.logFile.Write(jsonData); err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	if _, err := rl.logFile.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write newline: %w", err)
	}

	// Flush to ensure data is written immediately
	return rl.logFile.Sync()
}

// Close closes the log file
func (rl *RequestLogger) Close() error {
	if rl.logFile != nil {
		return rl.logFile.Close()
	}
	return nil
}
