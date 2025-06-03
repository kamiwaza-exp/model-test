package services

import (
	"context"
	"encoding/json"
	"fmt"
	"model-test/models"

	"github.com/openai/openai-go"
)

// ToolExecutor handles the execution of tool calls
type ToolExecutor struct {
	productService *ProductService
	cartService    *CartService
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(productService *ProductService, cartService *CartService) *ToolExecutor {
	return &ToolExecutor{
		productService: productService,
		cartService:    cartService,
	}
}

// ExecuteToolCalls executes the tool calls from OpenAI
func (te *ToolExecutor) ExecuteToolCalls(ctx context.Context, toolCalls []openai.ChatCompletionMessageToolCall, sessionID string) ([]models.ToolCallResult, error) {
	var results []models.ToolCallResult

	for _, toolCall := range toolCalls {
		result := te.executeToolCall(ctx, toolCall, sessionID)
		results = append(results, result)
	}

	return results, nil
}

// executeToolCall executes a single tool call
func (te *ToolExecutor) executeToolCall(ctx context.Context, toolCall openai.ChatCompletionMessageToolCall, sessionID string) models.ToolCallResult {
	functionName := toolCall.Function.Name
	arguments := toolCall.Function.Arguments
	toolCallID := toolCall.ID

	switch functionName {
	case "search_products":
		return te.handleSearchProducts(arguments, toolCallID)
	case "add_to_cart":
		return te.handleAddToCart(arguments, sessionID, toolCallID)
	case "remove_from_cart":
		return te.handleRemoveFromCart(arguments, sessionID, toolCallID)
	case "view_cart":
		return te.handleViewCart(sessionID, toolCallID)
	case "checkout":
		return te.handleCheckout(sessionID, toolCallID)
	default:
		return models.ToolCallResult{
			CallID:    toolCallID,
			ToolName:  functionName,
			Success:   false,
			Error:     fmt.Sprintf("Unknown tool: %s", functionName),
			Arguments: arguments,
		}
	}
}

// handleSearchProducts handles product search tool calls
func (te *ToolExecutor) handleSearchProducts(arguments string, toolCallID string) models.ToolCallResult {
	var args struct {
		Query    string `json:"query"`
		Category string `json:"category"`
		Limit    int    `json:"limit"`
	}

	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return models.ToolCallResult{
			CallID:    toolCallID,
			ToolName:  "search_products",
			Success:   false,
			Error:     "Invalid arguments",
			Arguments: arguments,
		}
	}

	if args.Limit == 0 {
		args.Limit = 10
	}

	filter := models.ProductFilter{
		Query:    args.Query,
		Category: args.Category,
		Limit:    args.Limit,
	}

	results, err := te.productService.SearchProducts(filter)
	if err != nil {
		return models.ToolCallResult{
			CallID:    toolCallID,
			ToolName:  "search_products",
			Success:   false,
			Error:     err.Error(),
			Arguments: arguments,
		}
	}

	return models.ToolCallResult{
		CallID:    toolCallID,
		ToolName:  "search_products",
		Success:   true,
		Result:    results,
		Arguments: arguments,
	}
}

// handleAddToCart handles add to cart tool calls
func (te *ToolExecutor) handleAddToCart(arguments string, sessionID string, toolCallID string) models.ToolCallResult {
	var args struct {
		ProductName string `json:"product_name"`
		Quantity    int    `json:"quantity"`
	}

	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return models.ToolCallResult{
			CallID:    toolCallID,
			ToolName:  "add_to_cart",
			Success:   false,
			Error:     "Invalid arguments",
			Arguments: arguments,
		}
	}

	if args.Quantity == 0 {
		args.Quantity = 1
	}

	cartSummary, err := te.cartService.AddToCart(sessionID, args.ProductName, args.Quantity)
	if err != nil {
		return models.ToolCallResult{
			CallID:    toolCallID,
			ToolName:  "add_to_cart",
			Success:   false,
			Error:     err.Error(),
			Arguments: arguments,
		}
	}

	return models.ToolCallResult{
		CallID:    toolCallID,
		ToolName:  "add_to_cart",
		Success:   true,
		Result:    cartSummary,
		Arguments: arguments,
	}
}

// handleRemoveFromCart handles remove from cart tool calls
func (te *ToolExecutor) handleRemoveFromCart(arguments string, sessionID string, toolCallID string) models.ToolCallResult {
	var args struct {
		ProductName string `json:"product_name"`
	}

	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return models.ToolCallResult{
			CallID:    toolCallID,
			ToolName:  "remove_from_cart",
			Success:   false,
			Error:     "Invalid arguments",
			Arguments: arguments,
		}
	}

	cartSummary, err := te.cartService.RemoveFromCart(sessionID, args.ProductName)
	if err != nil {
		return models.ToolCallResult{
			CallID:    toolCallID,
			ToolName:  "remove_from_cart",
			Success:   false,
			Error:     err.Error(),
			Arguments: arguments,
		}
	}

	return models.ToolCallResult{
		CallID:    toolCallID,
		ToolName:  "remove_from_cart",
		Success:   true,
		Result:    cartSummary,
		Arguments: arguments,
	}
}

// handleViewCart handles view cart tool calls
func (te *ToolExecutor) handleViewCart(sessionID string, toolCallID string) models.ToolCallResult {
	cartSummary := te.cartService.GetCartSummary(sessionID)
	return models.ToolCallResult{
		CallID:    toolCallID,
		ToolName:  "view_cart",
		Success:   true,
		Result:    cartSummary,
		Arguments: "{}",
	}
}

// handleCheckout handles checkout tool calls
func (te *ToolExecutor) handleCheckout(sessionID string, toolCallID string) models.ToolCallResult {
	checkoutResult, err := te.cartService.CheckoutCart(sessionID)
	if err != nil {
		return models.ToolCallResult{
			CallID:    toolCallID,
			ToolName:  "checkout",
			Success:   false,
			Error:     err.Error(),
			Arguments: "{}",
		}
	}

	return models.ToolCallResult{
		CallID:    toolCallID,
		ToolName:  "checkout",
		Success:   true,
		Result:    checkoutResult,
		Arguments: "{}",
	}
}
