package services

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"model-test/models"
	"model-test/tools"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/param"
)

// OpenAIService handles interactions with the OpenAI API using an agent loop
type OpenAIService struct {
	client        openai.Client
	shoppingTools *tools.ShoppingTools
	toolExecutor  *ToolExecutor
	cartService   *CartService
	defaultModel  string
	baseURL       string
	logger        *RequestLogger
}

// NewOpenAIServiceWithLogger creates a new OpenAI service instance with logging
func NewOpenAIServiceWithLogger(apiKey, baseURL, defaultModel string, logger *RequestLogger) *OpenAIService {
	options := []option.RequestOption{
		option.WithBaseURL(baseURL),
		option.WithAPIKey(apiKey),
	}

	// Disable SSL verification for localhost HTTPS connections (Kamiwaza, etc.)
	if strings.HasPrefix(baseURL, "https://localhost") || strings.Contains(baseURL, "https://127.0.0.1") {
		httpClient := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
		options = append(options, option.WithHTTPClient(httpClient))
	}

	client := openai.NewClient(options...)

	// Initialize services
	productService := NewProductService()
	cartService := NewCartService()
	toolExecutor := NewToolExecutor(productService, cartService)

	// Set default model if not provided
	if defaultModel == "" {
		defaultModel = "gpt-4o-mini"
	}

	return &OpenAIService{
		client:        client,
		shoppingTools: tools.NewShoppingTools(),
		toolExecutor:  toolExecutor,
		cartService:   cartService,
		defaultModel:  defaultModel,
		baseURL:       baseURL,
		logger:        logger,
	}
}

// ProcessChatMessage processes a chat message with test case context for logging
func (ai *OpenAIService) ProcessChatMessage(ctx context.Context, userMessage string, session *models.ChatSession, testCase string) (*models.ChatResponse, error) {
	// Generate session ID if not provided
	sessionID := session.SessionID
	if sessionID == "" {
		sessionID = ai.generateSessionID()
	}

	// Define the tools available to the AI
	t := ai.getToolDefinitions()

	// Build messages including conversation history
	messages := ai.buildMessagesFromSession(session, userMessage)

	var cartSummary *models.CartSummary
	var toolResults []models.ToolCallResult
	var responseMessage string

	// Track LLM request metrics
	var llmRequests int
	var totalLLMTime time.Duration

	// Maximum number of tool call iterations
	maxIterations := 5
	currentIteration := 0

	for currentIteration < maxIterations {
		// Track LLM request time
		llmStart := time.Now()

		// Prepare request parameters
		requestParams := openai.ChatCompletionNewParams{
			Model:       ai.defaultModel,
			Messages:    messages,
			Tools:       t,
			Temperature: param.Opt[float64]{Value: 0},
		}

		// Create the chat completion request
		completion, err := ai.client.Chat.Completions.New(ctx, requestParams)

		// Record LLM request metrics
		llmDuration := time.Since(llmStart)
		llmRequests++
		totalLLMTime += llmDuration

		// Log the request/response or error
		if ai.logger != nil {
			if err != nil {
				if logErr := ai.logger.LogError(testCase, currentIteration+1, requestParams, err, ai.baseURL); logErr != nil {
					fmt.Printf("Failed to log error: %v\n", logErr)
				}
			} else {
				if logErr := ai.logger.LogRequest(testCase, currentIteration+1, requestParams, completion, ai.baseURL); logErr != nil {
					fmt.Printf("Failed to log request: %v\n", logErr)
				}
			}
		}

		if err != nil {
			return nil, fmt.Errorf("failed to get AI response: %w", err)
		}

		// Process the response
		choice := completion.Choices[0]
		responseMessage = choice.Message.Content

		// If no tool calls, we're done
		if len(choice.Message.ToolCalls) == 0 {
			break
		}

		// Add the model's function call message to the conversation
		messages = append(messages, choice.Message.ToParam())

		// Execute tool calls
		iterationResults, err := ai.toolExecutor.ExecuteToolCalls(ctx, choice.Message.ToolCalls, sessionID)
		if err != nil {
			// Log error but don't stop the loop
			fmt.Printf("Error executing tool calls: %v\n", err)
		}

		// Add results to our collection
		toolResults = append(toolResults, iterationResults...)

		// Add tool results to the conversation as function call outputs
		for _, result := range iterationResults {
			// Convert the result to JSON string
			resultJSON, err := json.Marshal(result.Result)
			if err != nil {
				fmt.Printf("Error marshaling tool result: %v\n", err)
				continue
			}

			// Add the function call output message
			messages = append(messages, openai.ToolMessage(string(resultJSON), result.CallID))
		}

		currentIteration++
	}

	// If we hit the maximum iterations, add a warning message
	if currentIteration >= maxIterations {
		responseMessage = "I've reached the maximum number of operations I can perform. Let me know if you need anything else!"
	}

	// Get the final cart summary after all tool executions
	cartSummary = ai.cartService.GetCartSummary(sessionID)

	return &models.ChatResponse{
		Message:      responseMessage,
		SessionID:    sessionID,
		CartSummary:  cartSummary,
		Timestamp:    time.Now(),
		ToolCalls:    toolResults,
		LLMRequests:  llmRequests,
		LLMTotalTime: totalLLMTime,
	}, nil
}

// buildMessagesFromSession converts chat session messages to OpenAI format
func (ai *OpenAIService) buildMessagesFromSession(session *models.ChatSession, userMessage string) []openai.ChatCompletionMessageParamUnion {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(ai.getSystemPrompt()),
	}

	// Add previous messages from the session (if any)
	if session != nil {
		for _, msg := range session.Messages {
			switch models.ChatRole(msg.Role) {
			case models.RoleUser:
				messages = append(messages, openai.UserMessage(msg.Content))
			case models.RoleAssistant:
				messages = append(messages, openai.AssistantMessage(msg.Content))
			case models.RoleSystem:
				// Skip system messages as we already have one
				continue
			}
		}
	}

	// Add the current user message
	messages = append(messages, openai.UserMessage(userMessage))

	return messages
}

// getSystemPrompt returns the system prompt for the shopping assistant
func (ai *OpenAIService) getSystemPrompt() string {
	return `You are a helpful shopping assistant. You can help users search for products, manage their shopping cart, and complete purchases.

Available tools:
- search_products: Search for products by query, category, or both
- add_to_cart: Add products to the shopping cart
- remove_from_cart: Remove products from the shopping cart  
- view_cart: View current cart contents and totals
- checkout: Process checkout for the current cart

Always be helpful and provide clear information about products and cart operations.
If the user asks anything else, politely decline and say you are a shopping assistant.
`
}

// getToolDefinitions returns the tool definitions for OpenAI function calling
func (ai *OpenAIService) getToolDefinitions() []openai.ChatCompletionToolParam {
	return ai.shoppingTools.GetToolDefinitions()
}

// InitializeCartForTest initializes the cart with predefined state for testing
func (ai *OpenAIService) InitializeCartForTest(sessionID string, initialState *models.InitialCartState) error {
	return ai.cartService.InitializeCartState(sessionID, initialState)
}

// generateSessionID generates a random session ID
func (ai *OpenAIService) generateSessionID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("session_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("session_%s", hex.EncodeToString(bytes))
}
