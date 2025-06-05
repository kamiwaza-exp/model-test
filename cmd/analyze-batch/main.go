package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"model-test/models"
)

// MetricSet represents precision, recall, and F1 metrics
type MetricSet struct {
	Precision      float64 `json:"precision"`
	Recall         float64 `json:"recall"`
	F1             float64 `json:"f1"`
	TruePositives  int     `json:"true_positives"`
	FalsePositives int     `json:"false_positives"`
	TrueNegatives  int     `json:"true_negatives"`
	FalseNegatives int     `json:"false_negatives"`
}

// ModelAnalysis represents the analysis results for a single model
type ModelAnalysis struct {
	ModelName           string    `json:"model_name"`
	ToolInvocation      MetricSet `json:"tool_invocation"`       // Binary: should call tool vs did call tool
	ToolSelection       MetricSet `json:"tool_selection"`        // Specific: right tool vs wrong tool
	AverageResponseTime float64   `json:"average_response_time"` // Average response time in seconds
	TotalTests          int       `json:"total_tests"`
	TotalRuns           int       `json:"total_runs"`
	ResultFiles         []string  `json:"result_files"`
}

// BatchAnalysisReport represents the complete analysis report
type BatchAnalysisReport struct {
	BatchDirectory string          `json:"batch_directory"`
	AnalysisDate   time.Time       `json:"analysis_date"`
	Models         []ModelAnalysis `json:"models"`
	Summary        string          `json:"summary"`
}

func main() {
	var (
		outputFile = flag.String("o", "", "Output file path (default: stdout)")
		format     = flag.String("format", "text", "Output format: text or json")
	)
	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <batch_directory>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	batchDir := flag.Args()[0]

	// Validate batch directory exists
	if _, err := os.Stat(batchDir); os.IsNotExist(err) {
		log.Fatalf("Batch directory does not exist: %s", batchDir)
	}

	// Analyze the batch
	report, err := analyzeBatch(batchDir)
	if err != nil {
		log.Fatalf("Failed to analyze batch: %v", err)
	}

	// Generate output
	var output string
	if *format == "json" {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %v", err)
		}
		output = string(data)
	} else {
		output = generateTextReport(report)
	}

	// Write output
	if *outputFile != "" {
		err := os.WriteFile(*outputFile, []byte(output), 0644)
		if err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
		fmt.Printf("Analysis report written to: %s\n", *outputFile)
	} else {
		fmt.Print(output)
	}
}

// analyzeBatch analyzes all result files in a batch directory
func analyzeBatch(batchDir string) (*BatchAnalysisReport, error) {
	// Find all result files
	resultFiles, err := findResultFiles(batchDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find result files: %w", err)
	}

	if len(resultFiles) == 0 {
		return nil, fmt.Errorf("no result files found in directory: %s", batchDir)
	}

	// Group files by model
	modelFiles := groupFilesByModel(resultFiles)

	// Analyze each model
	var models []ModelAnalysis
	for modelName, files := range modelFiles {
		analysis, err := analyzeModel(modelName, files)
		if err != nil {
			log.Printf("Warning: failed to analyze model %s: %v", modelName, err)
			continue
		}
		models = append(models, *analysis)
	}

	// Sort models by F1 score (tool selection) descending
	sort.Slice(models, func(i, j int) bool {
		return models[i].ToolSelection.F1 > models[j].ToolSelection.F1
	})

	report := &BatchAnalysisReport{
		BatchDirectory: batchDir,
		AnalysisDate:   time.Now(),
		Models:         models,
		Summary:        generateSummary(models),
	}

	return report, nil
}

