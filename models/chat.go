package models

import (
	"time"
)

// ChatRole represents the role of a message in a chat conversation
type ChatRole string

const (
	RoleUser      ChatRole = "user"
	RoleAssistant ChatRole = "assistant"
	RoleSystem    ChatRole = "system"
)

// ChatMessage represents a single message in a chat conversation
type ChatMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// ChatSession represents a conversation session with message history
type ChatSession struct {
	SessionID string        `json:"session_id"`
	Messages  []ChatMessage `json:"messages"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// ChatResponse represents the response from processing a chat message
type ChatResponse struct {
	Message     string           `json:"message"`
	SessionID   string           `json:"session_id"`
	CartSummary *CartSummary     `json:"cart_summary,omitempty"`
	Timestamp   time.Time        `json:"timestamp"`
	ToolCalls   []ToolCallResult `json:"tool_calls,omitempty"`
}

// ToolCallResult represents the result of executing a tool call
type ToolCallResult struct {
	CallID    string      `json:"call_id"`
	ToolName  string      `json:"tool_name"`
	Success   bool        `json:"success"`
	Result    interface{} `json:"result,omitempty"`
	Error     string      `json:"error,omitempty"`
	Arguments string      `json:"arguments"`
}

// CartSummary represents the current state of a shopping cart
type CartSummary struct {
	SessionID string     `json:"session_id"`
	Items     []CartItem `json:"items"`
	Total     float64    `json:"total"`
	ItemCount int        `json:"item_count"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CartItem represents an item in the shopping cart
type CartItem struct {
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	Subtotal    float64 `json:"subtotal"`
}

// ProductFilter represents search parameters for products
type ProductFilter struct {
	Query    string `json:"query,omitempty"`
	Category string `json:"category,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// Product represents a product in the catalog
type Product struct {
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Description string  `json:"description,omitempty"`
	InStock     bool    `json:"in_stock"`
}

// CheckoutResult represents the result of a checkout operation
type CheckoutResult struct {
	Success   bool      `json:"success"`
	OrderID   string    `json:"order_id,omitempty"`
	Total     float64   `json:"total"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// AgentTestResult represents the result of testing the agent loop
type AgentTestResult struct {
	TestCase     TestCase      `json:"test_case"`
	ModelName    string        `json:"model_name"`
	Config       TestConfig    `json:"config"`
	Response     *ChatResponse `json:"response"`
	Success      bool          `json:"success"`
	MatchedPath  string        `json:"matched_path,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
	ResponseTime time.Duration `json:"response_time"`
}

// AgentReport contains the results of an agent test suite
type AgentReport struct {
	Timestamp   time.Time         `json:"timestamp"`
	TestSuite   string            `json:"test_suite"`
	Results     []AgentTestResult `json:"results"`
	TotalTests  int               `json:"total_tests"`
	PassedTests int               `json:"passed_tests"`
	FailedTests int               `json:"failed_tests"`
	AverageTime time.Duration     `json:"average_time"`
}
