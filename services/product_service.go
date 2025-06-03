package services

import (
	"model-test/models"
	"strings"
)

// ProductService handles product search and catalog operations
type ProductService struct {
	products []models.Product
}

// NewProductService creates a new product service with mock data
func NewProductService() *ProductService {
	return &ProductService{
		products: getMockProducts(),
	}
}

// SearchProducts searches for products based on the provided filter
func (ps *ProductService) SearchProducts(filter models.ProductFilter) ([]models.Product, error) {
	var results []models.Product

	// Set default limit if not specified
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}

	for _, product := range ps.products {
		// Filter by category if specified
		if filter.Category != "" && !strings.EqualFold(product.Category, filter.Category) {
			continue
		}

		// Filter by query if specified (search in name and description)
		if filter.Query != "" {
			query := strings.ToLower(filter.Query)
			name := strings.ToLower(product.Name)
			description := strings.ToLower(product.Description)

			if !strings.Contains(name, query) && !strings.Contains(description, query) {
				continue
			}
		}

		results = append(results, product)

		// Stop if we've reached the limit
		if len(results) >= limit {
			break
		}
	}

	return results, nil
}

// getMockProducts returns a list of mock products for testing
func getMockProducts() []models.Product {
	return []models.Product{
		{
			Name:        "iPhone 15",
			Category:    "electronics",
			Price:       999.99,
			Description: "Latest Apple smartphone with advanced features",
			InStock:     true,
		},
		{
			Name:        "Samsung Galaxy S24",
			Category:    "electronics",
			Price:       899.99,
			Description: "Premium Android smartphone with excellent camera",
			InStock:     true,
		},
		{
			Name:        "Wireless Headphones",
			Category:    "electronics",
			Price:       199.99,
			Description: "High-quality wireless headphones with noise cancellation",
			InStock:     true,
		},
		{
			Name:        "MacBook Pro",
			Category:    "electronics",
			Price:       1999.99,
			Description: "Professional laptop for developers and creators",
			InStock:     true,
		},
		{
			Name:        "Running Shoes",
			Category:    "clothing",
			Price:       129.99,
			Description: "Comfortable running shoes for daily exercise",
			InStock:     true,
		},
		{
			Name:        "Winter Jacket",
			Category:    "clothing",
			Price:       89.99,
			Description: "Warm winter jacket for cold weather",
			InStock:     true,
		},
		{
			Name:        "Coffee Maker",
			Category:    "home",
			Price:       79.99,
			Description: "Automatic coffee maker for perfect morning brew",
			InStock:     true,
		},
		{
			Name:        "Vacuum Cleaner",
			Category:    "home",
			Price:       149.99,
			Description: "Powerful vacuum cleaner for home cleaning",
			InStock:     true,
		},
		{
			Name:        "Programming Book",
			Category:    "books",
			Price:       49.99,
			Description: "Learn programming with this comprehensive guide",
			InStock:     true,
		},
		{
			Name:        "Cookbook",
			Category:    "books",
			Price:       29.99,
			Description: "Delicious recipes for home cooking",
			InStock:     true,
		},
		{
			Name:        "Tennis Racket",
			Category:    "sports",
			Price:       159.99,
			Description: "Professional tennis racket for competitive play",
			InStock:     true,
		},
		{
			Name:        "Yoga Mat",
			Category:    "sports",
			Price:       39.99,
			Description: "Non-slip yoga mat for comfortable practice",
			InStock:     true,
		},
		{
			Name:        "Face Cream",
			Category:    "beauty",
			Price:       24.99,
			Description: "Moisturizing face cream for healthy skin",
			InStock:     true,
		},
		{
			Name:        "Shampoo",
			Category:    "beauty",
			Price:       12.99,
			Description: "Gentle shampoo for all hair types",
			InStock:     true,
		},
		{
			Name:        "Board Game",
			Category:    "toys",
			Price:       34.99,
			Description: "Fun board game for family entertainment",
			InStock:     true,
		},
		{
			Name:        "Action Figure",
			Category:    "toys",
			Price:       19.99,
			Description: "Collectible action figure for kids and collectors",
			InStock:     true,
		},
		{
			Name:        "Organic Pasta",
			Category:    "food",
			Price:       4.99,
			Description: "Organic whole wheat pasta for healthy meals",
			InStock:     true,
		},
		{
			Name:        "Green Tea",
			Category:    "food",
			Price:       8.99,
			Description: "Premium green tea with antioxidants",
			InStock:     true,
		},
	}
}