// findResultFiles finds all agent test result files in the directory
func findResultFiles(dir string) ([]string, error) {
	var files []string
	pattern := regexp.MustCompile(`.*_agent_test_results_.*\.json$`)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && pattern.MatchString(d.Name()) {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// groupFilesByModel groups result files by model name
func groupFilesByModel(files []string) map[string][]string {
	modelFiles := make(map[string][]string)

	// Pattern to extract model name from filename
	// Expected format: {model}_agent_test_results_{model}_{timestamp}.json
	pattern := regexp.MustCompile(`^(.+?)_agent_test_results_`)

	for _, file := range files {
		basename := filepath.Base(file)
		matches := pattern.FindStringSubmatch(basename)

		var modelName string
		if len(matches) > 1 {
			modelName = matches[1]
		} else {
			// Fallback: try to extract model name from the middle part
			parts := strings.Split(basename, "_")
			if len(parts) >= 4 {
				modelName = parts[0]
			} else {
				modelName = "unknown"
			}
		}

		modelFiles[modelName] = append(modelFiles[modelName], file)
	}

	return modelFiles
}

// analyzeModel analyzes all result files for a single model
func analyzeModel(modelName string, files []string) (*ModelAnalysis, error) {
	var allResults []models.AgentTestResult

	// Load and aggregate all results from all files
	for _, file := range files {
		results, err := loadResultFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to load file %s: %w", file, err)
		}
		allResults = append(allResults, results...)
	}

	if len(allResults) == 0 {
		return nil, fmt.Errorf("no test results found for model %s", modelName)
	}

	// Calculate metrics
	toolInvocation := calculateToolInvocationMetrics(allResults)
	toolSelection := calculateToolSelectionMetrics(allResults)
	averageResponseTime := calculateAverageResponseTime(allResults)

	analysis := &ModelAnalysis{
		ModelName:           modelName,
		ToolInvocation:      toolInvocation,
		ToolSelection:       toolSelection,
		AverageResponseTime: averageResponseTime,
		TotalTests:          len(allResults),
		TotalRuns:           len(files),
		ResultFiles:         files,
	}

	return analysis, nil
}

// loadResultFile loads test results from a JSON file
func loadResultFile(filename string) ([]models.AgentTestResult, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var report models.AgentReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, err
	}

	return report.Results, nil
}

// calculateToolInvocationMetrics calculates binary tool invocation metrics
func calculateToolInvocationMetrics(results []models.AgentTestResult) MetricSet {
	var tp, fp, tn, fn int

	for _, result := range results {
		shouldCallTool := shouldCallAnyTool(result.TestCase)

		// Handle nil response - treat as no tools called
		var didCallTool bool
		if result.Response != nil {
			didCallTool = len(result.Response.ToolCalls) > 0
		} else {
			didCallTool = false
		}

		if shouldCallTool && didCallTool {
			tp++ // Should call and did call
		} else if !shouldCallTool && !didCallTool {
			tn++ // Should not call and did not call
		} else if !shouldCallTool && didCallTool {
			fp++ // Should not call but did call
		} else {
			fn++ // Should call but did not call
		}
	}

	return calculateMetrics(tp, fp, tn, fn)
}

// calculateToolSelectionMetrics calculates specific tool selection metrics
func calculateToolSelectionMetrics(results []models.AgentTestResult) MetricSet {
	var tp, fp, tn, fn int

	for _, result := range results {
		expectedTools := getExpectedTools(result.TestCase)
		actualTools := getActualTools(result.Response)

		if len(expectedTools) == 0 && len(actualTools) == 0 {
			tn++ // No tools expected, no tools called
			continue
		}

		if len(expectedTools) == 0 && len(actualTools) > 0 {
			fp++ // No tools expected, but tools called
			continue
		}

		if len(expectedTools) > 0 && len(actualTools) == 0 {
			fn++ // Tools expected, but no tools called
			continue
		}

		// Check if actual tools match any expected variant
		if matchesAnyVariant(result.TestCase, actualTools) {
			tp++ // Correct tools called
		} else {
			fp++ // Wrong tools called
		}
	}

	return calculateMetrics(tp, fp, tn, fn)
}

// shouldCallAnyTool determines if any tool should be called for a test case
func shouldCallAnyTool(testCase models.TestCase) bool {
	for _, variant := range testCase.ExpectedToolVariants {
		if len(variant.Tools) > 0 {
			return true
		}
	}
	return false
}

// getExpectedTools gets all expected tools from all variants
func getExpectedTools(testCase models.TestCase) []string {
	var tools []string
	for _, variant := range testCase.ExpectedToolVariants {
		for _, tool := range variant.Tools {
			tools = append(tools, tool.Name)
		}
	}
	return tools
}

// getActualTools gets all actual tool names called
func getActualTools(response *models.ChatResponse) []string {
	if response == nil {
		return nil
	}

	var tools []string
	for _, toolCall := range response.ToolCalls {
		tools = append(tools, toolCall.ToolName)
	}
	return tools
}

// matchesAnyVariant checks if actual tools match any expected variant
func matchesAnyVariant(testCase models.TestCase, actualTools []string) bool {
	for _, variant := range testCase.ExpectedToolVariants {
		if matchesVariant(variant.Tools, actualTools) {
			return true
		}
	}
	return false
}

// matchesVariant checks if actual tools match a specific variant
func matchesVariant(expectedTools []models.ExpectedToolCall, actualTools []string) bool {
	if len(expectedTools) != len(actualTools) {
		return false
	}

	for i, expected := range expectedTools {
		if i >= len(actualTools) || expected.Name != actualTools[i] {
			return false
		}
	}

	return true
}

// calculateAverageResponseTime calculates the average response time in seconds
func calculateAverageResponseTime(results []models.AgentTestResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	var totalTime time.Duration
	for _, result := range results {
		totalTime += result.ResponseTime
	}

	// Convert to seconds and calculate average
	averageNanoseconds := float64(totalTime) / float64(len(results))
	return averageNanoseconds / 1e9 // Convert nanoseconds to seconds
}

