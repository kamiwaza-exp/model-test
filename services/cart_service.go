package services

import (
	"fmt"
	"model-test/models"
	"sync"
	"time"
)

// CartService handles shopping cart operations for different sessions
type CartService struct {
	carts map[string]*models.CartSummary
	mutex sync.RWMutex
}

// NewCartService creates a new cart service
func NewCartService() *CartService {
	return &CartService{
		carts: make(map[string]*models.CartSummary),
	}
}

// AddToCart adds a product to the cart for the given session
func (cs *CartService) AddToCart(sessionID, productName string, quantity int) (*models.CartSummary, error) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if quantity <= 0 {
		quantity = 1
	}

	// Get or create cart for session
	cart := cs.getOrCreateCart(sessionID)

	// Find existing item or create new one
	found := false
	for i, item := range cart.Items {
		if item.ProductName == productName {
			cart.Items[i].Quantity += quantity
			cart.Items[i].Subtotal = float64(cart.Items[i].Quantity) * cart.Items[i].Price
			found = true
			break
		}
	}

	if !found {
		// Get product price (mock pricing)
		price := cs.getProductPrice(productName)
		newItem := models.CartItem{
			ProductName: productName,
			Quantity:    quantity,
			Price:       price,
			Subtotal:    float64(quantity) * price,
		}
		cart.Items = append(cart.Items, newItem)
	}

	cs.updateCartTotals(cart)
	return cart, nil
}

// RemoveFromCart removes a product from the cart for the given session
func (cs *CartService) RemoveFromCart(sessionID, productName string) (*models.CartSummary, error) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cart := cs.getOrCreateCart(sessionID)

	// Find and remove the item
	for i, item := range cart.Items {
		if item.ProductName == productName {
			cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			break
		}
	}

	cs.updateCartTotals(cart)
	return cart, nil
}

// GetCartSummary returns the current cart summary for the given session
func (cs *CartService) GetCartSummary(sessionID string) *models.CartSummary {
	cs.mutex.RLock()
	defer cs.mutex.RUnlock()

	cart := cs.getOrCreateCart(sessionID)
	return cart
}

// CheckoutCart processes checkout for the given session and clears the cart
func (cs *CartService) CheckoutCart(sessionID string) (*models.CheckoutResult, error) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	cart := cs.getOrCreateCart(sessionID)
	total := cart.Total
	orderID := fmt.Sprintf("ORD-%d", time.Now().Unix())

	// Clear the cart after checkout
	cart.Items = []models.CartItem{}
	cart.Total = 0
	cart.ItemCount = 0
	cart.UpdatedAt = time.Now()

	return &models.CheckoutResult{
		Success:   true,
		OrderID:   orderID,
		Total:     total,
		Message:   "Order processed successfully",
		Timestamp: time.Now(),
	}, nil
}

// getOrCreateCart gets an existing cart or creates a new one for the session
func (cs *CartService) getOrCreateCart(sessionID string) *models.CartSummary {
	cart, exists := cs.carts[sessionID]
	if !exists {
		cart = &models.CartSummary{
			SessionID: sessionID,
			Items:     []models.CartItem{},
			Total:     0,
			ItemCount: 0,
			UpdatedAt: time.Now(),
		}
		cs.carts[sessionID] = cart
	}
	return cart
}

// updateCartTotals recalculates the cart totals
func (cs *CartService) updateCartTotals(cart *models.CartSummary) {
	total := 0.0
	itemCount := 0

	for _, item := range cart.Items {
		total += item.Subtotal
		itemCount += item.Quantity
	}

	cart.Total = total
	cart.ItemCount = itemCount
	cart.UpdatedAt = time.Now()
}

// InitializeCartState sets up the cart with predefined items for testing
func (cs *CartService) InitializeCartState(sessionID string, initialState *models.InitialCartState) error {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()

	if initialState == nil {
		return nil
	}

	// Create a new cart for the session
	cart := &models.CartSummary{
		SessionID: sessionID,
		Items:     []models.CartItem{},
		Total:     0,
		ItemCount: 0,
		UpdatedAt: time.Now(),
	}

	// Add each item from the initial state
	for _, initialItem := range initialState.Items {
		if initialItem.Quantity <= 0 {
			continue
		}

		price := cs.getProductPrice(initialItem.ProductName)
		subtotal := float64(initialItem.Quantity) * price

		cartItem := models.CartItem{
			ProductName: initialItem.ProductName,
			Quantity:    initialItem.Quantity,
			Price:       price,
			Subtotal:    subtotal,
		}
		cart.Items = append(cart.Items, cartItem)
	}

	// Update totals
	cs.updateCartTotals(cart)

	// Store the cart
	cs.carts[sessionID] = cart

	return nil
}

// getProductPrice returns a mock price for a product
func (cs *CartService) getProductPrice(productName string) float64 {
	// Mock pricing based on product name
	priceMap := map[string]float64{
		"iPhone 15":           999.99,
		"Samsung Galaxy S24":  899.99,
		"Wireless Headphones": 199.99,
		"MacBook Pro":         1999.99,
		"Running Shoes":       129.99,
		"Winter Jacket":       89.99,
		"Coffee Maker":        79.99,
		"Vacuum Cleaner":      149.99,
		"Programming Book":    49.99,
		"Cookbook":            29.99,
		"Tennis Racket":       159.99,
		"Yoga Mat":            39.99,
		"Face Cream":          24.99,
		"Shampoo":             12.99,
		"Board Game":          34.99,
		"Action Figure":       19.99,
		"Organic Pasta":       4.99,
		"Green Tea":           8.99,
	}

	if price, exists := priceMap[productName]; exists {
		return price
	}

	// Default price for unknown products
	return 99.99
}
