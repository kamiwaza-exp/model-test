package tools

import (
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/param"
	"github.com/openai/openai-go/shared"
)

// ShoppingTools provides the shopping cart tool definitions and mock implementations
type ShoppingTools struct {
	cart map[string]int // product_name -> quantity
}

// NewShoppingTools creates a new instance of shopping tools
func NewShoppingTools() *ShoppingTools {
	return &ShoppingTools{
		cart: make(map[string]int),
	}
}

// GetToolDefinitions returns the tool definitions for OpenAI function calling
func (st *ShoppingTools) GetToolDefinitions() []openai.ChatCompletionToolParam {
	return []openai.ChatCompletionToolParam{
		{
			Type: "function",
			Function: shared.FunctionDefinitionParam{
				Name:        "search_products",
				Description: param.NewOpt("Search for products by query, category, or price range"),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "Search query for product name or description",
						},
						"category": map[string]interface{}{
							"type":        "string",
							"description": "Product category (electronics, clothing, books, home, sports, beauty, toys, food)",
						},
						"limit": map[string]interface{}{
							"type":        "integer",
							"description": "Maximum number of results to return (default: 10)",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: shared.FunctionDefinitionParam{
				Name:        "add_to_cart",
				Description: param.NewOpt("Add a product to the shopping cart"),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"product_name": map[string]interface{}{
							"type":        "string",
							"description": "The name of the product to add",
						},
						"quantity": map[string]interface{}{
							"type":        "integer",
							"description": "Quantity to add (default: 1)",
						},
					},
					"required": []string{"product_name"},
				},
			},
		},
		{
			Type: "function",
			Function: shared.FunctionDefinitionParam{
				Name:        "remove_from_cart",
				Description: param.NewOpt("Remove a product from the shopping cart"),
				Parameters: shared.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"product_name": map[string]interface{}{
							"type":        "string",
							"description": "The name of the product to remove",
						},
					},
					"required": []string{"product_name"},
				},
			},
		},
		{
			Type: "function",
			Function: shared.FunctionDefinitionParam{
				Name:        "view_cart",
				Description: param.NewOpt("View the current shopping cart contents and totals"),
				Parameters: shared.FunctionParameters{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: shared.FunctionDefinitionParam{
				Name:        "checkout",
				Description: param.NewOpt("Process checkout for the current cart"),
				Parameters: shared.FunctionParameters{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
	}
}
