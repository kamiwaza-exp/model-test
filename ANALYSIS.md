# Batch Analysis Tool

This document describes the batch analysis functionality for evaluating model performance on tool calling tasks.

## Overview

The `analyze-batch` tool provides comprehensive analysis of batch test results, calculating precision, recall, and F1 scores for both tool invocation (binary) and tool selection (specific tool choice) metrics.

## Quick Start

```bash
# Build the analysis tool
make build-analyzer

# Analyze specific batch
make analyze-batch BATCH_DIR=results/batch_test_20240604_112030

# Get JSON output
make analyze-batch-json BATCH_DIR=results/batch_test_20240604_112030
```

## Metrics Explained

### Tool Invocation (Binary)
Evaluates whether the model correctly decides when to call ANY tool:

- **Precision**: `correct_tool_invocations / total_tool_invocations_made`
  - Of all the times the model called a tool, how many times was it correct?
- **Recall**: `correct_tool_invocations / total_tool_invocations_expected`
  - Of all the times a tool should have been called, how many times did the model actually call it?
- **F1**: `2 * (precision * recall) / (precision + recall)`
  - Harmonic mean of precision and recall

### Tool Selection (Which Tool)
Evaluates whether the model calls the RIGHT specific tools:

- **Precision**: `correct_specific_tools / total_specific_tools_called`
  - Of all the tool calls made, how many were the correct tools?
- **Recall**: `correct_specific_tools / total_specific_tools_expected`
  - Of all the correct tool calls possible, how many were actually made?
- **F1**: `2 * (precision * recall) / (precision + recall)`
  - Harmonic mean of precision and recall

## Command Line Usage

### Direct Tool Usage

```bash
# Basic analysis
./analyze-batch results/batch_test_20240604_112030/

# JSON output
./analyze-batch results/batch_test_20240604_112030/ --format json

# Save to file
./analyze-batch results/batch_test_20240604_112030/ -o analysis_report.txt

# JSON to file
./analyze-batch results/batch_test_20240604_112030/ --format json -o analysis.json
```

### Makefile Integration

```bash
# Build the analysis tool
make build-analyzer

# Analyze specific batch
make analyze-batch BATCH_DIR=results/batch_test_20240604_112030

# Analyze with JSON output
make analyze-batch-json BATCH_DIR=results/batch_test_20240604_112030
```

## Sample Output

### Text Format

```
Batch Analysis Report
=====================
Batch Directory: results/batch_test_20240604_112030
Analysis Date: 2024-06-04 12:46:35

Model Performance Summary:
--------------------------
gpt-4:
  Runs: 2, Tests: 34
  Tool Invocation (Binary):
    Precision: 0.850 (17/20)
    Recall: 0.944 (17/18)
    F1: 0.895
  
  Tool Selection:
    Precision: 0.750 (15/20)
    Recall: 0.833 (15/18)
    F1: 0.789

claude-3-sonnet:
  Runs: 2, Tests: 34
  Tool Invocation (Binary):
    Precision: 0.900 (18/20)
    Recall: 0.900 (18/20)
    F1: 0.900
  
  Tool Selection:
    Precision: 0.800 (16/20)
    Recall: 0.800 (16/20)
    F1: 0.800

Overall Rankings (by Tool Selection F1):
-----------------------------------------
1. claude-3-sonnet (F1: 0.800)
2. gpt-4 (F1: 0.789)

Summary:
--------
Analyzed 2 models with 68 total tests across 4 runs.
Best performing model: claude-3-sonnet (Tool Selection F1: 0.800)
```

### JSON Format

```json
{
  "batch_directory": "results/batch_test_20240604_112030",
  "analysis_date": "2024-06-04T12:46:35Z",
  "models": [
    {
      "model_name": "claude-3-sonnet",
      "tool_invocation": {
        "precision": 0.900,
        "recall": 0.900,
        "f1": 0.900,
        "true_positives": 18,
        "false_positives": 2,
        "true_negatives": 12,
        "false_negatives": 2
      },
      "tool_selection": {
        "precision": 0.800,
        "recall": 0.800,
        "f1": 0.800,
        "true_positives": 16,
        "false_positives": 4,
        "true_negatives": 12,
        "false_negatives": 2
      },
      "total_tests": 34,
      "total_runs": 2,
      "result_files": [
        "results/batch_test_20240604_112030/claude-3-sonnet_agent_test_results_claude-3-sonnet_20240604_112035.json",
        "results/batch_test_20240604_112030/claude-3-sonnet_agent_test_results_claude-3-sonnet_20240604_112145.json"
      ]
    }
  ],
  "summary": "Analyzed 2 models with 68 total tests across 4 runs.\nBest performing model: claude-3-sonnet (Tool Selection F1: 0.800)"
}
```

## Model Aggregation

The analysis tool automatically handles multiple test runs for the same model:

1. **File Discovery**: Finds all `*_agent_test_results_*.json` files in the batch directory
2. **Model Grouping**: Groups files by model name (extracted from filename)
3. **Result Aggregation**: Combines all test results across runs for each model
4. **Unified Metrics**: Calculates single precision/recall/F1 scores per model

