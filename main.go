package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"model-test/models"
	"model-test/services"
)

func main() {
	// Command line flags
	var (
		apiKey     = flag.String("api-key", "DMR", "OpenAI API key (or set OPENAI_API_KEY env var)")
		baseURL    = flag.String("base-url", "http://localhost:12434/engines/v1", "OpenAI API base URL (or set OPENAI_BASE_URL env var)")
		model      = flag.String("model", "", "Model to use (or set OPENAI_MODEL env var, defaults to gpt-4o-mini)")
		configFile = flag.String("config", "config/test_cases.json", "Path to test cases configuration file")
		testCase   = flag.String("test-case", "", "Run only the specified test case by name")
	)
	flag.Parse()

	// Get API key from flag or environment
	if *apiKey == "" {
		*apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if *apiKey == "" {
		*apiKey = "DMR"
	}

	// Get base URL from flag or environment
	if *baseURL == "" {
		*baseURL = os.Getenv("OPENAI_BASE_URL")
	}
	if *baseURL == "" {
		*baseURL = "http://localhost:12434/engines/v1"
	}

	// Get model from flag or environment
	if *model == "" {
		*model = os.Getenv("OPENAI_MODEL")
	}
	// Default model will be set in the service if empty

	// Load test cases
	testCases, err := loadTestCases(*configFile, *testCase)
	if err != nil {
		log.Fatalf("Failed to load test cases: %v", err)
	}

	// Generate output filenames with model name
	sanitizedModel := sanitizeModelName(*model)
	timestamp := time.Now().Format("20060102_150405")
	outputFile := fmt.Sprintf("results/agent_test_results_%s_%s.json", sanitizedModel, timestamp)
	logFile := fmt.Sprintf("logs/agent_test_logs_%s_%s.log", sanitizedModel, timestamp)

	// Ensure directories exist
	if err := os.MkdirAll("results", 0755); err != nil {
		log.Fatalf("Failed to create results directory: %v", err)
	}
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	// Create request logger
	logger, err := services.NewRequestLogger(logFile)
	if err != nil {
		log.Fatalf("Failed to create request logger: %v", err)
	}
	defer logger.Close()

	// Create test runner with logger
	runner := services.NewTestRunnerWithLogger(*apiKey, *baseURL, *model, logger)

	// Print test configuration
	fmt.Printf("🚀 Starting Agent Loop Tool Efficiency Test\n")
	fmt.Printf("📊 Configuration:\n")
	fmt.Printf("   Base URL: %s\n", *baseURL)
	modelName := *model
	if modelName == "" {
		modelName = "gpt-4o-mini (default)"
	}
	fmt.Printf("   Model: %s\n", modelName)
	if *testCase != "" {
		fmt.Printf("   Single Test Case: %s\n", *testCase)
	}
	fmt.Printf("   Test Cases: %d\n", len(testCases))
	fmt.Printf("   Output: %s\n", outputFile)
	fmt.Printf("   Log File: %s\n", logFile)
	fmt.Println()

	// Run tests
	ctx := context.Background()

	fmt.Println("🔄 Running agent tests...")
	startTime := time.Now()

	report, err := runner.RunAgentTestSuite(ctx, testCases)
	if err != nil {
		log.Fatalf("Failed to run agent test suite: %v", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ Tests completed in %v\n\n", duration)

	// Save results
	if err := runner.SaveResults(outputFile, report); err != nil {
		log.Fatalf("Failed to save results: %v", err)
	}

	// Print summary
	printAgentSummary(report)

	fmt.Printf("\n💾 Results saved to: %s\n", outputFile)
	fmt.Printf("📝 Request logs saved to: %s\n", logFile)
}

// loadTestCases loads test cases from a JSON file, optionally filtering by test case name
func loadTestCases(filename string, testCaseName string) ([]models.TestCase, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read test cases file: %w", err)
	}

	var allTestCases []models.TestCase
	if err := json.Unmarshal(data, &allTestCases); err != nil {
		return nil, fmt.Errorf("failed to parse test cases: %w", err)
	}

	// If no specific test case is requested, return all test cases
	if testCaseName == "" {
		return allTestCases, nil
	}

	// Filter for the specific test case
	var filteredTestCases []models.TestCase
	for _, testCase := range allTestCases {
		if testCase.Name == testCaseName {
			filteredTestCases = append(filteredTestCases, testCase)
			break
		}
	}

	// Validate that the test case was found
	if len(filteredTestCases) == 0 {
		return nil, fmt.Errorf("test case '%s' not found in configuration file", testCaseName)
	}

	return filteredTestCases, nil
}

// printAgentSummary prints a summary of the agent test results
func printAgentSummary(report *models.AgentReport) {
	fmt.Println("📈 Agent Test Results")
	fmt.Println(strings.Repeat("=", 50))

	// Print overall statistics
	fmt.Printf("Total Tests: %d\n", report.TotalTests)
	fmt.Printf("✅ Passed: %d\n", report.PassedTests)
	fmt.Printf("❌ Failed: %d\n", report.FailedTests)
	fmt.Printf("⏱️  Total LLM Time: %v\n", report.TotalLLMTime)
	fmt.Printf("⏱️  Average Time per Request: %v\n", report.AvgTimePerReq)
	fmt.Println()

	// Print results by test case
	fmt.Println("📋 Test Case Results:")
	fmt.Println(strings.Repeat("-", 50))

	for _, result := range report.Results {
		status := "❌ FAILED"
		if result.Success {
			status = "✅ PASSED"
		}

		fmt.Printf("Test Case: %s\n", result.TestCase.Name)
		fmt.Printf("  Status: %s\n", status)
		if result.MatchedPath != "" {
			fmt.Printf("  Matched Path: %s\n", result.MatchedPath)
		}
		fmt.Printf("  Response Time: %v\n", result.ResponseTime)

		if result.Response != nil {
			fmt.Printf("  Tool Calls: %d\n", len(result.Response.ToolCalls))
			if len(result.Response.ToolCalls) > 0 {
				fmt.Printf("  Tools Used: ")
				for i, toolCall := range result.Response.ToolCalls {
					if i > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", toolCall.ToolName)
				}
				fmt.Println()
			}
		}

		if result.ErrorMessage != "" {
			fmt.Printf("  Error: %s\n", result.ErrorMessage)
		}

		fmt.Println(strings.Repeat("-", 30))
	}

	// Print failed tests details
	if report.FailedTests > 0 {
		fmt.Println("\n❌ Failed Tests Details:")
		fmt.Println(strings.Repeat("-", 50))
		for _, result := range report.Results {
			if !result.Success {
				fmt.Printf("Test Case: %s\n", result.TestCase.Name)
				fmt.Printf("Expected Tool Variants: %d\n", len(result.TestCase.ExpectedToolVariants))
				for i, variant := range result.TestCase.ExpectedToolVariants {
					fmt.Printf("  Variant %d (%s): %d tools\n", i+1, variant.Name, len(variant.Tools))
				}

				if result.Response != nil {
					fmt.Printf("Actual Tool Calls: %d\n", len(result.Response.ToolCalls))
					for i, toolCall := range result.Response.ToolCalls {
						fmt.Printf("  %d. %s\n", i+1, toolCall.ToolName)
					}
				}

				if result.ErrorMessage != "" {
					fmt.Printf("Error: %s\n", result.ErrorMessage)
				}
				fmt.Printf("Response Time: %v\n", result.ResponseTime)
				fmt.Println(strings.Repeat("-", 30))
			}
		}
	}

	// Print overall success rate
	successRate := float64(report.PassedTests) / float64(report.TotalTests) * 100
	fmt.Printf("\n📊 Overall Success Rate: %.2f%%\n", successRate)
}

// sanitizeModelName sanitizes the model name for use in filenames
func sanitizeModelName(modelName string) string {
	if modelName == "" {
		modelName = "gpt-4o-mini"
	}

	// Replace problematic characters with underscores
	sanitized := strings.ReplaceAll(modelName, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, " ", "_")
	sanitized = strings.ReplaceAll(sanitized, ":", "_")

	return sanitized
}