// calculateMetrics calculates precision, recall, and F1 from confusion matrix values
func calculateMetrics(tp, fp, tn, fn int) MetricSet {
	var precision, recall, f1 float64

	if tp+fp > 0 {
		precision = float64(tp) / float64(tp+fp)
	}

	if tp+fn > 0 {
		recall = float64(tp) / float64(tp+fn)
	}

	if precision+recall > 0 {
		f1 = 2 * (precision * recall) / (precision + recall)
	}

	return MetricSet{
		Precision:      precision,
		Recall:         recall,
		F1:             f1,
		TruePositives:  tp,
		FalsePositives: fp,
		TrueNegatives:  tn,
		FalseNegatives: fn,
	}
}

// generateTextReport generates a human-readable text report
func generateTextReport(report *BatchAnalysisReport) string {
	var sb strings.Builder

	sb.WriteString("Batch Analysis Report\n")
	sb.WriteString("=====================\n")
	sb.WriteString(fmt.Sprintf("Batch Directory: %s\n", report.BatchDirectory))
	sb.WriteString(fmt.Sprintf("Analysis Date: %s\n\n", report.AnalysisDate.Format("2006-01-02 15:04:05")))

	sb.WriteString("Model Performance Summary:\n")
	sb.WriteString("--------------------------\n")

	for _, model := range report.Models {
		sb.WriteString(fmt.Sprintf("%s:\n", model.ModelName))
		sb.WriteString(fmt.Sprintf("  Runs: %d, Tests: %d\n", model.TotalRuns, model.TotalTests))
		sb.WriteString(fmt.Sprintf("  Average Response Time: %.2fs\n", model.AverageResponseTime))
		sb.WriteString("  Tool Invocation (Binary):\n")
		sb.WriteString(fmt.Sprintf("    Precision: %.3f (%d/%d)\n",
			model.ToolInvocation.Precision,
			model.ToolInvocation.TruePositives,
			model.ToolInvocation.TruePositives+model.ToolInvocation.FalsePositives))
		sb.WriteString(fmt.Sprintf("    Recall: %.3f (%d/%d)\n",
			model.ToolInvocation.Recall,
			model.ToolInvocation.TruePositives,
			model.ToolInvocation.TruePositives+model.ToolInvocation.FalseNegatives))
		sb.WriteString(fmt.Sprintf("    F1: %.3f\n", model.ToolInvocation.F1))

		sb.WriteString("  Tool Selection:\n")
		sb.WriteString(fmt.Sprintf("    Precision: %.3f (%d/%d)\n",
			model.ToolSelection.Precision,
			model.ToolSelection.TruePositives,
			model.ToolSelection.TruePositives+model.ToolSelection.FalsePositives))
		sb.WriteString(fmt.Sprintf("    Recall: %.3f (%d/%d)\n",
			model.ToolSelection.Recall,
			model.ToolSelection.TruePositives,
			model.ToolSelection.TruePositives+model.ToolSelection.FalseNegatives))
		sb.WriteString(fmt.Sprintf("    F1: %.3f\n\n", model.ToolSelection.F1))
	}

	if len(report.Models) > 1 {
		sb.WriteString("Overall Rankings (by Tool Selection F1):\n")
		sb.WriteString("-----------------------------------------\n")
		for i, model := range report.Models {
			sb.WriteString(fmt.Sprintf("%d. %s (F1: %.3f)\n", i+1, model.ModelName, model.ToolSelection.F1))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(report.Summary)

	return sb.String()
}

// generateSummary generates a summary of the analysis
func generateSummary(models []ModelAnalysis) string {
	if len(models) == 0 {
		return "No models analyzed."
	}

	var sb strings.Builder
	sb.WriteString("Summary:\n")
	sb.WriteString("--------\n")

	if len(models) == 1 {
		model := models[0]
		sb.WriteString(fmt.Sprintf("Analyzed 1 model (%s) with %d tests across %d runs.\n",
			model.ModelName, model.TotalTests, model.TotalRuns))
	} else {
		totalTests := 0
		totalRuns := 0
		for _, model := range models {
			totalTests += model.TotalTests
			totalRuns += model.TotalRuns
		}

		best := models[0]
		sb.WriteString(fmt.Sprintf("Analyzed %d models with %d total tests across %d runs.\n",
			len(models), totalTests, totalRuns))
		sb.WriteString(fmt.Sprintf("Best performing model: %s (Tool Selection F1: %.3f)\n",
			best.ModelName, best.ToolSelection.F1))
	}

	return sb.String()
}
