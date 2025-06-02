package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"model-test/models"
)

// TestRunner orchestrates the execution of test suites
type TestRunner struct {
	openaiService *OpenAIService
	results       []models.TestResult
	mutex         sync.Mutex
}

// NewTestRunner creates a new test runner instance
func NewTestRunner(apiKey, baseURL string) *TestRunner {
	return &TestRunner{
		openaiService: NewOpenAIService(apiKey, baseURL),
		results:       make([]models.TestResult, 0),
	}
}

// RunTestSuite executes a complete test suite with multiple models and configurations
func (tr *TestRunner) RunTestSuite(ctx context.Context, testCases []models.TestCase, modelNames []string, configs []models.TestConfig) (*models.Report, error) {
	fmt.Printf("Starting test suite with %d test cases, %d models, and %d configurations\n",
		len(testCases), len(modelNames), len(configs))

	var wg sync.WaitGroup
	resultsChan := make(chan models.TestResult, len(testCases)*len(modelNames)*len(configs))

	// Execute tests concurrently
	for _, testCase := range testCases {
		for _, modelName := range modelNames {
			for _, config := range configs {
				wg.Add(1)
				go func(tc models.TestCase, mn string, cfg models.TestConfig) {
					defer wg.Done()

					execution := models.TestExecution{
						TestCase:  tc,
						ModelName: mn,
						Config:    cfg,
					}

					fmt.Printf("Running test: %s with model: %s\n", tc.Name, mn)
					result, err := tr.openaiService.ExecuteTest(ctx, execution)
					if err != nil {
						fmt.Printf("Error running test %s with model %s: %v\n", tc.Name, mn, err)
						// Create a failed result
						result = &models.TestResult{
							TestExecution: execution,
							Success:       false,
							ErrorMessage:  err.Error(),
							Timestamp:     time.Now(),
							Metrics: models.TestMetrics{
								ResponseTime: 0,
							},
						}
					}

					resultsChan <- *result
				}(testCase, modelName, config)
			}
		}
	}

	// Wait for all tests to complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var results []models.TestResult
	var totalTime time.Duration
	passedTests := 0
	failedTests := 0

	for result := range resultsChan {
		results = append(results, result)
		totalTime += result.Metrics.ResponseTime
		if result.Success {
			passedTests++
		} else {
			failedTests++
		}
	}

	// Calculate average time
	var averageTime time.Duration
	if len(results) > 0 {
		averageTime = totalTime / time.Duration(len(results))
	}

	report := &models.Report{
		Timestamp:   time.Now(),
		TestSuite:   "OpenAI Model Tool Efficiency Test",
		Results:     results,
		TotalTests:  len(results),
		PassedTests: passedTests,
		FailedTests: failedTests,
		AverageTime: averageTime,
	}

	return report, nil
}

// SaveResults saves test results to a JSON file
func (tr *TestRunner) SaveResults(filename string, report *models.Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	return os.WriteFile(filename, data, 0644)
}
