package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"model-test/models"
	"model-test/tools"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
)

// OpenAIService handles interactions with the OpenAI API
type OpenAIService struct {
	client        openai.Client
	shoppingTools *tools.ShoppingTools
}

// NewOpenAIService creates a new OpenAI service instance
func NewOpenAIService(apiKey, baseURL string) *OpenAIService {
	options := []option.RequestOption{
		option.WithBaseURL(baseURL),
		option.WithAPIKey(apiKey),
	}

	client := openai.NewClient(options...)

	return &OpenAIService{
		client:        client,
		shoppingTools: tools.NewShoppingTools(),
	}
}

// ExecuteTest runs a single test case and returns the result
func (s *OpenAIService) ExecuteTest(ctx context.Context, execution models.TestExecution) (*models.TestResult, error) {
	startTime := time.Now()

	// Prepare the chat completion request
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(execution.TestCase.Prompt),
	}

	// Build request parameters
	params := openai.ChatCompletionNewParams{
		Model:    execution.ModelName,
		Messages: messages,
		Tools:    s.shoppingTools.GetToolDefinitions(),
	}

	// Apply configuration
	if execution.Config.SystemPrompt != "" {
		//Append first message as system message
		params.Messages = append([]openai.ChatCompletionMessageParamUnion{openai.SystemMessage(execution.Config.SystemPrompt)}, params.Messages...)
	}
	if execution.Config.Temperature != nil {
		params.Temperature = param.NewOpt(float64(*execution.Config.Temperature))
	}
	if execution.Config.MaxTokens != nil {
		params.MaxTokens = param.NewOpt(int64(*execution.Config.MaxTokens))
	}

	// Create request object for logging
	request := &models.APIRequest{
		Method: "POST",
		URL:    "/v1/chat/completions",
		Body:   params,
	}

	// Make the API call
	response, err := s.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return &models.TestResult{
			TestExecution: execution,
			Success:       false,
			ErrorMessage:  err.Error(),
			Timestamp:     time.Now(),
			Request:       request,
			Response: &models.APIResponse{
				StatusCode: 0,
				Duration:   time.Since(startTime),
			},
			Metrics: models.TestMetrics{
				ResponseTime: time.Since(startTime),
			},
		}, nil
	}

	responseTime := time.Since(startTime)

	// Extract tool calls from response
	actualTools := s.extractToolCalls(response)

	// Calculate metrics
	metrics := s.calculateMetrics(execution.TestCase, actualTools, response, responseTime)

	// Check success and get matched path
	success, matchedPath := s.isTestSuccessfulWithPath(execution.TestCase, actualTools)

	// Create response object for logging
	apiResponse := &models.APIResponse{
		StatusCode: 200, // Assuming success if no error
		Body:       response,
		Duration:   responseTime,
	}

	return &models.TestResult{
		TestExecution: execution,
		Metrics:       metrics,
		ActualTools:   actualTools,
		Success:       success,
		MatchedPath:   matchedPath,
		Timestamp:     time.Now(),
		Request:       request,
		Response:      apiResponse,
	}, nil
}

// extractToolCalls extracts tool calls from the OpenAI response
func (s *OpenAIService) extractToolCalls(response *openai.ChatCompletion) []models.ActualToolCall {
	var actualTools []models.ActualToolCall

	if len(response.Choices) == 0 {
		return actualTools
	}

	choice := response.Choices[0]
	if choice.Message.ToolCalls == nil {
		return actualTools
	}

	for _, toolCall := range choice.Message.ToolCalls {
		if toolCall.Function.Name == "" {
			continue
		}

		// Parse function arguments
		var args map[string]interface{}
		if toolCall.Function.Arguments != "" {
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				// If parsing fails, store the raw string
				args = map[string]interface{}{
					"_raw_arguments": toolCall.Function.Arguments,
					"_parse_error":   err.Error(),
				}
			}
		}

		actualTools = append(actualTools, models.ActualToolCall{
			Name:      toolCall.Function.Name,
			Arguments: args,
		})
	}

	return actualTools
}

