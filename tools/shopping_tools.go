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

// Mock implementations for testing purposes

// SearchProducts simulates product search
func (st *ShoppingTools) SearchProducts(query, category string, limit int) []map[string]interface{} {
	// Mock product data
	products := []map[string]interface{}{
		{"name": "iPhone 15", "category": "electronics", "price": 999.99},
		{"name": "Samsung Galaxy S24", "category": "electronics", "price": 899.99},
		{"name": "Wireless Headphones", "category": "electronics", "price": 199.99},
		{"name": "Running Shoes", "category": "clothing", "price": 129.99},
		{"name": "Coffee Maker", "category": "home", "price": 79.99},
		{"name": "Programming Book", "category": "books", "price": 49.99},
	}

	// Simple filtering logic for mock
	var results []map[string]interface{}
	for _, product := range products {
		if category != "" && product["category"] != category {
			continue
		}
		if query != "" {
			// Simple string matching
			name := product["name"].(string)
			if !contains(name, query) {
				continue
			}
		}
		results = append(results, product)
		if len(results) >= limit {
			break
		}
	}

	return results
}

// AddToCart adds a product to the cart
func (st *ShoppingTools) AddToCart(productName string, quantity int) bool {
	if quantity <= 0 {
		quantity = 1
	}
	st.cart[productName] += quantity
	return true
}

// RemoveFromCart removes a product from the cart
func (st *ShoppingTools) RemoveFromCart(productName string) bool {
	delete(st.cart, productName)
	return true
}

// ViewCart returns the current cart contents
func (st *ShoppingTools) ViewCart() map[string]interface{} {
	total := 0.0
	items := make([]map[string]interface{}, 0)

	for product, quantity := range st.cart {
		// Mock price calculation
		price := 99.99 // Default price
		subtotal := price * float64(quantity)
		total += subtotal

		items = append(items, map[string]interface{}{
			"product":  product,
			"quantity": quantity,
			"price":    price,
			"subtotal": subtotal,
		})
	}

	return map[string]interface{}{
		"items":      items,
		"total":      total,
		"item_count": len(st.cart),
	}
}

// Checkout processes the checkout
func (st *ShoppingTools) Checkout() map[string]interface{} {
	cart := st.ViewCart()
	st.cart = make(map[string]int) // Clear cart after checkout

	return map[string]interface{}{
		"success":  true,
		"order_id": "ORD-12345",
		"total":    cart["total"],
		"message":  "Order processed successfully",
	}
}

// Helper function for string matching
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
