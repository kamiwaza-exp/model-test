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
		apiKey     = flag.String("api-key", "", "OpenAI API key (or set OPENAI_API_KEY env var)")
		baseURL    = flag.String("base-url", "", "OpenAI API base URL (or set OPENAI_BASE_URL env var)")
		modelsFlag = flag.String("models", "ai/llama3.2", "Comma-separated list of models to test")
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

	// Parse models
	modelNames := strings.Split(*modelsFlag, ",")
	for i, model := range modelNames {
		modelNames[i] = strings.TrimSpace(model)
	}

	// Create configurations
	var configs []models.TestConfig
	config := models.TestConfig{
		Temperature:  0.0,
		TopK:         0,
		MaxTokens:    1000,
		SystemPrompt: "You are a helpful shopping assistant for chat2cart, an AI-powered shopping platform. Your role is to help users discover products, manage their shopping cart, and complete purchases through natural conversation.\n\nKey capabilities:\n- Search for products by name, description, or category\n- Add products to the shopping cart\n- Remove products from the cart\n- Update quantities in the cart\n- Show cart contents and totals\n- Process checkout\n\nGuidelines:\n- Be friendly, helpful, and conversational\n- Ask clarifying questions when needed (e.g., quantity, specific product details)\n- Provide product recommendations when appropriate\n- Always confirm actions like adding/removing items\n- Help users understand their cart contents and totals\n- Guide users through the checkout process\n\nAvailable product categories: electronics, clothing, books, home, sports\n\nWhen users ask about products, use the search_products tool to find relevant items. When they want to add items to their cart, use the appropriate cart management tools.",
	}
	configs = append(configs, config)

	// Load test cases
	testCases, err := loadTestCases(*configFile, *testCase)
	if err != nil {
		log.Fatalf("Failed to load test cases: %v", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	outputFile := fmt.Sprintf("results/test_results_%s.json", timestamp)

	// Ensure results directory exists
	if err := os.MkdirAll("results", 0755); err != nil {
		log.Fatalf("Failed to create results directory: %v", err)
	}

	// Create test runner
	runner := services.NewTestRunner(*apiKey, *baseURL)

	// Print test configuration
	fmt.Printf("🚀 Starting OpenAI Model Tool Efficiency Test\n")
	fmt.Printf("📊 Configuration:\n")
	fmt.Printf("   Models: %v\n", modelNames)

	if *baseURL != "" {
		fmt.Printf("   Base URL: %s\n", *baseURL)
	} else {
		fmt.Printf("   Base URL: https://api.openai.com/v1 (default)\n")
	}
	if *testCase != "" {
		fmt.Printf("   Single Test Case: %s\n", *testCase)
	}
	fmt.Printf("   Test Cases: %d\n", len(testCases))
	fmt.Printf("   Total Tests: %d\n", len(testCases)*len(modelNames)*len(configs))
	fmt.Printf("   Output: %s\n", outputFile)
	fmt.Printf("   System Prompt: %s\n", config.SystemPrompt)
	fmt.Printf("   Temperature: %v\n", config.Temperature)
	fmt.Printf("   TopK: %v\n", config.TopK)
	fmt.Printf("   Max Tokens: %v\n", config.MaxTokens)
	fmt.Println()

	// Run tests
	ctx := context.Background()

	fmt.Println("🔄 Running tests...")
	startTime := time.Now()

	report, err := runner.RunTestSuite(ctx, testCases, modelNames, configs)
	if err != nil {
		log.Fatalf("Failed to run test suite: %v", err)
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ Tests completed in %v\n\n", duration)

	// Save results
	if err := runner.SaveResults(outputFile, report); err != nil {
		log.Fatalf("Failed to save results: %v", err)
	}

	// Print summary
	printSummary(report)

	fmt.Printf("\n💾 Results saved to: %s\n", outputFile)
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

// printSummary prints a summary of the test results
func printSummary(report *models.Report) {
	fmt.Println("📈 Test Results")
	fmt.Println(strings.Repeat("=", 50))

	// Print overall statistics
	fmt.Printf("Total Tests: %d\n", report.TotalTests)
	fmt.Printf("✅ Passed: %d\n", report.PassedTests)
	fmt.Printf("❌ Failed: %d\n", report.FailedTests)
	fmt.Printf("⏱️  Average Response Time: %v\n", report.AverageTime)
	fmt.Println()

	// Group results by test case
	testCaseResults := make(map[string][]models.TestResult)
	for _, result := range report.Results {
		testCaseResults[result.TestExecution.TestCase.Name] = append(testCaseResults[result.TestExecution.TestCase.Name], result)
	}

	// Print results by test case
	fmt.Println("📋 Test Case Results:")
	fmt.Println(strings.Repeat("-", 50))
	for testCaseName, results := range testCaseResults {
		passed := 0
		failed := 0
		var totalTime time.Duration

		for _, result := range results {
			if result.Success {
				passed++
			} else {
				failed++
			}
			totalTime += result.Metrics.ResponseTime
		}

		avgTime := totalTime / time.Duration(len(results))
		successRate := float64(passed) / float64(len(results)) * 100

		fmt.Printf("Test Case: %s\n", testCaseName)
		fmt.Printf("  Total Runs: %d\n", len(results))
		fmt.Printf("  ✅ Passed: %d\n", passed)
		fmt.Printf("  ❌ Failed: %d\n", failed)
		fmt.Printf("  ⏱️  Average Time: %v\n", avgTime)
		fmt.Printf("  📊 Success Rate: %.2f%%\n", successRate)
		fmt.Println(strings.Repeat("-", 30))
	}

	// Print failed tests details
	if report.FailedTests > 0 {
		fmt.Println("\n❌ Failed Tests Details:")
		fmt.Println(strings.Repeat("-", 50))
		for _, result := range report.Results {
			if !result.Success {
				fmt.Printf("Test Case: %s\n", result.TestExecution.TestCase.Name)
				fmt.Printf("Model: %s\n", result.TestExecution.ModelName)
				if result.ErrorMessage != "" {
					fmt.Printf("Error: %s\n", result.ErrorMessage)
				}
				fmt.Printf("Response Time: %v\n", result.Metrics.ResponseTime)
				fmt.Println(strings.Repeat("-", 30))
			}
		}
	}

	// Print overall success rate
	successRate := float64(report.PassedTests) / float64(report.TotalTests) * 100
	fmt.Printf("\n📊 Overall Success Rate: %.2f%%\n", successRate)
}
