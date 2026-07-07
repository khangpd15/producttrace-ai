package service

import (
	"context"

	"github.com/google/uuid"
)

// IProductService acts as an anti-corruption layer interface for calling the Product Module
type IProductService interface {
	// FindProductByQR returns the Product ID if found.
	FindProductByQR(ctx context.Context, qrCode string) (uuid.UUID, error)

	// ValidateProductOwnershipStatus ensures that the product's ownership status is acceptable (e.g., Unregistered)
	ValidateProductOwnershipStatus(ctx context.Context, productID uuid.UUID) error

	// UpdateOwnershipStatus updates the status on the Product record
	UpdateOwnershipStatus(ctx context.Context, productID uuid.UUID, status string) error

	// GetProductItemDetail returns basic info to display on the ownership detail page
	GetProductItemDetail(ctx context.Context, productItemID uuid.UUID) (name string, sku string, err error)

	// SearchProductItemIDs dùng để filter search ownership theo sản phẩm (FR-042)
	SearchProductItemIDs(ctx context.Context, productName string, productCode string) ([]uuid.UUID, error)
}