### Example File Structure

```
results/batch_test_20240604_112030/
├── gpt-4_agent_test_results_gpt-4_20240604_112035.json      # Run 1
├── gpt-4_agent_test_results_gpt-4_20240604_112145.json      # Run 2
├── claude-3-sonnet_agent_test_results_claude-3-sonnet_20240604_112035.json  # Run 1
├── claude-3-sonnet_agent_test_results_claude-3-sonnet_20240604_112145.json  # Run 2
├── test_execution.log
└── summary_report.json
```

All `gpt-4` results are aggregated together, and all `claude-3-sonnet` results are aggregated together.

## Analysis Logic

### Tool Invocation (Binary)

For each test case:
1. **Expected**: Check if ANY tool should be called (from any expected variant)
2. **Actual**: Check if the model called ANY tool
3. **Classification**:
   - **True Positive**: Should call tool AND did call tool
   - **True Negative**: Should NOT call tool AND did NOT call tool
   - **False Positive**: Should NOT call tool BUT did call tool
   - **False Negative**: Should call tool BUT did NOT call tool

### Tool Selection (Specific)

For each test case:
1. **Expected**: Get all expected tool sequences from all variants
2. **Actual**: Get the actual tool sequence called by the model
3. **Matching**: Check if actual sequence matches ANY expected variant exactly
4. **Classification**:
   - **True Positive**: Correct tools called in correct order
   - **False Positive**: Wrong tools called or wrong order
   - **True Negative**: No tools expected and no tools called
   - **False Negative**: Tools expected but not called correctly

### Matching Rules

- **Binary Matching**: Tool names must match exactly (case-sensitive)
- **Order Sensitive**: Tool sequence order must match expected variants
- **Complete Sequences**: Partial matches are considered incorrect
- **Variant Support**: Any expected variant match counts as correct

## Integration Workflow

### 1. Run Batch Tests

```bash
# Test all models
make test-all-models

# Test specific models
make test-models MODELS="gpt-4,claude-3-sonnet"

# Test with specific case
make test-all-models-case TEST_CASE="simple_add_iphone"
```

### 2. Analyze Results

```bash
# Detailed analysis of specific batch
make analyze-batch BATCH_DIR=results/batch_test_20240604_112030
```

### 3. Export and Compare

```bash
# Export to JSON for further analysis
make analyze-batch-json BATCH_DIR=results/batch_test_20240604_112030 > analysis.json

# Compare multiple batches
./analyze-batch results/batch_test_20240604_112030/ > batch1_analysis.txt
./analyze-batch results/batch_test_20240605_143022/ > batch2_analysis.txt
```

## Error Handling

The analysis tool includes comprehensive error handling:

- **Missing Directories**: Clear error if batch directory doesn't exist
- **No Result Files**: Warning if no `*_agent_test_results_*.json` files found
- **Malformed JSON**: Graceful handling of corrupted result files
- **Model Extraction**: Fallback logic for extracting model names from filenames
- **Empty Results**: Handles cases where models have no test results

## Performance Considerations

- **Memory Efficient**: Processes files sequentially, not loading all into memory
- **Fast Analysis**: Optimized for large batch directories with many models
- **Scalable**: Handles hundreds of test results and multiple model runs
- **Concurrent Safe**: Can be run while batch tests are still executing

## Troubleshooting

### No Results Found

```bash
# Check if batch directory exists
ls -la results/batch_test_20240604_112030/

# Look for result files
find results/batch_test_20240604_112030/ -name "*_agent_test_results_*.json"
```

### Model Name Issues

If models aren't being grouped correctly, check filename patterns:
- Expected: `{model}_agent_test_results_{model}_{timestamp}.json`
- The tool extracts model name from the prefix before `_agent_test_results_`

### JSON Parsing Errors

```bash
# Validate JSON files
jq . results/batch_test_20240604_112030/gpt-4_agent_test_results_*.json
```

## Advanced Usage

### Custom Analysis Scripts

The JSON output can be used for custom analysis:

```bash
# Extract F1 scores for all models
./analyze-batch results/batch_test_20240604_112030/ --format json | \
  jq '.models[] | {model: .model_name, f1: .tool_selection.f1}'

# Compare tool invocation vs selection performance
./analyze-batch results/batch_test_20240604_112030/ --format json | \
  jq '.models[] | {model: .model_name, invocation_f1: .tool_invocation.f1, selection_f1: .tool_selection.f1}'
```

### Batch Comparison

```bash
# Compare multiple batches
for batch in results/batch_test_*/; do
  echo "=== $batch ==="
  ./analyze-batch "$batch" | grep "Best performing model"
done
```

## Future Enhancements

Potential improvements for the analysis tool:

1. **Statistical Significance**: Add confidence intervals and significance testing
2. **Visualization**: Generate charts and graphs from analysis data
3. **Detailed Breakdowns**: Per-test-case and per-tool analysis
4. **Trend Analysis**: Compare performance across multiple batches over time
5. **Export Formats**: CSV, Excel, and other formats for data analysis tools