// calculateMetrics computes efficiency metrics for the test
func (s *OpenAIService) calculateMetrics(testCase models.TestCase, actual []models.ActualToolCall, response *openai.ChatCompletion, responseTime time.Duration) models.TestMetrics {
	// Get expected tools from the first variant as baseline
	var expected []models.ExpectedToolCall
	if len(testCase.ExpectedToolVariants) > 0 {
		expected = testCase.ExpectedToolVariants[0].Tools
	}
	metrics := models.TestMetrics{
		ResponseTime:       responseTime,
		TotalExpectedCalls: len(expected),
		TotalActualCalls:   len(actual),
	}

	// Extract token usage
	if response.Usage.PromptTokens > 0 {
		metrics.InputTokens = int(response.Usage.PromptTokens)
		metrics.OutputTokens = int(response.Usage.CompletionTokens)
		metrics.TotalTokens = int(response.Usage.TotalTokens)
	}

	// Calculate tool call accuracy
	correctCalls := 0
	for _, expectedTool := range expected {
		for _, actualTool := range actual {
			if s.isToolCallCorrect(expectedTool, actualTool) {
				correctCalls++
				break
			}
		}
	}

	metrics.CorrectToolCalls = correctCalls

	if len(expected) > 0 {
		metrics.ToolCallAccuracy = float64(correctCalls) / float64(len(expected))
	} else {
		// If no tools expected and none called, that's 100% accuracy
		if len(actual) == 0 {
			metrics.ToolCallAccuracy = 1.0
		} else {
			metrics.ToolCallAccuracy = 0.0
		}
	}

	// Calculate argument accuracy
	metrics.ArgumentAccuracy = s.calculateArgumentAccuracy(expected, actual)

	// Calculate completion rate (did we get the expected number of tool calls?)
	if len(expected) > 0 {
		metrics.CompletionRate = float64(min(len(actual), len(expected))) / float64(len(expected))
	} else {
		metrics.CompletionRate = 1.0
	}

	return metrics
}

// isToolCallCorrect checks if an actual tool call matches an expected one
func (s *OpenAIService) isToolCallCorrect(expected models.ExpectedToolCall, actual models.ActualToolCall) bool {
	if expected.Name != actual.Name {
		return false
	}

	// Check if all expected arguments are present and correct
	for key, expectedValue := range expected.Arguments {
		actualValue, exists := actual.Arguments[key]
		if !exists {
			return false
		}

		// Simple equality check (could be enhanced for more complex comparisons)
		if fmt.Sprintf("%v", expectedValue) != fmt.Sprintf("%v", actualValue) {
			return false
		}
	}

	return true
}

// calculateArgumentAccuracy calculates the accuracy of function arguments
func (s *OpenAIService) calculateArgumentAccuracy(expected []models.ExpectedToolCall, actual []models.ActualToolCall) float64 {
	if len(expected) == 0 {
		return 1.0
	}

	totalArguments := 0
	correctArguments := 0

	for _, expectedTool := range expected {
		// Find matching actual tool call
		var matchingActual *models.ActualToolCall
		for _, actualTool := range actual {
			if actualTool.Name == expectedTool.Name {
				matchingActual = &actualTool
				break
			}
		}

		if matchingActual == nil {
			// Tool not called at all, all arguments are wrong
			totalArguments += len(expectedTool.Arguments)
			continue
		}

		// Check each expected argument
		for key, expectedValue := range expectedTool.Arguments {
			totalArguments++
			if actualValue, exists := matchingActual.Arguments[key]; exists {
				if fmt.Sprintf("%v", expectedValue) == fmt.Sprintf("%v", actualValue) {
					correctArguments++
				}
			}
		}
	}

	if totalArguments == 0 {
		return 1.0
	}

	return float64(correctArguments) / float64(totalArguments)
}

// isTestSuccessful checks if the actual tool calls match any of the expected paths
func (s *OpenAIService) isTestSuccessful(testCase models.TestCase, actual []models.ActualToolCall) bool {
	success, _ := s.isTestSuccessfulWithPath(testCase, actual)
	return success
}

// isTestSuccessfulWithPath checks if the actual tool calls match any of the expected paths and returns which path matched
func (s *OpenAIService) isTestSuccessfulWithPath(testCase models.TestCase, actual []models.ActualToolCall) (bool, string) {
	// Check all variants to find a match
	if len(testCase.ExpectedToolVariants) > 0 {
		for _, variant := range testCase.ExpectedToolVariants {
			if s.isPathSuccessful(variant.Tools, actual) {
				return true, variant.Name
			}
		}
		return false, ""
	}

	// No expected tools defined - success if no actual tools called
	if len(actual) == 0 {
		return true, "no_tools_expected"
	}
	return false, ""
}

// isPathSuccessful checks if actual tool calls match a specific expected path
func (s *OpenAIService) isPathSuccessful(expected []models.ExpectedToolCall, actual []models.ActualToolCall) bool {
	// First check: exact count match
	if len(actual) != len(expected) {
		return false
	}

	// Second check: all expected tools must be called correctly in order
	for i, expectedTool := range expected {
		if i >= len(actual) || !s.isToolCallCorrect(expectedTool, actual[i]) {
			return false
		}
	}

	return true
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
