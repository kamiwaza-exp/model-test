package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"model-test/models"
)

// TestRunner orchestrates the execution of test suites using the agent loop
type TestRunner struct {
	openaiService *OpenAIService
	results       []models.AgentTestResult
	mutex         sync.Mutex
	defaultModel  string
}

// NewTestRunner creates a new test runner instance
func NewTestRunner(apiKey, baseURL, defaultModel string) *TestRunner {
	return &TestRunner{
		openaiService: NewOpenAIService(apiKey, baseURL, defaultModel),
		results:       make([]models.AgentTestResult, 0),
		defaultModel:  defaultModel,
	}
}

// RunAgentTestSuite executes a test suite using the agent loop approach
func (tr *TestRunner) RunAgentTestSuite(ctx context.Context, testCases []models.TestCase) (*models.AgentReport, error) {
	fmt.Printf("Starting agent test suite with %d test cases\n", len(testCases))

	var wg sync.WaitGroup
	resultsChan := make(chan models.AgentTestResult, len(testCases))

	// Execute tests concurrently
	for _, testCase := range testCases {
		wg.Add(1)
		go func(tc models.TestCase) {
			defer wg.Done()

			fmt.Printf("Running agent test: %s\n", tc.Name)
			result := tr.runAgentTest(ctx, tc)
			resultsChan <- result
		}(testCase)
	}

	// Wait for all tests to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results and aggregate LLM metrics
	var results []models.AgentTestResult
	var totalTime time.Duration
	var totalLLMRequests int
	var totalLLMTime time.Duration
	passedTests := 0
	failedTests := 0

	for result := range resultsChan {
		results = append(results, result)
		totalTime += result.ResponseTime

		// Aggregate LLM metrics from successful responses
		if result.Response != nil {
			totalLLMRequests += result.Response.LLMRequests
			totalLLMTime += result.Response.LLMTotalTime
		}

		if result.Success {
			passedTests++
		} else {
			failedTests++
		}
	}

	// Calculate average times
	var averageTime time.Duration
	var avgTimePerReq time.Duration
	if len(results) > 0 {
		averageTime = totalTime / time.Duration(len(results))
	}
	if totalLLMRequests > 0 {
		avgTimePerReq = totalLLMTime / time.Duration(totalLLMRequests)
	}

	report := &models.AgentReport{
		Timestamp:        time.Now(),
		TestSuite:        "Agent Loop Tool Efficiency Test",
		Results:          results,
		TotalTests:       len(results),
		PassedTests:      passedTests,
		FailedTests:      failedTests,
		AverageTime:      averageTime,
		TotalLLMRequests: totalLLMRequests,
		TotalLLMTime:     totalLLMTime,
		AvgTimePerReq:    avgTimePerReq,
	}

	return report, nil
}

// runAgentTest executes a single test case using the agent loop
func (tr *TestRunner) runAgentTest(ctx context.Context, testCase models.TestCase) models.AgentTestResult {
	startTime := time.Now()

	// Generate a unique session ID for this test
	sessionID := fmt.Sprintf("test_%s_%d", testCase.Name, time.Now().UnixNano())

	// Create a session for the test
	session := &models.ChatSession{
		SessionID: sessionID,
		Messages:  []models.ChatMessage{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Initialize cart state if specified in the test case
	if testCase.InitialCartState != nil {
		err := tr.openaiService.InitializeCartForTest(sessionID, testCase.InitialCartState)
		if err != nil {
			return models.AgentTestResult{
				TestCase:     testCase,
				ModelName:    tr.getModelName(),
				Success:      false,
				ErrorMessage: fmt.Sprintf("Failed to initialize cart state: %v", err),
				Timestamp:    time.Now(),
				ResponseTime: time.Since(startTime),
			}
		}
	}

	// Execute the test using the agent loop
	response, err := tr.openaiService.ProcessChatMessage(ctx, testCase.Prompt, session)
	responseTime := time.Since(startTime)

	if err != nil {
		return models.AgentTestResult{
			TestCase:     testCase,
			ModelName:    tr.getModelName(),
			Success:      false,
			ErrorMessage: err.Error(),
			Timestamp:    time.Now(),
			ResponseTime: responseTime,
		}
	}

	// Evaluate if the test was successful by checking tool calls
	success, matchedPath := tr.evaluateAgentResponse(testCase, response)

	return models.AgentTestResult{
		TestCase:     testCase,
		ModelName:    tr.getModelName(),
		Response:     response,
		Success:      success,
		MatchedPath:  matchedPath,
		Timestamp:    time.Now(),
		ResponseTime: responseTime,
	}
}

// evaluateAgentResponse checks if the agent response matches expected tool calls
func (tr *TestRunner) evaluateAgentResponse(testCase models.TestCase, response *models.ChatResponse) (bool, string) {
	if len(testCase.ExpectedToolVariants) == 0 {
		// No expected tools - success if no tools were called
		return len(response.ToolCalls) == 0, "no_tools_expected"
	}

	// Extract actual tool calls from response
	actualTools := make([]models.ActualToolCall, len(response.ToolCalls))
	for i, toolResult := range response.ToolCalls {
		actualTools[i] = models.ActualToolCall{
			Name:      toolResult.ToolName,
			Arguments: tr.parseArguments(toolResult.Arguments),
		}
	}

	// Check all variants to find a match
	for _, variant := range testCase.ExpectedToolVariants {
		if tr.isPathSuccessful(variant.Tools, actualTools) {
			return true, variant.Name
		}
	}

	return false, ""
}

// parseArguments parses the arguments string into a map
func (tr *TestRunner) parseArguments(arguments string) map[string]interface{} {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		// If parsing fails, return the raw string
		return map[string]interface{}{
			"_raw_arguments": arguments,
			"_parse_error":   err.Error(),
		}
	}
	return args
}

// isPathSuccessful checks if actual tool calls match a specific expected path
func (tr *TestRunner) isPathSuccessful(expected []models.ExpectedToolCall, actual []models.ActualToolCall) bool {
	// First check: exact count match
	if len(actual) != len(expected) {
		return false
	}

	// Second check: all expected tools must be called correctly in order
	for i, expectedTool := range expected {
		if i >= len(actual) || !tr.isToolCallCorrect(expectedTool, actual[i]) {
			return false
		}
	}

	return true
}

// isToolCallCorrect checks if an actual tool call matches an expected one
func (tr *TestRunner) isToolCallCorrect(expected models.ExpectedToolCall, actual models.ActualToolCall) bool {
	if expected.Name != actual.Name {
		return false
	}

	// Check if all expected arguments are present and correct
	for key, expectedValue := range expected.Arguments {
		actualValue, exists := actual.Arguments[key]
		if !exists {
			return false
		}

		// Simple equality check using case-insensitive comparison
		if !strings.EqualFold(fmt.Sprintf("%v", expectedValue), fmt.Sprintf("%v", actualValue)) {
			return false
		}
	}

	return true
}

// getModelName returns the model name to use for test results
func (tr *TestRunner) getModelName() string {
	if tr.defaultModel == "" {
		return "gpt-4o-mini"
	}
	return tr.defaultModel
}

// SaveResults saves test results to a JSON file
func (tr *TestRunner) SaveResults(filename string, report *models.AgentReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}
